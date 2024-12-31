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
This file contains config package's handlers for reflect types encountered in
configuration parsing and handling
*/
package configurature

import (
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/spf13/pflag"
)

var (
	// pfgFlagMap maps the reflect.Type and the pflag.FlagSet method name to add
	// the flag to a pflag.FlagSet
	pfgFlagMap = make(map[reflect.Type]string)

	// Custom types for this package and the function that can add an instance of
	// the custom type to the FlagSet
	customFlagMap = make(map[reflect.Type]func(string, string, string, string, *pflag.FlagSet))
)

// Value interface for config types
type Value interface {
	Set(string) error // Set the internal value based on the string. If invalid, return error
	String() string   // Internal value as a string
	Type() string     // Type of the value. Appears in Usage() help
}

func init() {

	// Get all types that pflag supports by iterating over a pflag.FlagSet's
	// methods - excluding methods that don't end in "P" (aren't for adding
	// flags with a short name) and methods that contain "Var" (are used for
	// adding Var type flags).
	t := reflect.TypeFor[*pflag.FlagSet]()
	for i := 0; i < t.NumMethod(); i++ {
		name := t.Method(i).Name
		if !strings.HasSuffix(name, "P") || strings.Contains(name, "Var") {
			continue
		}
		// Use the return type of the method as the map key
		pfgFlagMap[t.Method(i).Type.Out(0).Elem()] = name
	}

	// Add Configurature custom types
	AddMapValueType("", map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	})
	AddType[ConfigFile]()

}

// GetSupportedTypes returns all supported struct field types
func getSupportedTypes() []string {
	supported := make([]string, 0, len(pfgFlagMap)+len(customFlagMap))
	for t := range pfgFlagMap {
		supported = append(supported, t.String())
	}
	for t := range customFlagMap {
		supported = append(supported, t.String())
	}
	return supported
}

