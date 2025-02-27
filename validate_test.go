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
	"fmt"
	"testing"

	co "github.com/imoore76/configurature"
	"github.com/stretchr/testify/assert"
)

type ValidConfig struct {
	FooInt      uint32         `desc:"FooInt" default:"100" validate:"gt=5,lt=200"`
	FooInts     []uint         `desc:"FooInts" default:"100" validate:"dive,gt=5,lt=200"`
	FooFloat    float32        `desc:"FooFloat" default:"4.14" validate:"gt=4,lt=200"`
	FooFloats   []float64      `desc:"FooFloats" default:"100,3.12" validate:"dive,gt=3,lt=200"`
	SliceLen    []int          `desc:"SliceLen" default:"100,2" validate:"gte=2,lt=5"`
	PhoneNumber string         `desc:"PhoneNumber" default:"222-222-2224" validate:"regex=^\\d{3}-\\d{3}-\\d{4}$"`
	File        string         `desc:"File" default:"./go.mod" validate:"file"`
	Dir         string         `desc:"Dir" default:"/" validate:"dir"`
	OneOf       string         `desc:"OneOf" default:"a" enum:"a,b,c"`
	Empty       string         `desc:"Empty" default:"here" validate:"not_blank"`
	URL         string         `desc:"URL" default:"https://github.com/spf13/cobra" validate:"url"`
	NameAges    map[string]int `desc:"NameAgeMap" default:"a=22,b=35,c=48" validate:"dive,gt=18,lt=67"`
}

type ValidConfigMustSet struct {
	MustSet string `desc:"MustSet" default:"" validate:"required,not_blank"`
}

type ValidConfigCustom struct {
	Color string `desc:"Color" validate:"starts_with=red"`
}

func TestValidation_Defaults(t *testing.T) {
	assert := assert.New(t)

	c := co.Configure[ValidConfig](&co.Options{NoRecover: true})
	assert.Equal(uint32(100), c.FooInt, "FooInt should be 100")
	assert.Equal([]uint{100}, c.FooInts, "FooInts should be [100]")
	assert.Equal(float32(4.14), c.FooFloat, "FooFloat should be 4.14")
	assert.Equal([]float64{100, 3.12}, c.FooFloats, "FooFloats should be [100, 3.12]")
	assert.Equal([]int{100, 2}, c.SliceLen, "SliceLen should be [100, 2]")
	assert.Equal("222-222-2224", c.PhoneNumber, "PhoneNumber should be 222-222-2224")
	assert.Equal("./go.mod", c.File, "File should be ./go.mod")
	assert.Equal("/", c.Dir, "Dir should be /")
	assert.Equal("a", c.OneOf, "OneOf should be a")
	assert.Equal("here", c.Empty, "Empty should be 'here'")
	assert.Equal("https://github.com/spf13/cobra", c.URL, "URL should be https://github.com/spf13/cobra")
	assert.Equal(map[string]int{"a": 22, "b": 35, "c": 48}, c.NameAges, "NameAges should be a=22,b=35,c=48")
}

func TestValidation_RequiredNotBlank(t *testing.T) {
	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Sprintf("%v", r)
			}
		}()
		co.Configure[ValidConfigMustSet](&co.Options{
			Args:      []string{},
			NoRecover: true,
		})
	}()

	assert.Equal(t, "validation failed; must_set is required", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfigMustSet](&co.Options{
			Args:      []string{"--must_set=", " "},
			NoRecover: true,
		})
	}()

	assert.Equal(t, "validation failed; must_set is required", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfigMustSet](&co.Options{
			Args:      []string{"--must_set=107"},
			NoRecover: true,
		})
	}()

	assert.Equal(t, "", err)

}

func TestValidation_FooInt(t *testing.T) {
	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--foo_int=2"},
			NoRecover: true,
		})
	}()

	assert.Equal(t, "validation failed; foo_int must be greater than 5", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--foo_int=200000"},
			NoRecover: true,
		})
	}()

	assert.Equal(t, "validation failed; foo_int must be less than 200", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--foo_int=107"},
			NoRecover: true,
		})
	}()

	assert.Equal(t, "", err)

}

func TestValidation_FooInts(t *testing.T) {
	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--foo_ints=2,5,7"},
			NoRecover: true,
		})
	}()

	assert.Equal(t, "validation failed; foo_ints[0] must be greater than 5, foo_ints[1] must be greater than 5", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--foo_ints=200000,5,6"},
			NoRecover: true,
		})
	}()

	assert.Equal(t, "validation failed; foo_ints[0] must be less than 200, foo_ints[1] must be greater than 5", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--foo_ints=107,34,22"},
			NoRecover: true,
		})
	}()

	assert.Equal(t, "", err)

}

