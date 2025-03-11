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

package configurature_test

import (
	"testing"

	co "github.com/imoore76/configurature"
	"github.com/stretchr/testify/assert"
)

func TestValidation_Enum(t *testing.T) {
	type ValidConfig struct {
		OneOf string `config:"one_of" enum:"a,b,c"`
	}

	assert := assert.New(t)
	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--one_of=d"},
			NoRecover: true,
		})
	}()

	assert.Equal("one_of must be one of a, b, c", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--one_of=a,b"},
			NoRecover: true,
		})
	}()

	assert.Equal("one_of must be one of a, b, c", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--one_of=c"},
			NoRecover: true,
		})
	}()

	assert.Equal("", err)

}

func TestValidation_Required(t *testing.T) {
	type T struct {
		MyStringReq             string `required:""`
		MyStringEnum            string `required:"" enum:"a,b,c"`
		MyStringEnumNotRequired string `enum:"a,b,c" default:"a"`
	}
	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[T](&co.Options{
			NoRecover: true,
		})
	}()

	assert.Equal(t, "my_string_req is required, my_string_enum must be one of a, b, c", err)
}

func TestValidation_NestedFieldNames(t *testing.T) {
	type T6 struct {
		SubFooString string `help:"Something" required:""`
	}
	type T5 struct {
		T5Str string `help:"t5" default:"t5"`
		Not   T6     `name:"buried"`
	}
	type T4 struct {
		T4Str        string `help:"t4" default:"t4"`
		SubFooString string `help:"T4 Something" default:"t4there"`
		Thing        T5
	}
	type T3 struct {
		T3Str string `help:"t3" default:"t3"`
		Other T4
	}
	type T2 struct {
		T2Str        string `help:"t2" default:"t2"`
		SubFooString string `help:"T2 Something" default:"t2there"`
		Something    T3     `name:"server"`
	}
	type T1 struct {
		T1Str string `help:"t1" default:"t1"`
		Sub   T2
	}

	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[T1](&co.Options{
			NoRecover: true,
		})
	}()

	assert.Equal(t, "sub_server_other_thing_buried_sub_foo_string is required", err, err)

}

func TestValidation_RequireNoDefaults(t *testing.T) {
	type T struct {
		MyStringReq     string `required:""`
		MyStringNotReq  string
		MyStringDefault string `default:"foo"`
	}
	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[T](&co.Options{
			NoRecover:         true,
			RequireNoDefaults: true,
		})
	}()
	assert.Equal(t, "my_string_req is required, my_string_not_req is required", err)

}
