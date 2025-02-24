package bsonkit

import (
	"fmt"
	"strconv"

	"github.com/unix-world/smartgoext/db/mongo-driver/bson"
)

// MissingType is the type of the Missing value.
type MissingType struct{}

// Missing represents the absence of a value in a document.
var Missing = MissingType{}

// Get returns the value in the document specified by path. It returns Missing
// if the value has not been found. Dots may be used to descend into nested
// documents e.g. "foo.bar.baz" and numbers may be used to descend into arrays
// e.g. "foo.2.bar".
func Get(doc Doc, path string) interface{} {
	value, _ := get(*doc, path, false, false)
	return value
}

// All has the basic behaviour as Get but additionally collects values from
// embedded documents in arrays. It returns and array and true if values from
// multiple documents, haven been collected. Missing values are skipped and
// intermediary arrays flattened if compact is set to true. By enabling merge,
// a resulting array of embedded document may be merged to on array containing
// all values.
func All(doc Doc, path string, compact, merge bool) (interface{}, bool) {
	// get value
	value, nested := get(*doc, path, true, compact)
	if !nested || !merge {
		return value, nested
	}

	// get array
	array, ok := value.(bson.A)
	if !ok {
		return value, nested
	}

	// prepare result
	result := make(bson.A, 0, len(array))

	// merge arrays
	for _, item := range array {
		if a, ok := item.(bson.A); ok {
			result = append(result, a...)
		} else {
			result = append(result, item)
		}
	}

	return result, nested
}

func get(v interface{}, path string, collect, compact bool) (interface{}, bool) {
	// check path
	if path == PathEnd {
		return v, false
	}

	// check if empty
	if path == "" {
		return Missing, false
	}

	// get key
	key := PathSegment(path)

	// get document field
	if doc, ok := v.(bson.D); ok {
		for _, el := range doc {
			if el.Key == key {
				return get(el.Value, ReducePath(path), collect, compact)
			}
		}
	}

	// get array field
	if arr, ok := v.(bson.A); ok {
		// get indexed array element if number
		if index, ok := ParseIndex(key); ok {
			if index >= 0 && index < len(arr) {
				return get(arr[index], ReducePath(path), collect, compact)
			}
		}

		// collect values from embedded documents
		if collect {
			res := make(bson.A, 0, len(arr))
			for _, item := range arr {
				value, ok := get(item, path, collect, compact)
				if value == Missing && !compact {
					res = append(res, value)
				} else if value != Missing {
					if ok && compact {
						res = append(res, value.(bson.A)...)
					} else {
						res = append(res, value)
					}
				}
			}
			return res, true
		}
	}

	return Missing, false
}

// Put will store the value in the document at the location specified by path
// and return the previously stored value. It will automatically create document
// fields, array elements and embedded documents to fulfill the request. If
// prepends is set to true, new values are inserted at the beginning of the array
// or document. If the path contains a number e.g. "foo.1.bar" and no array
// exists at that levels, a document with the key "1" is created.
func Put(doc Doc, path string, value interface{}, prepend bool) (interface{}, error) {
	// check value
	if value == Missing {
		return nil, fmt.Errorf("cannot put missing value at %s", path)
	}

	// put value
	res, ok := put(*doc, path, value, prepend, func(v interface{}) {
		*doc = v.(bson.D)
	})
	if !ok {
		return nil, fmt.Errorf("cannot put value at %s", path)
	}

	return res, nil
}

// Unset will remove the value at the location in the document specified by path
// and return the previously stored value. If the path specifies an array element
// e.g. "foo.2" the element is nilled, but not removed from the array. This
// prevents unintentional effects through position shifts in the array.
func Unset(doc Doc, path string) interface{} {
	// unset value
	res, _ := put(*doc, path, Missing, false, func(v interface{}) {
		*doc = v.(bson.D)
	})

	return res
}