func TestValidation_FooFloat(t *testing.T) {
	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--foo_float=2.3"},
			NoRecover: true,
		})
	}()

	assert.Equal(t, "validation failed; foo_float must be greater than 4", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--foo_float=2000.40"},
			NoRecover: true,
		})
	}()

	assert.Equal(t, "validation failed; foo_float must be less than 200", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--foo_float=107.6"},
			NoRecover: true,
		})
	}()

	assert.Equal(t, "", err)
}

func TestValidation_FooFloats(t *testing.T) {
	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--foo_floats=2.6,5.2,7.8"},
			NoRecover: true,
		})
	}()

	assert.Equal(t, "validation failed; foo_floats[0] must be greater than 3", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--foo_floats=200000.2,5.8,6.2,2"},
			NoRecover: true,
		})
	}()

	assert.Equal(t, "validation failed; foo_floats[0] must be less than 200, foo_floats[3] must be greater than 3", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--foo_floats=107.8,34.0,22"},
			NoRecover: true,
		})
	}()

	assert.Equal(t, "", err)

}

func TestValidation_SliceLen(t *testing.T) {
	assert := assert.New(t)
	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--slice_len=1"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; slice_len must contain at least 2 items", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--slice_len=1,23,4,5,6,34,5"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; slice_len must contain less than 5 items", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--slice_len=107,204,6"},
			NoRecover: true,
		})
	}()

	assert.Equal("", err)

}

func TestValidation_Regex(t *testing.T) {
	assert := assert.New(t)
	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--phone_number=123-4234-2342"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; phone_number 123-4234-2342 does not match pattern ^\\d{3}-\\d{3}-\\d{4}$", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--phone_number=123-232-32"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; phone_number 123-232-32 does not match pattern ^\\d{3}-\\d{3}-\\d{4}$", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--phone_number=123-222-8765"},
			NoRecover: true,
		})
	}()

	assert.Equal("", err)
}

func TestValidation_FileExists(t *testing.T) {
	assert := assert.New(t)
	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--file=./not_exist.txt"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; file ./not_exist.txt does not exist, is not a file, or is not accessible", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--file=/"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; file / does not exist, is not a file, or is not accessible", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--file=./LICENSE"},
			NoRecover: true,
		})
	}()

	assert.Equal("", err)

}

func TestValidation_DirExists(t *testing.T) {
	assert := assert.New(t)
	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--dir=./not_exist"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; dir ./not_exist does not exist, is not a directory, or is not accessible", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--dir=./LICENSE"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; dir ./LICENSE does not exist, is not a directory, or is not accessible", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--dir=./"},
			NoRecover: true,
		})
	}()

	assert.Equal("", err)

}

func TestValidation_Enum(t *testing.T) {
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

	assert.Equal("validation failed; one_of must be one of a, b, c", err)

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

	assert.Equal("validation failed; one_of must be one of a, b, c", err)

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

func TestValidation_NotBlank(t *testing.T) {
	assert := assert.New(t)
	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--empty="},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; empty must contain at least 1 non-whitespace character", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--empty=1"},
			NoRecover: true,
		})
	}()

	assert.Equal("", err)

}

func TestValidation_URL(t *testing.T) {
	assert := assert.New(t)
	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--url=yahoo.com"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; url must be a valid URL", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--url=yahoo.com/test?foo=bar"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; url must be a valid URL", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--url=ftp://site.net/test?foo=bar"},
			NoRecover: true,
		})
	}()

	assert.Equal("", err)
}

func TestValidation_MapElement(t *testing.T) {
	assert := assert.New(t)
	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--name_ages", "a=1,b=22"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; name_ages[a] must be greater than 18", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--name_ages", "a=99,b=22,c=33"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; name_ages[a] must be less than 67", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[ValidConfig](&co.Options{
			Args:      []string{"--name_ages", "a=22,asdf=55,foo=43"},
			NoRecover: true,
		})
	}()

	assert.Equal("", err)
}

