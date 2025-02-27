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
This file provides the Configure function and its helpers
*/
package configurature

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/fatih/structtag"
	"github.com/iancoleman/strcase"
	"github.com/spf13/pflag"
)

// Used for config file special casing in code
var configFileType = reflect.TypeFor[ConfigFile]()

// configurer is used to populate a config struct
type configurer struct {
	config     any
	opts       *Options
	configFile struct {
		Flag  string
		Short string
		Value *string
	}
}

// Configure options
type Options struct {
	EnvPrefix         string               // Prefix for environment variables
	Args              []string             // Arguments to parse
	NilPtrs           bool                 // Leave pointers set to nil if values aren't specified
	Usage             func(*pflag.FlagSet) // Usage function called when configuration is incorrect or for --help
	NoRecover         bool                 // Don't recover from panic
	ShowInternalFlags bool                 // Show hidden internal flags
	NoShortHelp       bool                 // Don't add "h" as a short help flag
	RequireNoDefaults bool                 // Require any fields that don't have a default value
}

// Configure will populate the supplied struct with options specified on the
// command line or by environment variables prefixed by the specified envPrefix
func Configure[T any](opts *Options) *T {
	if opts == nil {
		opts = &Options{
			Args: os.Args[1:],
		}
	} else if opts.Args == nil {
		opts.Args = os.Args[1:]
	}

	c := &configurer{
		config: new(T),
		opts:   opts,
	}

	// Create a flagset
	f := flagSetFromOptions(opts)

	// This is a chicken and egg situation where we need to parse flags to
	// determine what the config file flags are, but we want to load the config
	// from the file first so that flags specified on the command line or via
	// env vars can overwrite any values set in the config file.

	// Set config file pointer if specified. loadFlags() needs to know what the
	// flag is for specifying the file.
	c.setConfigFile()

	// loadFlags(), makes the flagset is aware of all the config fields. It
	// returns setters that will actually set the config values after args have
	// been parsed.
	setters := c.loadFlags(c.config, f)

	// Recover from panic and print error
	if !opts.NoRecover {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "error parsing configuration: %s\n", r)
				os.Exit(1)
			}
		}()
	}

	// Load config file if the pointer was set by setConfigFile
	if c.configFile.Value != nil {
		c.loadConfigFile(f)
	}

	// Load values from environment
	if opts.EnvPrefix != "" {
		c.setFromEnv(c.config, f)
	}

	// Parse CLI args into flagset and run flag setter functions
	f.Parse(opts.Args)
	for _, fn := range setters {
		fn()
	}

	// Show usage if requested
	if help, _ := f.GetBool("help"); help {
		f.Usage()
	}

	// Generate .env template
	if ok, _ := f.GetBool("print_env_template"); ok {
		c.printEnvTemplate(f)
		os.Exit(0)
	}

	// Generate YAML template
	if ok, _ := f.GetBool("print_yaml_template"); ok {
		c.printYamlTemplate(f)
		os.Exit(0)
	}

	// Validate config
	c.validate(c.config, f)

	// Used by Get[T]()
	setLastConfig(c.config)

	return c.config.(*T)
}

// setFromEnv sets configuration values from environment
func (c *configurer) setFromEnv(s any, fs *pflag.FlagSet) {

	c.visitFields(s, func(f reflect.StructField, tags *structtag.Tags, v reflect.Value, ancestors []string) (stop bool) {
		fName := fieldNameToConfigName(f.Name, tags, ancestors)
		envVal := os.Getenv(
			fmt.Sprintf("%s%s", c.opts.EnvPrefix, strcase.ToScreamingSnake(fName)),
		)
		if envVal != "" {
			if err := setFlagValue(fName, envVal, fs); err != nil {
				panic(fmt.Sprintf("setFromEnv(): error setting value of field %s: %v", f.Name, err))
			}
		}
		return stop
	}, []string{})
}

