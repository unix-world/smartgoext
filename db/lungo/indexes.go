package lungo

import (
	"context"
	"time"

	"github.com/unix-world/smartgoext/db/mongo-driver/bson"
	"github.com/unix-world/smartgoext/db/mongo-driver/mongo"
	"github.com/unix-world/smartgoext/db/mongo-driver/mongo/options"

	"github.com/unix-world/smartgoext/db/lungo/bsonkit"
	"github.com/unix-world/smartgoext/db/lungo/mongokit"
)

var _ IIndexView = &IndexView{}

// IndexView wraps an Engine to be mongo compatible.
type IndexView struct {
	engine *Engine
	handle Handle
}

// CreateMany implements the IIndexView.CreateMany method.
func (v *IndexView) CreateMany(ctx context.Context, indexes []mongo.IndexModel, opts ...*options.CreateIndexesOptions) ([]string, error) {
	// merge options
	opt := options.MergeCreateIndexesOptions(opts...)

	// assert supported options
	assertOptions(opt, map[string]string{
		"MaxTime": ignored,
	})

	// check filer
	if len(indexes) == 0 {
		panic("lungo: missing indexes")
	}

	// created indexes separately
	var names []string
	for _, index := range indexes {
		name, err := v.CreateOne(ctx, index, opts...)
		if err != nil {
			return names, err
		}
		names = append(names, name)
	}

	return names, nil
}

// CreateOne implements the IIndexView.CreateOne method.
func (v *IndexView) CreateOne(ctx context.Context, index mongo.IndexModel, opts ...*options.CreateIndexesOptions) (string, error) {
	// merge options
	opt := options.MergeCreateIndexesOptions(opts...)

	// assert supported options
	assertOptions(opt, map[string]string{
		"MaxTime": ignored,
	})

	// assert supported index options
	if index.Options != nil {
		assertOptions(index.Options, map[string]string{
			"Background":              ignored,
			"ExpireAfterSeconds":      supported,
			"Name":                    supported,
			"Unique":                  supported,
			"Version":                 ignored,
			"PartialFilterExpression": supported,
		})
	}

	// transform key
	key, err := bsonkit.Transform(index.Keys)
	if err != nil {
		return "", err
	}

	// get expiry
	var expiry time.Duration
	if index.Options != nil && index.Options.ExpireAfterSeconds != nil {
		if *index.Options.ExpireAfterSeconds == 0 {
			expiry = time.Nanosecond
		} else {
			expiry = time.Duration(*index.Options.ExpireAfterSeconds) * time.Second
		}
	}

	// get name
	var name string
	if index.Options != nil && index.Options.Name != nil {
		name = *index.Options.Name
	}

	// get unique
	var unique bool
	if index.Options != nil && index.Options.Unique != nil {
		unique = *index.Options.Unique
	}

	// get partial
	var partial bsonkit.Doc
	if index.Options != nil && index.Options.PartialFilterExpression != nil {
		partial, err = bsonkit.Transform(index.Options.PartialFilterExpression)
		if err != nil {
			return "", err
		}
	}

	// begin transaction
	txn, err := v.engine.Begin(ctx, true)
	if err != nil {
		return "", err
	}

	// ensure abortion
	defer v.engine.Abort(txn)

	// create index
	name, err = txn.CreateIndex(v.handle, name, mongokit.IndexConfig{
		Key:     key,
		Unique:  unique,
		Partial: partial,
		Expiry:  expiry,
	})
	if err != nil {
		return "", err
	}

	// commit transaction
	err = v.engine.Commit(txn)
	if err != nil {
		return "", err
	}

	return name, nil
}

// DropAll implements the IIndexView.DropAll method.
func (v *IndexView) DropAll(ctx context.Context, opts ...*options.DropIndexesOptions) (bson.Raw, error) {
	// merge options
	opt := options.MergeDropIndexesOptions(opts...)

	// assert supported options
	assertOptions(opt, map[string]string{
		"MaxTime": ignored,
	})

	// begin transaction
	txn, err := v.engine.Begin(ctx, true)
	if err != nil {
		return nil, err
	}

	// ensure abortion
	defer v.engine.Abort(txn)

	// drop all indexes
	err = txn.DropIndex(v.handle, "")
	if err != nil {
		return nil, err
	}

	// commit transaction
	err = v.engine.Commit(txn)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// DropOne implements the IIndexView.DropOne method.
func (v *IndexView) DropOne(ctx context.Context, name string, opts ...*options.DropIndexesOptions) (bson.Raw, error) {
	// merge options
	opt := options.MergeDropIndexesOptions(opts...)

	// assert supported options
	assertOptions(opt, map[string]string{
		"MaxTime": ignored,
	})

	// check name
	if name == "" || name == "*" {
		panic("lungo: invalid index name")
	}

	// begin transaction
	txn, err := v.engine.Begin(ctx, true)
	if err != nil {
		return nil, err
	}

	// ensure abortion
	defer v.engine.Abort(txn)

	// drop all indexes
	err = txn.DropIndex(v.handle, name)
	if err != nil {
		return nil, err
	}

	// commit transaction
	err = v.engine.Commit(txn)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// List implements the IIndexView.List method.
func (v *IndexView) List(ctx context.Context, opts ...*options.ListIndexesOptions) (ICursor, error) {
	// merge options
	opt := options.MergeListIndexesOptions(opts...)

	// assert supported options
	assertOptions(opt, map[string]string{
		"BatchSize": ignored,
		"MaxTime":   ignored,
	})

	// begin transaction
	txn, err := v.engine.Begin(ctx, false)
	if err != nil {
		return nil, err
	}

	// list indexes
	list, err := txn.ListIndexes(v.handle)
	if err != nil {
		return nil, err
	}

	return &Cursor{list: list}, nil
}

// ListSpecifications implements the IIndexView.ListSpecifications method.
func (v *IndexView) ListSpecifications(context.Context, ...*options.ListIndexesOptions) ([]*mongo.IndexSpecification, error) {
	panic("lungo: not implemented")
}