func TestValidation_ParentOnlySlice(t *testing.T) {
	type TConf struct {
		StringSlice []string `desc:"string slice" validate:"lt=5,gt=2"`
	}
	assert := assert.New(t)

	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[TConf](&co.Options{
			Args:      []string{"--string_slice=a,b,c,d,e,f"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; string_slice must contain less than 5 items", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[TConf](&co.Options{
			Args:      []string{"--string_slice=a"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; string_slice must contain more than 2 items", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[TConf](&co.Options{
			Args:      []string{"--string_slice=a,b,c,d"},
			NoRecover: true,
		})
	}()

	assert.Equal("", err)

}

func TestValidation_ParentOnlyMap(t *testing.T) {
	type TConf struct {
		StringMap map[string]string `desc:"string map" validate:"lt=5,gt=2"`
	}
	assert := assert.New(t)

	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[TConf](&co.Options{
			Args:      []string{"--string_map", "a=x,b=y,c=z,d=w,e=bb,f=d"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; string_map must contain less than 5 items", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[TConf](&co.Options{
			Args:      []string{"--string_map", "foo=bar"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; string_map must contain more than 2 items", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[TConf](&co.Options{
			Args:      []string{"--string_map", "a=b,b=c,c=d,d=e"},
			NoRecover: true,
		})
	}()

	assert.Equal("", err)

}

func TestValidation_ChildOnlySlice(t *testing.T) {
	assert := assert.New(t)
	type TConf struct {
		StringSlice []string `desc:"string slice" validate:"dive,lt=5,gt=2"`
	}

	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[TConf](&co.Options{
			Args:      []string{"--string_slice=aasdf,ba,caspdiufhu"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; string_slice[0] must be less than 5 "+
		"characters in length, string_slice[1] must be greater than 2 "+
		"characters in length, string_slice[2] must be less than 5 "+
		"characters in length", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[TConf](&co.Options{
			Args:      []string{"--string_slice=a"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; string_slice[0] must be greater than 2 characters in length", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[TConf](&co.Options{
			Args:      []string{"--string_slice=abc,bdef,ceg,dec"},
			NoRecover: true,
		})
	}()

	assert.Equal("", err)
}

func TestValidation_ChildOnlyMap(t *testing.T) {
	type TConf struct {
		StringMap map[string]string `desc:"string map" validate:"dive,lt=5,gte=2"`
	}
	assert := assert.New(t)

	err := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[TConf](&co.Options{
			Args:      []string{"--string_map", "a=xd,d=wasdfasdfsfsd"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; string_map[d] must be less than 5 characters in length", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[TConf](&co.Options{
			Args:      []string{"--string_map", "foo=b,bar=asdf"},
			NoRecover: true,
		})
	}()

	assert.Equal("validation failed; string_map[foo] must be at least 2 characters in length", err)

	func() {
		err = ""
		defer func() {
			if r := recover(); r != nil {
				err = r.(string)
			}
		}()
		co.Configure[TConf](&co.Options{
			Args:      []string{"--string_map", "a=bfsd,b=cef,c=das,d=ee"},
			NoRecover: true,
		})
	}()

	assert.Equal("", err)
}

func TestValidation_NestedFieldNames(t *testing.T) {
	type T6 struct {
		SubFooString string `desc:"Something" default:"t6there" validate:"lt=5"`
	}
	type T5 struct {
		T5Str string `desc:"t5" default:"t5"`
		Not   T6     `name:"buried"`
	}
	type T4 struct {
		T4Str        string `desc:"t4" default:"t4"`
		SubFooString string `desc:"T4 Something" default:"t4there"`
		Thing        T5
	}
	type T3 struct {
		T3Str string `desc:"t3" default:"t3"`
		Other T4
	}
	type T2 struct {
		T2Str        string `desc:"t2" default:"t2"`
		SubFooString string `desc:"T2 Something" default:"t2there"`
		Something    T3     `name:"server"`
	}
	type T1 struct {
		T1Str string `desc:"t1" default:"t1"`
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

	assert.Equal(t, "validation failed; "+
		"sub_server_other_thing_buried_sub_foo_string must be less "+
		"than 5 characters in length", err, err)

}

func TestValidation_Ptrs(t *testing.T) {
	type T struct {
		Ptr *string `desc:"ptr" validate:"required,not_blank,gt=4,alpha,regex=^aa[a-z]{3}$"`
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
			Args:      []string{"--ptr=aasdf"},
		})
	}()
	assert.Equal(t, "", err)
}

func TestValidation_RequireNoDefaults(t *testing.T) {
	type T struct {
		MyStringReq     string `validate:"required"`
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
	assert.Equal(t, "validation failed; my_string_not_req is required, my_string_req is required", err)

}
