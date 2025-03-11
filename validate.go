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
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/spf13/pflag"
)

// validate configuration
func (c *configurer) validate(s any, fs *pflag.FlagSet) {

	errors := []string{}
	c.visitFields(s, func(f reflect.StructField, tags *reflect.StructTag, v reflect.Value, ancestors []string) (stop bool) {

		fName := fieldNameToConfigName(f.Name, tags, ancestors)

		// Check enums
		if val := tags.Get("enum"); val != "" {
			enums := strings.Split(val, ",")
			v := fs.Lookup(fName).Value.String()
			if !slices.Contains(enums, v) {
				errors = append(errors, fmt.Sprintf("%s must be one of %s", fName, strings.Join(enums, ", ")))
			}
			// This essentially validates required as well. No need to also check for required.
			return false // false == don't stop looping over fields
		}

		// Check that required values are specified
		_, required := tags.Lookup("required")
		if !required && c.opts.RequireNoDefaults {
			_, ok := tags.Lookup("default")
			required = !ok
		}

		if required && !fs.Lookup(fName).Changed {
			errors = append(errors, fmt.Sprintf("%s is required", fName))
		}

		return false // false == don't stop looping over fields
	}, []string{})

	if len(errors) > 0 {
		panic(strings.Join(errors, ", "))
	}
}
