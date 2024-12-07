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
	"strings"

	"github.com/fatih/structtag"
	"github.com/iancoleman/strcase"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

// Internal flags that should not be printed
var internalFlags = map[string]bool{
	"help":                true,
	"print_env_template":  true,
	"print_yaml_template": true,
}

// printEnvTemplate prints the usage information for environment variables
// based on the provided flag set.
//
// Parameters:
// - fs: the flag set containing the flag values
func (c *configurer) printEnvTemplate(fs *flag.FlagSet) {
	fmt.Printf("# Generated with\n# %s\n\n", c.opts.Args)
	fs.VisitAll(func(f *flag.Flag) {
		if _, ok := internalFlags[f.Name]; ok || f.Hidden {
			return
		}
		fmt.Printf("# %s\n", f.Usage)
		fmt.Printf("%s%s", c.opts.EnvPrefix, strcase.ToScreamingSnake(f.Name))
		fmt.Printf("=\"%s\"\n\n", strings.Replace(f.Value.String(), "\"", "\\\"", -1))
	})
}

// printYamlTemplate prints the usage information for YAML based on the
// configurature struct
//
// Parameters:
// - fs: the flag set containing the flag values
func (c *configurer) printYamlTemplate(fs *flag.FlagSet) {

	fmt.Printf("# Generated with\n# %s\n\n", c.opts.Args)

	ancestorsSeen := map[string]bool{}
	c.visitFields(c.config, func(f reflect.StructField, tags *structtag.Tags, v reflect.Value, ancestors []string) (stop bool) {
		if v.Elem().Type() == configFileType {
			return
		}

		fName := fieldNameToConfigName(f.Name, tags, ancestors)
		fl := fs.Lookup(fName)

		if _, ok := internalFlags[fl.Name]; ok || fl.Hidden {
			return
		}

		indent := strings.Repeat("  ", len(ancestors))

		if len(ancestors) > 0 {
			parent := ancestors[len(ancestors)-1]
			if ok := ancestorsSeen[parent]; !ok {
				ancestorsSeen[parent] = true
				fmt.Printf("%s%s:\n\n", strings.Repeat("  ", len(ancestors)-1), parent)
			}
		}

		ymlVal := strings.Builder{}
		encoder := yaml.NewEncoder(&ymlVal)
		encoder.Encode(map[string]interface{}{
			stripAncestors(fName, ancestors): v.Elem().Interface(),
		})
		encoder.Close()

		fmt.Printf("%s# %s\n", indent, fl.Usage)
		// Indent yaml string to current level
		ymlValStr := indent + strings.Replace(ymlVal.String(), "\n", "\n"+indent, strings.Count(ymlVal.String(), "\n")-1)
		fmt.Println(ymlValStr)

		return stop
	}, []string{})
}

// stripAncestors removes the ancestors from the given name.
//
// name: the name string to remove ancestors from.
// ancestors: a slice of strings representing the ancestors to remove.
// return: the name string with the ancestors removed.
func stripAncestors(name string, ancestors []string) string {
	s, _ := strings.CutPrefix(name, strings.Join(ancestors, "_")+"_")
	return s
}
