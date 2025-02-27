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
	fp "path/filepath"
	"strings"
	"testing"

	co "github.com/imoore76/configurature"
	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

/*
	Definitions for implementing the ImageFile custom type
*/
// type to be set as config struct field types
type ImageFile string

func (i *ImageFile) String() string {
	return (string)(*i)
}

func (i *ImageFile) Set(v string) error {
	// This will fail if the file does not exist or there is any other error
	// accessing the file
	if _, err := os.Stat(v); err != nil {
		return err
	}
	// This will fail if the file is not of the supported type
	switch ext := fp.Ext(strings.ToLower(v)); ext {
	case ".png", ".jpg", ".jpeg", ".gif":
		// ok
	default:
		return fmt.Errorf("file type \"%s\" not supported", ext)
	}
	*i = (ImageFile)(v)
	return nil
}

func (i *ImageFile) Type() string {
	return "imageFile"
}

func (i *ImageFile) Interface() any {
	return ImageFile(*i)
}

type MyConfig struct {
	StateFile string        `desc:"File in which to store lock state" short:"s"`
	Image     ImageFile     `desc:"Path to an image"`
	Images    []ImageFile   `desc:"Paths to images"`
	Conf      co.ConfigFile `desc:"Configuration file"`
}

func addImageFileTypes() {
	co.AddType[ImageFile]()
	co.AddType[[]ImageFile]()
}

func TestRegisterCustomType(t *testing.T) {
	addImageFileTypes()
}

func TestCustomType_Flag(t *testing.T) {
	addImageFileTypes()
	tmp, _ := os.CreateTemp("", "ldlm-test-*.png")
	tmp.Close()
	defer os.Remove(tmp.Name())

	conf := co.Configure[MyConfig](&co.Options{
		Args:  []string{"--image=" + tmp.Name()},
		Usage: func(_ *flag.FlagSet) {},
	})

	assert := assert.New(t)
	assert.Equal(ImageFile(tmp.Name()), conf.Image)
}

func TestCustomType_Error(t *testing.T) {
	addImageFileTypes()
	if os.Getenv("TEST_PASSTHROUGH") == "1" {

		co.Configure[MyConfig](&co.Options{
			Args:  []string{"--image", "./go.mod"},
			Usage: func(_ *flag.FlagSet) {},
		})
		os.Exit(0)
	}

	stdout, stderr := runExternal(t)
	assert := assert.New(t)
	assert.Equal(`invalid argument "./go.mod" for "--image" flag: file type `+
		`".mod" not supported`+"\n", stdout)
	assert.Equal(`invalid argument "./go.mod" for "--image" flag: file type `+
		`".mod" not supported`+"\n", stderr)

}

func TestCustomType_ErrorConfigFile(t *testing.T) {
	addImageFileTypes()
	if os.Getenv("TEST_PASSTHROUGH") == "1" {
		tmp, _ := os.CreateTemp("", "ldlm-test-*.yml")
		defer os.Remove(tmp.Name())
		tmp.Write([]byte("image: ./go.mod\n"))
		tmp.Close()

		co.Configure[MyConfig](&co.Options{
			Args:  []string{"--conf", tmp.Name()},
			Usage: func(_ *flag.FlagSet) {},
		})
		os.Exit(1)
	}

	stdout, stderr := runExternal(t)
	assert := assert.New(t)
	assert.Equal("", stdout)
	assert.Equal(`error parsing configuration: unable to set value for image: `+
		`file type ".mod" not supported`+"\n", stderr)

}

func TestCustomSliceType_Flag(t *testing.T) {
	addImageFileTypes()

	files := make([]string, 3)
	expected := make([]ImageFile, 3)
	for idx := range 3 {
		tmp, _ := os.CreateTemp("", "ldlm-test-*.png")
		tmp.Close()
		defer os.Remove(tmp.Name())

		files[idx] = tmp.Name()
		expected[idx] = ImageFile(tmp.Name())
	}

	conf := co.Configure[MyConfig](&co.Options{
		Args:  []string{"--images", strings.Join(files, ",")},
		Usage: func(_ *flag.FlagSet) {},
	})

	assert := assert.New(t)
	assert.Equal(expected, conf.Images)
}

func TestCustomSliceType_Error(t *testing.T) {
	addImageFileTypes()
	if os.Getenv("TEST_PASSTHROUGH") == "1" {

		co.Configure[MyConfig](&co.Options{
			Args:  []string{"--images", "./go.mod"},
			Usage: func(_ *flag.FlagSet) {},
		})
		os.Exit(0)
	}

	stdout, stderr := runExternal(t)
	assert := assert.New(t)
	assert.Equal(`invalid argument "./go.mod" for "--images" flag: file type `+
		`".mod" not supported`+"\n", stdout)
	assert.Equal(`invalid argument "./go.mod" for "--images" flag: file type `+
		`".mod" not supported`+"\n", stderr)

}

