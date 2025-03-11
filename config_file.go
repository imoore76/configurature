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
This file contains the ConfigFile type helpers
*/
package configurature

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	fp "path/filepath"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

// setConfigFile checks for a field of type File in the config struct and sets
// the configFile.Value pointer to its address
func (c *configurer) setConfigFile() {
	c.visitFields(c.config, func(f reflect.StructField, _ *reflect.StructTag, v reflect.Value, _ []string) (stop bool) {
		if v.Elem().Type() == configFileType {
			if c.configFile.Value != nil {
				panic("ConfigFile already set to " + *c.configFile.Value)
			}
			c.configFile.Value = (*string)(v.Interface().(*ConfigFile))
			stop = true
		}
		return stop
	}, []string{})
}

// Load configuration from config file
func (c *configurer) loadConfigFile(fs *pflag.FlagSet) {
	// Set from env since setFromEnv() has not been called yet
	// (chicken and egg)
	if envVal := os.Getenv(
		fmt.Sprintf("%s%s", c.opts.EnvPrefix, strcase.ToScreamingSnake(c.configFile.Flag)),
	); envVal != "" {
		*c.configFile.Value = envVal
	}

	// Set up a flagset that only contains the flags we are looking for to
	// get the config file. Parse args to get the value.
	f := pflag.NewFlagSet("cf", pflag.ContinueOnError)
	f.Usage = func() {}
	fileName := new(string)
	f.StringVarP(fileName, c.configFile.Flag, c.configFile.Short, *c.configFile.Value, "")
	f.Parse(c.opts.Args)

	// No config file specified, nothing to do
	if *fileName == "" {
		return
	}

	confFile, err := os.ReadFile(*fileName)
	if err != nil {
		panic(fmt.Sprintf("error reading config file %s: %v ", *fileName, err))
	}

	// Parse config file based on extension
	gMap := make(map[string]any)
	switch fp.Ext(strings.ToLower(*fileName)) {
	case ".json":
		err = json.Unmarshal(confFile, &gMap)
		if err != nil {
			panic(fmt.Sprintf("error parsing config file: %v", err))
		}
	case ".yml", ".yaml":
		err = yaml.Unmarshal(confFile, gMap)
		if err != nil {
			panic(fmt.Sprintf("error parsing config file: %v", err))
		}
	default:
		panic(fmt.Sprintf("unsupported config file type: %s. Supported "+
			"file types are .json, .yml, .yaml", fp.Base(*fileName)))
	}

	// Set config struct fields based on config values from file stored in
	// the generic map
	setFlagsFromGenericMap(&gMap, []string{}, fs)

}

// setFlagsFromGenericMap sets flag values from a generic map recursively. This
// is called after reading the config file.
//
// Parameters:
// - gMap: a pointer to a map[string]any
// - path: a slice of strings representing the path
// - fs: a pointer to a pflag.FlagSet
func setFlagsFromGenericMap(gMap *map[string]any, ancestors []string, fs *pflag.FlagSet) {
	for k, v := range *gMap {

		// Yaml unmarshals into a map[any]any for
		// sub-objects. Convert them to a map[string]any
		if ifaceIfaceMap, ok := v.(map[any]any); ok {
			newV := make(map[string]any)
			for kk, vv := range ifaceIfaceMap {
				newV[fmt.Sprintf("%v", kk)] = vv
			}
			v = newV
		}

		// If it is a map object, it is either an actual map or nested
		// configuration
		if nested, ok := v.(map[string]any); ok {
			// Name was found in FlagSet. It's an actual map
			mapk := strings.Join(append(ancestors, k), "_")
			if flg := fs.Lookup(mapk); flg != nil {
				vstr := []string{}
				for kk, vv := range nested {
					vstr = append(vstr, fmt.Sprintf("%s=%v", kk, vv))
				}
				v = strings.Join(vstr, ",")
			} else {
				// It's nested config
				setFlagsFromGenericMap(&nested, append(ancestors, k), fs)
				continue
			}
		}

		// Set the flag name correctly from path
		k = strings.Join(append(ancestors, k), "_")

		// Make sure flag exists
		if flg := fs.Lookup(k); flg == nil {
			panic(fmt.Sprintf("unknown configuration file field: %s", k))
		}

		// Reformat slice/array values so that pflag Values can parse them
		// If the value is a slice, join the values
		if reflect.ValueOf(v).Kind() == reflect.Slice {

			writeCsv := false
			vals := make([]string, len(v.([]any)))

			// Populate vals and check if we need to write csv
			for idx, val := range v.([]any) {
				vals[idx] = fmt.Sprintf("%v", val)
				if strings.Contains(vals[idx], `"`) || strings.Contains(vals[idx], `,`) {
					writeCsv = true
				}
			}

			// csv write for string type slices
			if writeCsv {
				b := &bytes.Buffer{}
				w := csv.NewWriter(b)
				w.Write(vals)
				w.Flush()
				v = b.String()

			} else {
				v = strings.Join(vals, ",")
			}
		}

		// Set the value
		if err := setFlagValue(k, fmt.Sprintf("%v", v), fs); err != nil {
			panic(fmt.Sprintf("unable to set value for %s: %v", k, err))
		}
	}
}
