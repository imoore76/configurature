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

package configurature

import (
	"errors"
	"reflect"
)

var (
	// lastConfigLoaded is the last loaded configuration
	lastConfigLoaded interface{}

	// ErrConfigNotLoaded is returned when the last loaded configuration is nil
	ErrConfigNotLoaded = errors.New("configuration not loaded - did you run Configure[]()?")

	// cache for getting config types
	getConfigTypeCache = make(map[reflect.Type]interface{})

	// For disabling type caching
	DisableGetTypeCache = false
)

// Get returns a pointer to the configuration of type T found anywhere in the
// last loaded configuration.
// Returns (nil, ErrConfigNotLoaded) if the last loaded configuration is nil.
// Returns (nil, nil) if no configuration of type T is found
func Get[T any]() (*T, error) {
	if lastConfigLoaded == nil {
		return nil, ErrConfigNotLoaded
	}
	switch t := lastConfigLoaded.(type) {
	case *T:
		return t, nil
	}

	var t interface{}
	if !DisableGetTypeCache {
		typeKey := reflect.TypeFor[T]()
		var ok bool
		t, ok = getConfigTypeCache[typeKey]
		if !ok || t == nil {
			t = findStructOfType[T](lastConfigLoaded)
			getConfigTypeCache[typeKey] = t
		}
	} else {
		t = findStructOfType[T](lastConfigLoaded)
	}
	return t.(*T), nil
}

// findStructOfType recursively searches for a struct of type T in struct s
func findStructOfType[T any](s interface{}) *T {
	v := reflect.ValueOf(s).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {

		if !t.Field(i).IsExported() {
			continue
		}

		// Handle struct and anonymous struct fields
		if t.Field(i).Type.Kind() == reflect.Struct || t.Field(i).Anonymous {
			fld := v.Field(i).Addr().Interface()
			switch t := fld.(type) {
			case *T:
				return t
			}
			if found := findStructOfType[T](fld); found != nil {
				return found
			}
		}

	}
	return nil
}

// setLastConfig sets the last loaded configuration
func setLastConfig(config interface{}) {
	// Set last config
	lastConfigLoaded = config

	// Clear getConfigTypeCache each time a new config is loaded
	getConfigTypeCache = make(map[reflect.Type]interface{})
}