// loadFlags() sets field values based on options specified on the command line
// or by environment variables
func (c *configurer) loadFlags(s any, fl *pflag.FlagSet) []func() {

	setters := []func(){}

	c.visitFields(s, func(f reflect.StructField, tags *structtag.Tags, v reflect.Value, ancestors []string) (stop bool) {

		fName := fieldNameToConfigName(f.Name, tags, ancestors)
		descTag, err := tags.Get("desc")
		if err != nil {
			descTag = &structtag.Tag{
				Key: "desc",
				Name: strings.ReplaceAll(
					fieldNameToConfigName(f.Name, tags, ancestors), "_", " ",
				),
			}
		}
		shortTag, _ := tags.Get("short")
		if shortTag == nil {
			shortTag = &structtag.Tag{}
		}
		noDefault := false
		defaultTag, _ := tags.Get("default")
		if defaultTag == nil {
			noDefault = true
			defaultTag = &structtag.Tag{}
		}

		// Special case for ConfigFile field
		if v.Elem().Type() == configFileType {
			c.configFile.Flag = fName
			c.configFile.Short = shortTag.Value()
		}

		desc := descTag.Value()
		enumProvided := false
		if enums, _ := tags.Get("enum"); enums != nil && enums.Value() != "" {
			desc += fmt.Sprintf(" (%s)", strings.Replace(enums.Value(), ",", "|", -1))
			enumProvided = true
		}
		addToFlagSet(v.Type(), enumProvided, fl, fName, shortTag.Value(), defaultTag.Value(), desc)

		// Hide hidden flags
		if _, err := tags.Get("hidden"); err == nil {
			fl.MarkHidden(fName)
		}

		isPtr := v.Kind() == reflect.Ptr
		setters = append(setters, func() {
			// Don't set pointers if
			// * No default value was provided
			// * the NilPtrs option is set
			// * the value hasn't changed (wasn't specified)
			if noDefault && c.opts.NilPtrs && isPtr && !fl.Lookup(fName).Changed {
				return
			}
			setNativeValue(v, fName, fl)
		})

		return false
	}, []string{})

	return setters
}

// visitFields visits the fields of the config struct and calls the
// provided function on each field.
func (c *configurer) visitFields(s any, f func(reflect.StructField, *structtag.Tags, reflect.Value, []string) bool, ancestors []string) bool {
	v := reflect.ValueOf(s).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {

		if !t.Field(i).IsExported() {
			continue
		}

		// Parse tags
		tags, err := structtag.Parse(string(t.Field(i).Tag))
		if err != nil {
			panic(fmt.Sprintf("error parsing field %s tags: %s", t.Field(i).Name, err.Error()))
		}

		// Skip any fields tagged with ignore:""
		if _, err := tags.Get("ignore"); err == nil {
			continue
		}

		// Handle anonymous struct fields, which are sub-configs
		if t.Field(i).Anonymous {
			fld := v.Field(i).Addr().Interface()
			if stop := c.visitFields(fld, f, ancestors); stop {
				return true
			}
			continue
		}

		// Handle nested config structs
		if t.Field(i).Type.Kind() == reflect.Struct {
			fld := v.Field(i).Addr().Interface()
			fName := t.Field(i).Name
			if name, err := tags.Get("name"); err == nil {
				fName = name.Value()
			}

			var newAncestors []string
			if fName != "" {
				newAncestors = append(ancestors, strcase.ToSnake(fName))
			} else {
				newAncestors = ancestors
			}
			if stop := c.visitFields(fld, f, newAncestors); stop {
				return true
			}
			continue
		}

		// Call function on field and stop if it returns true
		if f(t.Field(i), tags, v.Field(i).Addr(), ancestors) {
			return true
		}
	}
	return false
}

// fieldNameToConfigName converts a struct field name and its ancestor path to
// its flag name
func fieldNameToConfigName(name string, tags *structtag.Tags, ancestors []string) string {
	if nm, err := tags.Get("name"); err == nil && nm.Value() != "" {
		name = nm.Value()
	}
	return strings.Join(append(ancestors, strcase.ToSnake(name)), "_")
}

// flagSetFromOptions creates and returns a *pflag.FlagSet based on the
// provided options
func flagSetFromOptions(opts *Options) *pflag.FlagSet {

	f := pflag.NewFlagSet("config", pflag.ExitOnError)

	// Set up help flag
	if opts.NoShortHelp {
		f.Bool("help", false, "show help and exit")
	} else {
		f.BoolP("help", "h", false, "show help and exit")
	}

	// Set Usage function
	if opts.Usage != nil {
		f.Usage = func() { opts.Usage(f) }
	} else {
		f.Usage = func() {
			fmt.Println("Command usage:")
			fmt.Println(f.FlagUsages())
			os.Exit(0)
		}
	}

	// print_env_template flag setup
	f.Bool("print_env_template", false, "Print example environment variables and exit")
	if !opts.ShowInternalFlags {
		f.MarkHidden("print_env_template")
	}

	// print_yaml_template flag setup
	f.Bool("print_yaml_template", false, "Print example YAML config file and exit")
	if !opts.ShowInternalFlags {
		f.MarkHidden("print_yaml_template")
	}

	return f
}
