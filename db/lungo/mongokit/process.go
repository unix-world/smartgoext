package mongokit

import (
	"fmt"

	"github.com/unix-world/smartgoext/db/mongo-driver/bson"

	"github.com/unix-world/smartgoext/db/lungo/bsonkit"
)

// TODO: Add support for positional operator `$` (apply and project).

// Operator is a generic operator.
type Operator func(ctx Context, doc bsonkit.Doc, op, path string, v interface{}) error

// Context is the context passed to operators.
type Context struct {
	// The available top level operators.
	TopLevel map[string]Operator

	// The available expression operators.
	Expression map[string]Operator

	// Whether missing operators should just be skipped.
	SkipMissing bool

	// If enabled, top level operators will expect a document with multiple
	// invocations of the operator.
	MultiTopLevel bool

	// A custom value available to the operators.
	Value interface{}

	// The query used to resolve positional operators in top level operator
	// invocation paths.
	TopLevelQuery bsonkit.Doc

	// The array filters used to resolve positional operators in top level
	// operator invocation paths.
	TopLevelArrayFilters bsonkit.List
}

// Process will process a document with a query using the MongoDB operator
// processing algorithm.
func Process(ctx Context, doc bsonkit.Doc, query bson.D, prefix string, root bool) error {
	// process all expressions (implicit and)
	for _, exp := range query {
		err := ProcessExpression(ctx, doc, prefix, exp, root)
		if err != nil {
			return err
		}
	}

	return nil
}

// ProcessExpression will process a document with a query using the MongoDB
// operator algorithm.
func ProcessExpression(ctx Context, doc bsonkit.Doc, prefix string, pair bson.E, root bool) error {
	// handle operator
	if len(pair.Key) > 0 && pair.Key[0] == '$' {
		// lookup operator
		var operator Operator
		if root {
			operator = ctx.TopLevel[pair.Key]
			if operator == nil && ctx.SkipMissing {
				return nil
			} else if operator == nil {
				return fmt.Errorf("unknown top level operator %q", pair.Key)
			}
		} else {
			operator = ctx.Expression[pair.Key]
			if operator == nil && ctx.SkipMissing {
				return nil
			} else if operator == nil {
				return fmt.Errorf("unknown expression operator %q", pair.Key)
			}
		}

		// call operator if not a multi top level
		if !(root && ctx.MultiTopLevel) {
			return operator(ctx, doc, pair.Key, prefix, pair.Value)
		}

		// otherwise, get document
		update, ok := pair.Value.(bson.D)
		if !ok {
			return fmt.Errorf("%s: expected document", pair.Key)
		}

		// call operator for each pair
		for _, cond := range update {
			err := Resolve(cond.Key, ctx.TopLevelQuery, doc, ctx.TopLevelArrayFilters, func(path string) error {
				return operator(ctx, doc, pair.Key, path, cond.Value)
			})
			if err != nil {
				return err
			}
		}

		return nil
	}

	// get path
	path := pair.Key
	if prefix != "" {
		path = prefix + "." + path
	}

	// check for field expressions with a document which may contain either
	// only expression operators or only simple conditions
	if exps, ok := pair.Value.(bson.D); ok {
		// process all expressions (implicit and)
		for i, exp := range exps {
			// stop and leave document as a simple condition if the
			// first key does not look like an operator
			if i == 0 && (len(exp.Key) == 0 || exp.Key[0] != '$') {
				break
			}

			// check operator validity
			if len(exp.Key) == 0 || exp.Key[0] != '$' {
				return fmt.Errorf("expected operator, got %q", exp.Key)
			}

			// lookup operator
			operator := ctx.Expression[exp.Key]
			if operator == nil && ctx.SkipMissing {
				return nil
			} else if operator == nil {
				return fmt.Errorf("unknown expression operator %q", exp.Key)
			}

			// call operator
			err := operator(ctx, doc, exp.Key, path, exp.Value)
			if err != nil {
				return err
			}

			// return success if last one
			if i == len(exps)-1 {
				return nil
			}
		}
	}

	// handle pair as a simple condition

	// get the default operator
	operator := ctx.Expression[""]
	if operator == nil && ctx.SkipMissing {
		return nil
	} else if operator == nil {
		return fmt.Errorf("missing default operator")
	}

	// call operator
	err := operator(ctx, doc, "", path, pair.Value)
	if err != nil {
		return err
	}

	return nil
}
