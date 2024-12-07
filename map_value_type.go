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
This file contains the AddMapValueType[T] factory function and its helpers
*/
package configurature

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/pflag"
)

var (
	// Mapping used by getMapValueTypeValues()
	mapValueTypeKeys = make(map[string][]string)
)

// AddMapValueType creates a map value type based on the input mapping.
//
// The function takes a mapping of string keys to values and registers it as
// a Configurature type.
func AddMapValueType[T any](typeName string, mapping map[string]T) {
	nm := make(map[string]T)
	keys := []string{}
	for k, v := range mapping {
		// Convert to lower case
		nm[strings.ToLower(k)] = v
		// Store the keys for use in getMapValueTypeValues()
		keys = append(keys, k)
	}

	mapValueTypeKeys[reflect.TypeFor[T]().String()] = keys

	customFlagMap[reflect.TypeFor[T]()] = func(name string, short string, def string, desc string, fs *pflag.FlagSet) {
		l := &mapValueType[T]{
			mapping:  nm,
			typeName: typeName,
		}
		if def != "" {
			// Use Set() to set the default value of the Value
			reflect.ValueOf(l).MethodByName("Set").Call(
				[]reflect.Value{reflect.ValueOf(def)},
			)
		}
		// Add the Value to the flagset using VarP
		reflect.ValueOf(fs).MethodByName("VarP").Call(
			[]reflect.Value{
				reflect.ValueOf(l),
				reflect.ValueOf(name),
				reflect.ValueOf(short),
				reflect.ValueOf(desc),
			},
		)
	}
}

// getMapValueTypeValues returns a pointer to the values in the mapping for a
// mapValueType or nil if it does not exist
func getMapValueTypeValues(reflectType string) *[]string {
	if values, ok := mapValueTypeKeys[reflectType]; !ok {
		return nil
	} else {
		return &values
	}
}

// mapValueType is a Configurature type that maps strings to values and
// implements the Value interface.
type mapValueType[T any] struct {
	value    string
	typeName string
	mapping  map[string]T
}

func (m *mapValueType[T]) String() string {
	return m.value
}

func (m *mapValueType[T]) Set(v string) error {
	val := strings.ToLower(v)
	if _, ok := m.mapping[val]; ok {
		m.value = v
		return nil
	} else {
		return fmt.Errorf("invalid %s: \"%s\"", m.Type(), v)
	}
}

func (m *mapValueType[T]) Type() string {
	return m.getType()
}

func (m *mapValueType[T]) Interface() interface{} {
	if m.value == "" {
		return nil
	}
	if v, ok := m.mapping[strings.ToLower(m.value)]; !ok {
		panic(fmt.Sprintf("Invalid value for %s: %s", m.Type(), m.value))
	} else {
		return v
	}
}

// getType gets and cache the type name. Note: type() is an invalid method
// name in Go
func (m *mapValueType[T]) getType() string {
	if m.typeName == "" {
		// Get name from type of T
		vtypes := strings.Split(fmt.Sprintf("%T", new(T)), ".")
		m.typeName = vtypes[len(vtypes)-1]
		m.typeName = strings.TrimPrefix(m.typeName, "*")
	}
	return m.typeName
}
