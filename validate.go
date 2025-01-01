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
	"regexp"
	"strings"

	"github.com/fatih/structtag"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	extras "github.com/go-playground/validator/v10/non-standard/validators"
	en_tr "github.com/go-playground/validator/v10/translations/en"
	"github.com/spf13/pflag"
)

// getValidatorTranslator registers translations for v *validator.Validate object
//
// It takes a *validator.Validate object as a parameter and returns a ut.Translator object.
func getValidatorTranslator(v *validator.Validate) ut.Translator {
	en := en.New()
	uni := ut.New(en, en)

	trans, _ := uni.GetTranslator("en")

	en_tr.RegisterDefaultTranslations(v, trans)

	v.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} is required", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())

		return t
	})

	v.RegisterTranslation("not_blank", trans, func(ut ut.Translator) error {
		return ut.Add("not_blank", "{0} must contain at least 1 non-whitespace character", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("not_blank", fe.Field())
		return t
	})

	v.RegisterTranslation("file", trans, func(ut ut.Translator) error {
		return ut.Add("file", "{0} {1} does not exist, is not a file, or is not accessible", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("file", fe.Field(), fmt.Sprintf("%s", fe.Value()))

		return t
	})

	v.RegisterTranslation("dir", trans, func(ut ut.Translator) error {
		return ut.Add("dir", "{0} {1} does not exist, is not a directory, or is not accessible", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("dir", fe.Field(), fmt.Sprintf("%s", fe.Value()))

		return t
	})

	v.RegisterTranslation("regex", trans, func(ut ut.Translator) error {
		return ut.Add("regex", "{0} {1} does not match pattern {2}", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("regex", fe.Field(), fmt.Sprintf("%s", fe.Value()), fe.Param())
		return t
	})
	return trans
}

// getDefaultValidator returns a new instance of the validator.Validate struct with the required struct enabled and
// the "not_blank" and "regex" validations registered.
//
// Returns a pointer to the validator.Validate struct.
func getDefaultValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterValidation("not_blank", extras.NotBlank)
	v.RegisterValidation("regex", validateRegex)

	return v
}

// validateRegex validates a FieldLevel (interface) with a regular expression.
//
// fl: the validator.FieldLevel
// bool: returns true if the field matches the regular expression, else false.
func validateRegex(fl validator.FieldLevel) bool {
	field := fl.Field()
	for field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return false
		}
		field = field.Elem()
	}
	switch field.Kind() {
	case reflect.String:
		val := field.String()
		if r, err := regexp.Compile(fl.Param()); err != nil {
			panic("error compiling regex: " + err.Error())
		} else if !r.MatchString(val) {
			return false
		}
		return true
	default:
		return false
	}
}

// validate configuration
func (c *configurer) validate(s interface{}, fs *pflag.FlagSet) {

	v := getDefaultValidator()
	trans := getValidatorTranslator(v)

	errors := []string{}
	fldToNameMap := make(map[string]string)
	c.visitFields(s, func(f reflect.StructField, tags *structtag.Tags, v reflect.Value, ancestors []string) (stop bool) {

		fName := fieldNameToConfigName(f.Name, tags, ancestors)
		fldToNameMap[f.PkgPath+"."+f.Name] = fName

		// Check enums
		if val, err := tags.Get("enum"); err == nil && val.Value() != "" {
			enums := strings.Split(val.Value(), ",")
			v := fs.Lookup(fName).Value.String()
			for _, e := range enums {
				if v == e {
					return false // false == don't stop looping over fields
				}
			}
			errors = append(errors, fmt.Sprintf("%s must be one of %s", fName, strings.Join(enums, ", ")))
		}
		return false // false == don't stop looping over fields
	}, []string{})

	// Register function that does field to name translation
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		if name, ok := fldToNameMap[f.PkgPath+"."+f.Name]; ok {
			return name
		}
		return f.PkgPath + "." + f.Name
	})

	// Validate config struct
	if err := v.Struct(s); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			errors = append(errors, e.Translate(trans))
		}
	}
	if len(errors) > 0 {
		panic(fmt.Sprintf("validation failed; %s", strings.Join(errors, ", ")))
	}
}
