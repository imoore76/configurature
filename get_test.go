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

	"github.com/stretchr/testify/assert"

	co "github.com/imoore76/configurature"
)

func TestGet(t *testing.T) {
	assert := assert.New(t)

	co.Configure[TestNestedConfig](new(co.Options))

	sub, err := co.Get[OtherSubConfig]()

	assert.Nil(err)
	assert.NotNil(sub)
	assert.Equal("here", sub.SubFooString)
}

func TestGet_DisableCache(t *testing.T) {
	assert := assert.New(t)
	co.DisableGetTypeCache = true
	defer func() {
		co.DisableGetTypeCache = false
	}()
	co.Configure[TestNestedConfig](new(co.Options))

	sub, err := co.Get[OtherSubConfig]()

	assert.Nil(err)
	assert.NotNil(sub)
	assert.Equal("here", sub.SubFooString)
}

func TestGet_VeryNested(t *testing.T) {
	type T6 struct {
		SubFooString string `help:"Something" default:"t6there"`
	}
	type T5 struct {
		T5Str string `help:"t5" default:"t5"`
		Not   T6
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
		Something    T3
	}
	type T1 struct {
		T1Str string `help:"t1" default:"t1"`
		Sub   T2
	}

	assert := assert.New(t)

	co.Configure[T1](new(co.Options))

	sub, err := co.Get[T6]()

	assert.Nil(err)
	assert.NotNil(sub)
	assert.Equal("t6there", sub.SubFooString)

}

func TestGet_Anonymous(t *testing.T) {

	type T3 struct {
		T3Str string `help:"t3" default:"t3"`
	}
	type T2 struct {
		T2Str        string `help:"t2" default:"t2"`
		SubFooString string `help:"T2 Something" default:"t2there"`
		T3
	}
	type T1 struct {
		T1Str string `help:"t1" default:"t1"`
		T2
	}

	assert := assert.New(t)

	co.Configure[T1](new(co.Options))

	sub, err := co.Get[T3]()
	assert.Nil(err)
	assert.NotNil(sub)
	assert.Equal("t3", sub.T3Str)

	sub2, err := co.Get[T2]()
	assert.Nil(err)
	assert.NotNil(sub2)
	assert.Equal("t2there", sub2.SubFooString)

}

func TestGet_TopConfig(t *testing.T) {
	assert := assert.New(t)

	co.Configure[TestNestedConfig](new(co.Options))

	top, err := co.Get[TestNestedConfig]()

	assert.Nil(err)
	assert.NotNil(top)
}

func TestGet_NotFound(t *testing.T) {
	assert := assert.New(t)

	co.Configure[TestNestedConfig](new(co.Options))

	c, err := co.Get[TestConfigFileStruct]()

	assert.Nil(err)
	assert.Nil(c)
}