func TestCustomSliceType_ConfigFile(t *testing.T) {
	addImageFileTypes()
	files := make([]string, 3)
	expected := make([]ImageFile, 3)
	for idx := range 3 {
		tmp, _ := os.CreateTemp("", "ldlm-test-config-file*.png")
		tmp.Close()
		defer os.Remove(tmp.Name())

		files[idx] = tmp.Name()
		expected[idx] = ImageFile(tmp.Name())
	}

	tmp, _ := os.CreateTemp("", "ldlm-test-*.yaml")
	tmp.Write([]byte("images:\n  - " + strings.Join(files, "\n  - ") + "\n"))
	tmp.Close()
	defer os.Remove(tmp.Name())

	conf := co.Configure[MyConfig](&co.Options{
		Args:  []string{"--conf", tmp.Name()},
		Usage: func(_ *flag.FlagSet) {},
	})

	assert.Equal(t, expected, conf.Images)
}

func TestCustomSliceType_ErrorConfigFile(t *testing.T) {
	addImageFileTypes()
	if os.Getenv("TEST_PASSTHROUGH") == "1" {
		tmp, _ := os.CreateTemp("", "ldlm-test-*.yml")
		defer os.Remove(tmp.Name())
		tmp.Write([]byte("image: ./go.mod\n"))
		tmp.Close()

		co.Configure[MyConfig](&co.Options{
			Args:  []string{"--conf", tmp.Name()},
			Usage: func(_ *flag.FlagSet) {},
		})
		panic("Expected exit")
	}

	stdout, stderr := runExternal(t)
	assert := assert.New(t)
	assert.Equal("", stdout)
	assert.Equal(`error parsing configuration: unable to set value for image: `+
		`file type ".mod" not supported`+"\n", stderr)

}

func TestMapValueType(t *testing.T) {
	type Color string

	co.AddMapValueType("",
		[]string{"red", "blue", "green"},
		[]Color{"#ff0000", "#0000ff", "#00ff00"},
	)

	type CConf struct {
		Background Color `desc:"background color" default:"red"`
	}

	conf := co.Configure[CConf](&co.Options{
		Args:  []string{"--background", "blue"},
		Usage: func(_ *flag.FlagSet) {},
	})

	assert.Equal(t, conf.Background, Color("#0000ff"))
}

func TestMapValueType_NoValue(t *testing.T) {
	type Color string

	co.AddMapValueType("",
		[]string{"red", "blue", "green"},
		[]Color{"#ff0000", "#0000ff", "#00ff00"},
	)

	type CConf struct {
		Background Color `desc:"background color"`
	}

	conf := co.Configure[CConf](&co.Options{})

	assert.Equal(t, conf.Background, Color(""))
}

func TestMapValueType_BadValue(t *testing.T) {
	type Color string

	co.AddMapValueType("",
		[]string{"red", "blue", "green"},
		[]Color{"#ff0000", "#0000ff", "#00ff00"},
	)

	type CConf struct {
		Background Color `desc:"background color"`
	}

	if os.Getenv("TEST_PASSTHROUGH") == "1" {
		co.Configure[CConf](&co.Options{
			Args:  []string{"--background", "yellow"},
			Usage: func(_ *flag.FlagSet) {},
		})
		panic("Should have exited")
	}

	assert := assert.New(t)
	stdout, stderr := runExternal(t)

	assert.Equal(`invalid argument "yellow" for "--background" flag: `+
		`invalid Color: "yellow"`+"\n", stdout)
	assert.Equal(stdout, stderr)

}

func TestMapValueType_Usage(t *testing.T) {

	type Color string
	co.AddMapValueType("",
		[]string{"red", "blue", "green"},
		[]Color{"#ff0000", "#0000ff", "#00ff00"},
	)

	type MVT struct {
		Background Color `desc:"background color" default:"red"`
	}

	if os.Getenv("TEST_PASSTHROUGH") == "1" {
		co.Configure[MVT](&co.Options{
			Args: []string{"-h"},
		})
		panic("Should have exited")
	}

	assert := assert.New(t)
	stdout, stderr := runExternal(t)

	assert.Equal("", stderr)
	assert.True(strings.Contains(stdout, `--background Color   background color (red|blue|green) (default red)`), stdout)
}
