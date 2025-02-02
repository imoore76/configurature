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
	"net"
	"os"
	"testing"

	co "github.com/imoore76/configurature"
	"github.com/stretchr/testify/assert"
)

type LowerConf struct {
	Ages map[string]int `desc:"Names and ages map" default:"a=1,b=2,c=3"`
}
type SubConf1 struct {
	IP    net.IP `desc:"IP address on which to listen" default:"127.0.0.1"`
	Lower LowerConf
}

type SubConf2 struct {
	IPs *[]net.IP `desc:"IP addresses" default:"127.0.0.1" name:"ips"`
}

type YamlConf struct {
	Conf  co.ConfigFile `desc:"configuration file" short:"c"`
	Str   string        `desc:"a string" default:"yes\"no"`
	Strs  []string      `desc:"a string slice" default:"1,2,3"`
	StPtr *string       `desc:"a string pointer"`
	Int1  int           `desc:"an int" default:"1"`
	Int2  int           `desc:"an int"`
	Sub   SubConf1
	Sub2  SubConf2 `name:""`
}

func TestPrintEnvTemplate(t *testing.T) {
	type EnvConf struct {
		Str   string   `desc:"a string" default:"yes\"no"`
		Strs  []string `desc:"a string slice" default:"1,2,3"`
		StPtr *string  `desc:"a string pointer"`
		Int1  int      `desc:"an int" default:"1"`
		Int2  int      `desc:"an int"`
	}

	if os.Getenv("TEST_PASSTHROUGH") == "1" {
		co.Configure[EnvConf](&co.Options{
			Args:      []string{"--print_env_template", "--int_2", "88"},
			NoRecover: true,
			EnvPrefix: "FOO_",
		})
		os.Exit(0)
	}

	assert := assert.New(t)
	stdout, stderr := runExternal(t)

	assert.Equal("", stderr)
	assert.Equal(`# Generated with
# [--print_env_template --int_2 88]

# an int
FOO_INT_1="1"

# an int
FOO_INT_2="88"

# a string pointer
FOO_ST_PTR=""

# a string
FOO_STR="yes\"no"

# a string slice
FOO_STRS="[1,2,3]"

`, stdout, stdout)

}

func TestPrintYamlTemplate(t *testing.T) {
	if os.Getenv("TEST_PASSTHROUGH") == "1" {
		co.Configure[YamlConf](&co.Options{
			Args:              []string{"--print_yaml_template", "--st_ptr", "some-string"},
			NoRecover:         true,
			ShowInternalFlags: true,
		})
		os.Exit(0)
	}

	assert := assert.New(t)
	stdout, stderr := runExternal(t)

	assert.Equal("", stderr)
	assert.Equal(`# Generated with
# [--print_yaml_template --st_ptr some-string]

# a string
str: yes"no

# a string slice
strs:
    - "1"
    - "2"
    - "3"

# a string pointer
st_ptr: some-string

# an int
int_1: 1

# an int
int_2: 0

sub:

  # IP address on which to listen
  ip: 127.0.0.1

  lower:

    # Names and ages map
    ages:
        a: 1
        b: 2
        c: 3

# IP addresses
ips:
    - 127.0.0.1

`, stdout)

	tmpFl, _ := os.CreateTemp("", "test-*.yaml")
	defer os.Remove(tmpFl.Name())
	tmpFl.Write([]byte(stdout))
	tmpFl.Close()

	conf := co.Configure[YamlConf](&co.Options{
		Args:      []string{"--conf", tmpFl.Name()},
		NoRecover: true,
	})

	assert.Equal("some-string", *conf.StPtr)
	assert.Equal([]net.IP{net.ParseIP("127.0.0.1")}, *conf.Sub2.IPs)
}