func put(v interface{}, path string, value interface{}, prepend bool, set func(interface{})) (interface{}, bool) {
	// check path
	if path == PathEnd {
		set(value)
		return v, true
	}

	// check if empty
	if path == "" {
		return Missing, false
	}

	// get key
	key := PathSegment(path)

	// put document field
	if doc, ok := v.(bson.D); ok {
		for i, el := range doc {
			if el.Key == key {
				return put(doc[i].Value, ReducePath(path), value, prepend, func(v interface{}) {
					if v == Missing {
						set(append(doc[:i], doc[i+1:]...))
					} else {
						doc[i].Value = v
					}
				})
			}
		}

		// check if unset
		if value == Missing {
			return Missing, false
		}

		// capture value
		e := bson.E{Key: key}
		res, ok := put(Missing, ReducePath(path), value, prepend, func(v interface{}) {
			e.Value = v
		})
		if !ok {
			return res, false
		}

		// set appended/prepended document
		if prepend {
			set(append(bson.D{e}, doc...))
		} else {
			set(append(doc, e))
		}

		return Missing, true
	}

	// put array field
	if arr, ok := v.(bson.A); ok {
		index, err := strconv.Atoi(key)
		if err != nil || index < 0 {
			return Missing, false
		}

		// update existing element
		if index < len(arr) {
			return put(arr[index], ReducePath(path), value, prepend, func(v interface{}) {
				if v == Missing {
					arr[index] = nil
				} else {
					arr[index] = v
				}
			})
		}

		// check if unset
		if value == Missing {
			return Missing, false
		}

		// fill with nil elements
		for i := len(arr); i < index+1; i++ {
			arr = append(arr, nil)
		}

		// put in last element
		res, ok := put(Missing, ReducePath(path), value, prepend, func(v interface{}) {
			arr[index] = v
		})
		if !ok {
			return res, false
		}

		// set array
		set(arr)

		return Missing, true
	}

	// check if unset
	if value == Missing {
		return Missing, false
	}

	// put new document
	if v == Missing {
		// capture value
		e := bson.E{Key: key}
		res, ok := put(Missing, ReducePath(path), value, prepend, func(v interface{}) {
			e.Value = v
		})
		if !ok {
			return res, false
		}

		// set document
		set(bson.D{e})

		return Missing, true
	}

	return Missing, false
}

// Increment will add the increment to the value at the location in the document
// specified by path and return the new value. If the value is missing, the
// increment is added to the document. The type of the field may be changed as
// part of the operation.
func Increment(doc Doc, path string, increment interface{}) (interface{}, error) {
	// get field
	field := Get(doc, path)

	// ensure zero
	if field == Missing {
		field = int32(0)
	}

	// increment field
	field = Add(field, increment)
	if field == Missing {
		return nil, fmt.Errorf("incrementee or increment is not a number")
	}

	// update field
	_, err := Put(doc, path, field, false)
	if err != nil {
		return nil, err
	}

	return field, nil
}

// Multiply will multiply the multiplier with the value at the location in the
// document specified by path and return the new value. If the value is missing,
// a zero is added to the document. The type of the field may be changed as part
// of the operation.
func Multiply(doc Doc, path string, multiplier interface{}) (interface{}, error) {
	// get field
	field := Get(doc, path)

	// ensure zero
	if field == Missing {
		field = int32(0)
	}

	// multiply
	field = Mul(field, multiplier)
	if field == Missing {
		return nil, fmt.Errorf("multiplicand or multiplier is not a number")
	}

	// update field
	_, err := Put(doc, path, field, false)
	if err != nil {
		return nil, err
	}

	return field, nil
}

// Push will add the value to the array at the location in the document
// specified by path and return the new value. If the value is missing, the
// value is added to a new array.
func Push(doc Doc, path string, value interface{}) (interface{}, error) {
	// check value
	if value == Missing {
		return nil, fmt.Errorf("cannot push missing value at %s", path)
	}

	// get field
	field := Get(doc, path)

	// push field
	switch val := field.(type) {
	case bson.A:
		field = append(val, value)
	case MissingType:
		field = bson.A{value}
	default:
		return nil, fmt.Errorf("value at path %q is not an array", path)
	}

	// update field
	_, err := Put(doc, path, field, false)
	if err != nil {
		return nil, err
	}

	return field, nil
}

// Pop will remove the first or last element from the array at the location in
// the document specified byt path and return the updated array. If the array is
// empty, the value is missing or not an array, it will do nothing and return
// Missing.
func Pop(doc Doc, path string, last bool) (interface{}, error) {
	// get field
	field := Get(doc, path)

	// check if missing
	if field == Missing {
		return Missing, nil
	}

	// get and check array
	array, ok := field.(bson.A)
	if !ok {
		return nil, fmt.Errorf("value at path %q is not an array", path)
	}

	// check length
	if len(array) == 0 {
		return Missing, nil
	}

	// pop last or first value
	var res interface{}
	if last {
		res = array[len(array)-1]
		field = array[:len(array)-1]
	} else {
		res = array[0]
		field = array[1:]
	}

	// update field
	_, err := Put(doc, path, field, false)
	if err != nil {
		return nil, err
	}

	return res, nil
}
