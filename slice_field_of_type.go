// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
This file contains the struct definition and methods for sliceFieldOfType which is a wrapper
around a slice of custom field types.
*/

package configurature

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"reflect"
	"strings"
)

// addSliceType adds a custom slice type
func addSliceType[T any]() {

	// Create a new var of type *structFieldType and make sure it implements the
	// required Value interface
	ptrType := new(sliceFieldOfType[T])
	if !reflect.TypeOf(ptrType).Implements(reflect.TypeFor[Value]()) {
		panic(fmt.Sprintf("%T must implement Value", ptrType))
	}

	addToCustomFlagMap[sliceFieldOfType[T], T]()
}

// sliceFieldOfType is a wrapper around a slice of custom field types. It implements the Value
// interface. It is meant to be instantiated ass “sliceFieldOfType[[]customFieldType]“. This
// is done automatically when you call `AddType[[]CustomFieldType]`.
type sliceFieldOfType[T any] struct {
	typeName string
	values   T
}

// Return a string representation of the slice (csv format)
func (f *sliceFieldOfType[T]) String() string {

	vals := reflect.ValueOf(f.values)

	// Return empty string for no values
	if vals.IsNil() {
		return ""
	}

	buf := bytes.NewBuffer(nil)
	w := csv.NewWriter(buf)
	out := make([]string, vals.Len())
	for idx := range vals.Len() {
		out[idx] = fmt.Sprintf("%s", vals.Index(idx))
	}
	w.Write(out)
	w.Flush()
	return strings.TrimRight(buf.String(), "\n")
}

// Return the name of this type
func (f *sliceFieldOfType[T]) Type() string {
	if f.typeName == "" {
		tp := reflect.New(reflect.TypeFor[T]().Elem())
		f.typeName = fmt.Sprintf("[]%s", tp.MethodByName("Type").Call(nil)[0].String())
	}
	return f.typeName
}

// Set the slice values from a csv string
func (f *sliceFieldOfType[T]) Set(v string) error {
	stringReader := strings.NewReader(v)
	csvReader := csv.NewReader(stringReader)
	vals, err := csvReader.Read()

	if err != nil {
		return err
	}

	if reflect.TypeFor[T]().Kind() != reflect.Slice {
		panic("T must be a slice")
	}

	// Initialize the values slice
	newSlice := reflect.MakeSlice(reflect.TypeFor[T](), len(vals), len(vals)) //
	reflect.ValueOf(&(f.values)).Elem().Set(newSlice)

	for idx, v := range vals {

		// Ref to the slice element
		fv := reflect.ValueOf(f.values).Index(idx)

		// Create a new type of the slice element and call Set() on it
		nv := reflect.New(fv.Type())
		r := nv.MethodByName("Set").Call(
			[]reflect.Value{
				reflect.ValueOf(v),
			},
		)
		if !r[0].IsNil() {
			return r[0].Interface().(error)
		}

		// Set the slice element to the new value
		fv.Set(nv.Elem())

	}
	return nil
}

// Return the slice values
func (f *sliceFieldOfType[T]) Interface() any {
	return f.values
}
