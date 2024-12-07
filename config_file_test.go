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
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	co "github.com/imoore76/configurature"
)

func TestConfigFile(t *testing.T) {
	assert := assert.New(t)

	tmp, _ := os.CreateTemp("", "ldlm-test-*.yml")
	defer os.Remove(tmp.Name())
	tmp.Write([]byte("foo_int: 4\nsub_foo_string: 'yes'\nkeepalive_timeout: 3m\nbool: true\n"))
	tmp.Close()

	c := co.Configure[TestConfigFileStruct](&co.Options{
		NoRecover: true,
		Args:      []string{"--cool_file", tmp.Name()},
	})
	assert.Equal(uint32(4), c.FooInt)
	assert.Equal("yes", c.SubFooString)
	assert.Equal(time.Duration(3)*time.Minute, c.KeepaliveTimeout)
	assert.Equal(true, c.Bool)
}

func TestConfigFile_BadField(t *testing.T) {
	if os.Getenv("TEST_PASSTHROUGH") == "1" {
		tmp, _ := os.CreateTemp("", "ldlm-test-*.yml")
		defer os.Remove(tmp.Name())
		tmp.Write([]byte("foo_int: 4\nsub_string: 'yes'\nkeepalive_timeout: 3m\nbool: true\n"))
		tmp.Close()

		co.Configure[TestConfigFileStruct](&co.Options{
			Args: []string{"--cool_file", tmp.Name()},
		})
		os.Exit(0)
	}
	assert := assert.New(t)
	stdout, stderr := runExternal(t)

	assert.Equal("", stdout)
	assert.Equal("error parsing configuration: unknown configuration file field: sub_string\n", stderr)

}

func TestConfigFile_BadValue(t *testing.T) {
	if os.Getenv("TEST_PASSTHROUGH") == "1" {
		tmp, _ := os.CreateTemp("", "ldlm-test-*.yml")
		defer os.Remove(tmp.Name())
		tmp.Write([]byte("foo_int: asdf\nsub_foo_string: 'ok'\nkeepalive_timeout: 3m\nbool: true\n"))
		tmp.Close()

		co.Configure[TestConfigFileStruct](&co.Options{
			Args: []string{"--cool_file", tmp.Name()},
		})
		os.Exit(0)
	}
	assert := assert.New(t)
	stdout, stderr := runExternal(t)

	assert.Equal("", stdout)
	assert.Equal("error parsing configuration: unable to set value for foo_int: "+
		"strconv.ParseUint: parsing \"asdf\": invalid syntax\n", stderr)

}

func TestConfigFile_UnsupportedType(t *testing.T) {
	if os.Getenv("TEST_PASSTHROUGH") == "1" {
		tmp, _ := os.CreateTemp("", "ldlm-test-*.foo")
		defer os.Remove(tmp.Name())
		tmp.Write([]byte("foo_int: asdf\nsub_foo_string: 'ok'\nkeepalive_timeout: 3m\nbool: true\n"))
		tmp.Close()

		co.Configure[TestConfigFileStruct](&co.Options{
			Args: []string{"--cool_file", tmp.Name()},
		})
		os.Exit(0)
	}
	assert := assert.New(t)
	stdout, stderr := runExternal(t)

	assert.Equal("", stdout)
	assert.True(strings.HasPrefix(stderr, "error parsing configuration: unsupported config file type: "), stderr)
	assert.True(strings.HasSuffix(stderr, "Supported file types are .json, .yml, .yaml\n"))

}

func TestNested_Yaml(t *testing.T) {
	assert := assert.New(t)

	tmp, _ := os.CreateTemp("", "ldlm-test-*.yml")
	defer os.Remove(tmp.Name())
	tmp.Write([]byte(`s_slice: [af, 'as"oidjf', 3]
my_map:
  foo: bar
  a: b
name_age_map:
  joshua: 33
  josh: 24
  dave: 35
sub:
  default_lock_timeout: 30s
  foo_seconds: 20
  foo_int: 2
  foo_ints: 
    - 2
    - 4
    - 5
os:
  sub_foo_string: "there and everywhere"
`))
	tmp.Close()

	c := co.Configure[TestNestedConfig](&co.Options{
		Args:      []string{"--cool_file", tmp.Name()},
		NoRecover: true,
	})

	assert.Equal(map[string]int{"joshua": 33, "josh": 24, "dave": 35}, c.NameAgeMap, "NameAgeMap should be joshua[33] josh[24] dave[35]")
	assert.Equal(map[string]string{"foo": "bar", "a": "b"}, c.MyMap, "MyMap should be foo[bar] a[b]")
	assert.Equal([]string{"af", "as\"oidjf", "3"}, c.SSlice, "SSlie should be af,as\"oidjf,3")
	assert.Equal(time.Duration(30)*time.Second, c.Sub.DefaultLockTimeout, "DefaultLockTimeout should be 30s")
	assert.Equal(uint(20), c.Sub.FooSeconds, "FooSeconds should be 20")
	assert.Equal(uint32(2), c.Sub.FooInt, "FooInt should be 2")
	assert.Equal([]uint{2, 4, 5}, c.Sub.FooInts, "FooInts should be [2 4 5]")
	assert.Equal("there and everywhere", c.OS.SubFooString, "SubFooString should be there and everywhere")
}

func TestNested_JSON(t *testing.T) {
	assert := assert.New(t)

	tmp, _ := os.CreateTemp("", "ldlm-test-*.json")
	defer os.Remove(tmp.Name())
	tmp.Write([]byte(`{"s_slice": ["af", "asoidjf", "3"], "my_map": {"foo": "bar", "baz": "qux"}, "sub": {"default_lock_timeout": "30s", "foo_seconds": 20, "foo_int": 2, "foo_ints": [2, 4, 5]}, "os": {"sub_foo_string": "there and everywhere"}}`))
	tmp.Close()

	c := co.Configure[TestNestedConfig](&co.Options{
		Args:      []string{"--cool_file", tmp.Name()},
		NoRecover: true,
	})

	assert.Equal(map[string]string{"foo": "bar", "baz": "qux"}, c.MyMap, "MyMap should be foo[bar] baz[qux]")
	assert.Equal([]string{"af", "asoidjf", "3"}, c.SSlice, "SSlie should be af,asoidjf,3")
	assert.Equal(time.Duration(30)*time.Second, c.Sub.DefaultLockTimeout, "DefaultLockTimeout should be 30s")
	assert.Equal(uint(20), c.Sub.FooSeconds, "FooSeconds should be 20")
	assert.Equal(uint32(2), c.Sub.FooInt, "FooInt should be 2")
	assert.Equal("[2 4 5]", fmt.Sprintf("%v", c.Sub.FooInts), "FooInts should be [2 4 5]")
	assert.Equal("there and everywhere", c.OS.SubFooString, "SubFooString should be there and everywhere")
}