// AddType adds a custom type to the customFlagMap. You can also use this
// to replace the behavior of existing types.
//
// Parameters:
// - structFieldType: The type of struct field
func AddType[structFieldType any]() {
	rt := reflect.TypeFor[structFieldType]()

	// Create a new var of type *structFieldType and make sure it implements the
	// required Value interface
	ptrType := new(structFieldType)
	if !reflect.TypeOf(ptrType).Implements(reflect.TypeFor[Value]()) {
		panic(fmt.Sprintf("%T must implement Value", ptrType))
	}

	// Add the value to the customFlagMap with a method that will add a flag
	// of that type to the FlagSet
	customFlagMap[rt] = func(name string, short string, def string, desc string, fs *pflag.FlagSet) {
		l := new(structFieldType)
		if def != "" {
			// Use Set() to set the default value of the Value
			r := reflect.ValueOf(l).MethodByName("Set").Call(
				[]reflect.Value{reflect.ValueOf(def)},
			)
			if !r[0].IsNil() {
				panic(fmt.Sprintf("Error setting default value for field %s: %s", name, r[0]))
			}
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

// addToFlagSet adds a flag to the provided FlagSet based on the given type.
//
// Parameters:
// - t: the reflect.Type of the flag
// - fs: the pointer to the pflag.FlagSet to add the flag to
// - name: the name of the flag
// - short: the short name of the flag
// - def: the default value of the flag
// - desc: the description of the flag
func addToFlagSet(t reflect.Type, enumProvided bool, fs *pflag.FlagSet, name string, short string, def string, desc string) {

	isPtr := t.Elem().Kind() == reflect.Ptr
	if isPtr {
		t = t.Elem()
	}

	// Check in the customFlagMap
	if fn, ok := customFlagMap[t.Elem()]; ok {
		// It's a Configurature the function in customFlagMap takes a string
		// for a default value

		// If this is a map value type, add its values to the description
		if !enumProvided {
			if vals := getMapValueTypeValues(t.Elem().String()); vals != nil {
				desc += " (" + strings.Join(*vals, "|") + ")"
			}
		}

		fn(name, short, def, desc, fs)

	} else if method, ok := pfgFlagMap[t.Elem()]; ok {
		// Check for a pflag method in pfgFlagMap

		// Get the method
		m := reflect.ValueOf(fs).MethodByName(method)

		// The pflag.FlagSet method to create a new flag needs a default value
		// of the native type. E.g. int -> 123. The following code is to
		// convert the string value passed to this function to a native type
		// expected by the pflag.FlagSet method.
		defVal := reflect.New(m.Type().Out(0).Elem())
		if def != "" {
			defFs := pflag.NewFlagSet("fake", pflag.PanicOnError)
			reflect.ValueOf(defFs).MethodByName(method).Call([]reflect.Value{
				reflect.ValueOf(name),
				reflect.ValueOf("s"),
				defVal.Elem(),
				reflect.ValueOf(name),
			})
			setFlagValue(name, def, defFs)
			setNativeValue(defVal, name, defFs)
		}

		// Call the flag method on the actual flagset
		m.Call([]reflect.Value{
			reflect.ValueOf(name), reflect.ValueOf(short), defVal.Elem(), reflect.ValueOf(desc)},
		)

	} else {
		panic(fmt.Sprintf("addToFlagSet() unsupported type: %v", t.Elem()))
	}

}

// Set the value to the native type which is returned by the getter on the
// flagset
func setNativeValue(rv reflect.Value, name string, fs *pflag.FlagSet) {
	fv := fs.Lookup(name).Value

	isPtr := rv.Elem().Kind() == reflect.Ptr
	// Init pointer if nil
	if isPtr && rv.Elem().IsNil() {
		rv.Elem().Set(reflect.New(rv.Elem().Type().Elem()))
	}

	// Type of the field value and destination
	pfType := rv.Type().Elem()
	dest := rv.Elem()
	if isPtr {
		pfType = pfType.Elem()
		dest = dest.Elem()
	}

	// For Configurature MapValue types
	if _, ok := mapValueTypeKeys[pfType.String()]; ok {
		vs := reflect.ValueOf(fv).MethodByName("Interface").Call(nil)
		if !vs[0].IsNil() {
			dest.Set(vs[0].Elem())
		}
		return
	}

	// For Custom types
	if _, ok := customFlagMap[pfType]; ok {
		rv := reflect.ValueOf(fv)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		dest.Set(rv)
		return
	}

	// For pflag values

	// If the value type of the target struct field (rv) is in the
	// flagTypeMap, and the method "Get" + map element - "P" exists
	// as a method on the flagset, use that method name instead.
	// This is for complex types such as GetIPSlice. This is brittle,
	// but better than maintaining a static map
	var m reflect.Value
	if method, ok := pfgFlagMap[pfType]; !ok {
		panic("setNativeValue() unsupported type: " + rv.Type().Elem().String())
	} else {
		methodName := "Get" + strings.TrimSuffix(method, "P")
		if m = reflect.ValueOf(fs).MethodByName(methodName); !m.IsValid() {
			panic("setNativeValue()could not find Get method for type: " + rv.Type().Elem().String())
		}
	}

	// Call the method on the flagset
	vs := m.Call([]reflect.Value{
		reflect.ValueOf(name),
	})

	// Panic if error
	if !vs[1].IsNil() {
		panic(vs[1])
	}

	if isPtr {
		// Set pointer value
		rv.Elem().Elem().Set(vs[0])
	} else {
		// Set the value
		rv.Elem().Set(vs[0])
	}

}

// setFlagValue sets the value of a flag in the provided FlagSet.
//
// Parameters:
// - name: the name of the flag
// - value: the new value for the flag
// - fs: the pflag.FlagSet containing the flag
// Return type: error
func setFlagValue(name string, value string, fs *pflag.FlagSet) error {
	flg := fs.Lookup(name)
	if flg == nil {
		return fmt.Errorf("unknown flag: %s", name)
	}
	return flg.Value.Set(value)
}
