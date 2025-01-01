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
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"

	co "github.com/imoore76/configurature"
)

type SubConfig struct {
	StateFile           string        `desc:"File in which to store lock state" short:"s"`
	DefaultLockTimeout  time.Duration `desc:"Lock timeout to use when loading locks from state file on startup" default:"10m" short:"d"`
	NoClearOnDisconnect bool          `desc:"Do not clear locks on client disconnect" default:"false" short:"c"`
	ReqInt              int           `desc:"Required int"`
	FooSeconds          uint          `desc:"Something" default:"10" short:"f"`
	FooInt              uint32        `desc:"Something" default:"100" short:"o"`
	FooInts             []uint        `desc:"Something list of ints" default:"100,200,30"`
}

type OtherSubConfig struct {
	SubFooString string `desc:"Something" default:"here"`
}

type TestConfig struct {
	SubConfig
	IgnoredField      string        `ignore:""`
	Bool              bool          `desc:"Bool thing" default:"false"`
	KeepaliveInterval time.Duration `desc:"Interval at which to send keepalive pings to client" default:"60s" short:"k"`
	KeepaliveTimeout  time.Duration `desc:"Wait this duration for the ping ack before assuming the connection is dead" default:"5s" short:"t"`
	LockGcInterval    time.Duration `desc:"Interval at which to garbage collect unused locks." default:"30m" short:"g"`
	OtherSubConfig
	LockGcMinIdle time.Duration `desc:"Minimum time a lock has to be idle (no unlocks or locks) before being considered for garbage collection" default:"5m" short:"m"`
	ListenAddress string        `desc:"Address (host:port) at which to listen" default:"localhost:3144" short:"l"`
	LogLevel      slog.Level    `desc:"Log level" default:"info" short:"v"`
}

type TestConfigFileStruct struct {
	CoolFile co.ConfigFile `desc:"Configuration file"`
	TestConfig
}

type TestNestedConfig struct {
	CoolFile   co.ConfigFile     `desc:"Configuration file"`
	SSlice     []string          `desc:"Slice of strings" default:"a,b,c"`
	MyMap      map[string]string `desc:"Map of strings"`
	NameAgeMap map[string]int    `desc:"Map of ages"`
	HiddenFlag string            `desc:"hidden flag" default:"hidden" hidden:"true"`
	MyEnum     string            `desc:"My enum" enum:"a,b,c" default:"a"`
	OS         OtherSubConfig
	Sub        SubConfig
}

func runExternal(t *testing.T) (out string, err string) {
	cmd := exec.Command(os.Args[0], "-test.run="+t.Name())
	cmd.Env = append(os.Environ(), "TEST_PASSTHROUGH=1")
	cmdout, cmderr := cmd.Output()
	if cmdout != nil {
		out = string(cmdout)
	} else {
		out = ""
	}
	if e, ok := cmderr.(*exec.ExitError); ok {
		err = string(e.Stderr)
	} else {
		err = ""
	}
	return out, err
}

func TestDefaults(t *testing.T) {
	assert := assert.New(t)
	c := co.Configure[TestConfig](&co.Options{NoRecover: true})

	assert.Equal("[100 200 30]", fmt.Sprintf("%v", c.FooInts), "FooInts should be [100, 200, 30]")
	assert.Equal("here", c.SubFooString, "SubFooString should be 'here'")
	assert.Equal(uint32(100), c.FooInt, "FooInt should be 100")
	assert.Equal(uint(10), c.FooSeconds, "FooSeconds should be 10")
	assert.Equal(slog.LevelInfo, c.LogLevel, "LogLevel should be info")
	assert.Equal(time.Duration(60)*time.Second, c.KeepaliveInterval, "KeepaliveInterval should be 60s")
	assert.Equal(time.Duration(5)*time.Second, c.KeepaliveTimeout, "KeepaliveTimeout should be 5s")
	assert.Equal(time.Duration(30)*time.Minute, c.LockGcInterval, "LockGcInterval should be 30m")
	assert.Equal(time.Duration(5)*time.Minute, c.LockGcMinIdle, "LockGcMinIdle should be 5m")
	assert.Equal("localhost:3144", c.ListenAddress, "ListenAddress should be localhost:3144")
}

func TestParseFlags(t *testing.T) {
	c := co.Configure[TestConfig](&co.Options{
		NoRecover: true,
		Args: []string{"cmd", "-v", "warn", "-k", "80m", "-t", "333m", "-g",
			"1s", "-m", "29h", "-l", "0.0.0.0:1", "-o", "7", "--sub_foo_string",
			"yes", "--foo_ints", "3,2,1", "--bool",
		},
	})

	assert := assert.New(t)
	assert.Equal("[3 2 1]", fmt.Sprintf("%v", c.FooInts), "FooInts should be [3, 2, 1]")
	assert.True(c.Bool, "Bool should be true")
	assert.Equal("yes", c.SubFooString, "SubFooString should be 'yes'")
	assert.Equal(slog.LevelWarn, c.LogLevel, "LogLevel should be warn")
	assert.Equal(time.Duration(80)*time.Minute, c.KeepaliveInterval, "KeepaliveInterval should be 80m")
	assert.Equal(time.Duration(333)*time.Minute, c.KeepaliveTimeout, "KeepaliveTimeout should be 333m")
	assert.Equal(time.Duration(1)*time.Second, c.LockGcInterval, "LockGcInterval should be 1s")
	assert.Equal(time.Duration(29)*time.Hour, c.LockGcMinIdle, "LockGcMinIdle should be 29h")
	assert.Equal("0.0.0.0:1", c.ListenAddress, "ListenAddress should be 0.0.0.0:1")
	assert.Equal(uint32(7), c.FooInt, "FooInt should be 7")
}

func TestEnvVars(t *testing.T) {
	origEnviron := os.Environ()
	defer func() {
		for _, e := range origEnviron {
			pair := strings.SplitN(e, "=", 2)
			os.Setenv(pair[0], pair[1])
		}
	}()
	// Clear environ
	for _, e := range origEnviron {
		pair := strings.SplitN(e, "=", 2)
		os.Setenv(pair[0], "")
	}

	envPrefix := "FOO_"
	setEnv := func(name string, val string) {
		os.Setenv(
			fmt.Sprintf("%s%s", envPrefix, strcase.ToScreamingSnake(name)),
			val,
		)
	}
	setEnv("KeepaliveInterval", "7h")
	setEnv("KeepaliveTimeout", "21h")
	setEnv("LockGcInterval", "88h")
	setEnv("LockGcMinIdle", "25h")
	setEnv("ListenAddress", "127.0.0.1:22")
	setEnv("LogLevel", "ERROR")
	setEnv("FooInts", "5,6,7,100")

	c := co.Configure[TestConfig](&co.Options{
		NoRecover: true,
		EnvPrefix: envPrefix,
	})

	assert := assert.New(t)
	assert.Equal(slog.LevelError, c.LogLevel, "LogLevel should be ERROR")
	assert.Equal(fmt.Sprintf("%v", c.FooInts), "[5 6 7 100]", "FooInts should be [5, 6, 7, 100]")
	assert.Equal(time.Duration(7)*time.Hour, c.KeepaliveInterval, "KeepaliveInterval should be 7h")
	assert.Equal(time.Duration(21)*time.Hour, c.KeepaliveTimeout, "KeepaliveTimeout should be 21h")
	assert.Equal(time.Duration(88)*time.Hour, c.LockGcInterval, "LockGcInterval should be 88h")
	assert.Equal(time.Duration(25)*time.Hour, c.LockGcMinIdle, "LockGcMinIdle should be 25h")
	assert.Equal("127.0.0.1:22", c.ListenAddress, "ListenAddress should be 127.0.0.1:22")
}

func TestFlagOverrideEnv(t *testing.T) {
	/*
		test all Type()s of config values
	*/
	envPrefix := "BAR_"
	origEnviron := os.Environ()
	defer func() {
		for _, e := range origEnviron {
			pair := strings.SplitN(e, "=", 2)
			os.Setenv(pair[0], pair[1])
		}
	}()
	// Clear environ
	for _, e := range origEnviron {
		pair := strings.SplitN(e, "=", 2)
		os.Setenv(pair[0], "")
	}

	setEnv := func(name string, val string) {
		os.Setenv(
			fmt.Sprintf("%s%s", envPrefix, strcase.ToScreamingSnake(name)),
			val,
		)
	}
	setEnv("LogLevel", "warn")
	setEnv("KeepaliveTimeout", "88m")
	setEnv("KeepaliveInterval", "7h")
	setEnv("ListenAddress", "127.0.0.1:22")
	setEnv("LogLevel", "ERROR")

	args := []string{"-k", "88m", "-v", "warn", "-l", "0.0.0.0:443"}
	c := co.Configure[TestConfig](&co.Options{
		EnvPrefix: envPrefix,
		Args:      args,
		NoRecover: true,
	})

	assert := assert.New(t)
	assert.Equal(slog.LevelWarn, c.LogLevel, "LogLevel should be WARN")
	assert.Equal(time.Duration(88)*time.Minute, c.KeepaliveInterval, "KeepaliveInterval should be 88m")
	assert.Equal("0.0.0.0:443", c.ListenAddress, "ListenAddress should be 0.0.0.0:443")
}

func TestConfigFile_FromEnv(t *testing.T) {
	assert := assert.New(t)

	tmp, _ := os.CreateTemp("", "ldlm-test-*.yml")
	defer os.Remove(tmp.Name())
	tmp.Write([]byte("foo_int: 4\nsub_foo_string: 'yes'\nkeepalive_timeout: 3m\nbool: true\n"))
	tmp.Close()

	os.Setenv("TEST_CONF_COOL_FILE", tmp.Name())
	defer os.Setenv("TEST_CONF_COOL_FILE", "")

	c := co.Configure[TestConfigFileStruct](&co.Options{
		NoRecover: true,
		EnvPrefix: "TEST_CONF_",
	})

	assert.Equal(uint32(4), c.FooInt, "FooInt should be 4")
	assert.Equal("yes", c.SubFooString, "SubFooString should be 'yes'")
	assert.Equal(time.Duration(3)*time.Minute, c.KeepaliveTimeout, "KeepaliveTimeout should be 3m")
	assert.Equal(true, c.Bool, "Bool should be true")
}

func TestConfigFile_Precedence(t *testing.T) {
	assert := assert.New(t)

	os.Setenv("TEST_CONF_FOO_INT", "7")
	os.Setenv("TEST_CONF_SUB_FOO_STRING", "asdf")

	tmp, _ := os.CreateTemp("", "ldlm-test-*.yml")
	defer os.Remove(tmp.Name())
	tmp.Write([]byte("foo_int: 4\nsub_foo_string: 'yes'\nkeepalive_timeout: 3m\nbool: false\n"))
	tmp.Close()

	c := co.Configure[TestConfigFileStruct](&co.Options{
		NoRecover: true,
		EnvPrefix: "TEST_CONF_",
		Args:      []string{"--cool_file", tmp.Name(), "--foo_int", "22"},
	})
	assert.Equal(uint32(22), c.FooInt, "FooInt should be 22")
	assert.Equal("asdf", c.SubFooString, "SubFooString should be 'asdf'")
	assert.Equal(time.Duration(3)*time.Minute, c.KeepaliveTimeout, "KeepaliveTimeout should be 3m")
	assert.Equal(false, c.Bool, "Bool should be false")

}

func TestBadEnvVar(t *testing.T) {
	if os.Getenv("TEST_PASSTHROUGH") == "1" {
		os.Setenv("TEST_CONF_FOO_INT", "asdf")
		co.Configure[TestConfigFileStruct](&co.Options{
			EnvPrefix: "TEST_CONF_",
		})
		os.Exit(0)
	}

	assert := assert.New(t)
	stdout, stderr := runExternal(t)

	assert.Equal("", stdout)
	assert.Equal("error parsing configuration: setFromEnv(): error setting value of field FooInt: "+
		"strconv.ParseUint: parsing \"asdf\": invalid syntax\n", stderr)
}

func TestBadFlagValue(t *testing.T) {
	if os.Getenv("TEST_PASSTHROUGH") == "1" {
		co.Configure[TestConfigFileStruct](&co.Options{
			Args:      []string{"-o", "asdf"},
			NoRecover: true,
		})
		os.Exit(0)
	}

	assert := assert.New(t)
	stdout, stderr := runExternal(t)

	assert.Equal("", stderr)
	assert.True(strings.HasPrefix(stdout, "Command usage:"))
}

func TestBadFlag(t *testing.T) {
	if os.Getenv("TEST_PASSTHROUGH") == "1" {
		co.Configure[TestConfigFileStruct](&co.Options{
			Args:      []string{"--thing_here", "asdf"},
			NoRecover: true,
		})
		os.Exit(0)
	}

	assert := assert.New(t)
	stdout, stderr := runExternal(t)

	assert.Equal("", stderr)
	assert.True(strings.HasPrefix(stdout, "Command usage:"))
}

func TestNested_Defaults(t *testing.T) {
	assert := assert.New(t)
	c := co.Configure[TestNestedConfig](&co.Options{
		NoRecover: true,
	})

	assert.Equal([]string{"a", "b", "c"}, c.SSlice, "SSlie should be a,b,c")
	assert.Equal("hidden", c.HiddenFlag, "HiddenFlag should be hidden")
	assert.Equal(time.Duration(10)*time.Minute, c.Sub.DefaultLockTimeout, "DefaultLockTimeout should be 10m")
	assert.Equal(uint(10), c.Sub.FooSeconds, "FooSeconds should be 10")
	assert.Equal(uint32(100), c.Sub.FooInt, "FooInt should be 100")
	assert.Equal("[100 200 30]", fmt.Sprintf("%v", c.Sub.FooInts), "FooInts should be [100 200 30]")
	assert.Equal("here", c.OS.SubFooString, "SubFooString should be here")
}

func TestNested_Flags(t *testing.T) {
	assert := assert.New(t)

	c := co.Configure[TestNestedConfig](&co.Options{
		NoRecover: true,
		Args: []string{"--s_slice", "af,asoidjf,d3",
			"--sub_default_lock_timeout", "30s", "--sub_foo_seconds", "24",
			"--sub_foo_int", "2", "--sub_foo_ints", "8,4,5",
			"--os_sub_foo_string", "there and nowhere"},
	})

	assert.Equal([]string{"af", "asoidjf", "d3"}, c.SSlice, "SSlie should be af,asoidjf,d3")
	assert.Equal(time.Duration(30)*time.Second, c.Sub.DefaultLockTimeout, "DefaultLockTimeout should be 30s")
	assert.Equal(uint(24), c.Sub.FooSeconds, "FooSeconds should be 24")
	assert.Equal(uint32(2), c.Sub.FooInt, "FooInt should be 2")
	assert.Equal("[8 4 5]", fmt.Sprintf("%v", c.Sub.FooInts), "FooInts should be [8 4 5]")
	assert.Equal("there and nowhere", c.OS.SubFooString, "SubFooString should be there and nowhere")
}

func TestNested_Env(t *testing.T) {

	os.Setenv("TESTAPP_S_SLICE", "af,asoidjf,d3")
	os.Setenv("TESTAPP_SUB_DEFAULT_LOCK_TIMEOUT", "30s")
	os.Setenv("TESTAPP_SUB_FOO_SECONDS", "24")
	os.Setenv("TESTAPP_SUB_FOO_INT", "2")
	os.Setenv("TESTAPP_SUB_FOO_INTS", "8,4,5")
	os.Setenv("TESTAPP_OS_SUB_FOO_STRING", "there and nowhere")

	c := co.Configure[TestNestedConfig](&co.Options{
		Args:      []string{},
		EnvPrefix: "TESTAPP_",
		NoRecover: true,
	})

	assert := assert.New(t)
	assert.Equal(time.Duration(30)*time.Second, c.Sub.DefaultLockTimeout, "DefaultLockTimeout should be 30s")
	assert.Equal(uint(24), c.Sub.FooSeconds, "FooSeconds should be 24")
	assert.Equal(uint32(2), c.Sub.FooInt, "FooInt should be 2")
	assert.Equal("[8 4 5]", fmt.Sprintf("%v", c.Sub.FooInts), "FooInts should be [8 4 5]")
	assert.Equal("there and nowhere", c.OS.SubFooString, "SubFooString should be there and nowhere")
}

func TestUsage(t *testing.T) {
	if os.Getenv("TEST_PASSTHROUGH") == "1" {
		co.Configure[TestNestedConfig](&co.Options{
			Args: []string{"-h"},
		})
		panic("Should have exited")
	}

	assert := assert.New(t)
	stdout, stderr := runExternal(t)

	assert.Equal("", stderr)
	assert.Equal(`Command usage:
      --cool_file configFile                Configuration file
  -h, --help                                show help and exit
      --my_enum string                      My enum (a|b|c) (default "a")
      --my_map stringToString               Map of strings (default [])
      --name_age_map stringToInt            Map of ages (default [])
      --os_sub_foo_string string            Something (default "here")
      --s_slice strings                     Slice of strings (default [a,b,c])
  -d, --sub_default_lock_timeout duration   Lock timeout to use when loading locks from state file on startup (default 10m0s)
  -o, --sub_foo_int uint32                  Something (default 100)
      --sub_foo_ints uints                  Something list of ints (default [100,200,30])
  -f, --sub_foo_seconds uint                Something (default 10)
  -c, --sub_no_clear_on_disconnect          Do not clear locks on client disconnect
      --sub_req_int int                     Required int
  -s, --sub_state_file string               File in which to store lock state

`, stdout)
}

func TestUsage_Custom(t *testing.T) {
	if os.Getenv("TEST_PASSTHROUGH") == "1" {
		co.Configure[TestConfigFileStruct](&co.Options{
			Args: []string{"-h"},
			Usage: func(f *pflag.FlagSet) {
				fmt.Println("oh no")
			},
			NoRecover: true,
		})
		os.Exit(0)
	}

	assert := assert.New(t)
	stdout, stderr := runExternal(t)

	assert.Equal("", stderr)
	assert.Equal("oh no\n", stdout)
}

func TestUsage_ShowInternalFlags(t *testing.T) {
	if os.Getenv("TEST_PASSTHROUGH") == "1" {
		co.Configure[TestNestedConfig](&co.Options{
			Args:              []string{"-h"},
			ShowInternalFlags: true,
		})
		panic("Should have exited")
	}

	assert := assert.New(t)
	stdout, stderr := runExternal(t)

	assert.Equal("", stderr)
	assert.Equal(`Command usage:
      --cool_file configFile                Configuration file
  -h, --help                                show help and exit
      --my_enum string                      My enum (a|b|c) (default "a")
      --my_map stringToString               Map of strings (default [])
      --name_age_map stringToInt            Map of ages (default [])
      --os_sub_foo_string string            Something (default "here")
      --print_env_template                  Print example environment variables and exit
      --print_yaml_template                 Print example YAML config file and exit
      --s_slice strings                     Slice of strings (default [a,b,c])
  -d, --sub_default_lock_timeout duration   Lock timeout to use when loading locks from state file on startup (default 10m0s)
  -o, --sub_foo_int uint32                  Something (default 100)
      --sub_foo_ints uints                  Something list of ints (default [100,200,30])
  -f, --sub_foo_seconds uint                Something (default 10)
  -c, --sub_no_clear_on_disconnect          Do not clear locks on client disconnect
      --sub_req_int int                     Required int
  -s, --sub_state_file string               File in which to store lock state

`, stdout)
}

func TestSubConfig_EmptyName(t *testing.T) {
	type TConf struct {
		Bar OtherSubConfig `name:""`
	}

	var conf *TConf
	var err error = nil
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(r.(string))
			}
		}()
		conf = co.Configure[TConf](&co.Options{
			NoRecover: true,
			Args:      []string{"--sub_foo_string=aaaaasdfjaoisjdkfoiaj"},
			Usage: func(f *pflag.FlagSet) {
				f.PrintDefaults()
			},
		})
	}()
	assert.Nil(t, err)
	assert.Equal(t, "aaaaasdfjaoisjdkfoiaj", conf.Bar.SubFooString)

}

func TestSubConfig_EmptyNameEnv(t *testing.T) {
	type TConf struct {
		Bar OtherSubConfig `name:""`
	}

	var conf *TConf
	var err error = nil
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(r.(string))
			}
		}()
		orig := os.Getenv("SUB_FOO_STRING")
		defer os.Setenv("SUB_FOO_STRING", orig)
		os.Setenv("SUB_FOO_STRING", "aaaaasdfjaoisjdkfoiaj")
		conf = co.Configure[TConf](&co.Options{
			NoRecover: true,
			Args:      []string{},
			Usage: func(f *pflag.FlagSet) {
				f.PrintDefaults()
			},
		})
	}()
	assert.Nil(t, err)
	assert.Equal(t, "here", conf.Bar.SubFooString)

}

func TestSubConfig_GivenName(t *testing.T) {
	type TConf struct {
		Bar OtherSubConfig `name:"nope"`
	}

	var conf *TConf
	var err error = nil
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(r.(string))
			}
		}()
		conf = co.Configure[TConf](&co.Options{
			NoRecover: true,
			Args:      []string{"--nope_sub_foo_string=jenny8675309"},
			Usage: func(f *pflag.FlagSet) {
				f.PrintDefaults()
			},
		})
	}()
	assert.Nil(t, err)
	assert.Equal(t, "jenny8675309", conf.Bar.SubFooString)

}

func TestSubConfig_GivenNameEnv(t *testing.T) {
	type TConf struct {
		Bar OtherSubConfig `name:"nope"`
	}

	var conf *TConf
	var err error = nil
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(r.(string))
			}
		}()
		orig := os.Getenv("NOPE_SUB_FOO_STRING")
		defer os.Setenv("NOPE_SUB_FOO_STRING", orig)
		os.Setenv("NOPE_SUB_FOO_STRING", "jenny8675309")

		conf = co.Configure[TConf](&co.Options{
			NoRecover: true,
			Args:      []string{"--nope_sub_foo_string=jenny8675309"},
			Usage: func(f *pflag.FlagSet) {
				f.PrintDefaults()
			},
		})
	}()
	assert.Nil(t, err)
	assert.Equal(t, "jenny8675309", conf.Bar.SubFooString)

}

func TestNilPtrs_True(t *testing.T) {
	type TConf struct {
		PString     *string     `desc:"Pointer to string"`
		PLogLevel   *slog.Level `desc:"Pointer to log level" default:"debug"`
		PInt        *int        `desc:"Pointer to int"`
		PInts       *[]int      `desc:"Pointer to int slice"`
		PIntsDef    *[]int      `desc:"Pointer to int slice" default:"1,3,4"`
		PStrings    *[]string   `desc:"Pointer to string slice"`
		PStringsDef *[]string   `desc:"Pointer to string slice" default:"a,b,c"`
	}

	var conf *TConf
	var err error = nil
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(r.(string))
			}
		}()
		conf = co.Configure[TConf](&co.Options{
			NoRecover: true,
			NilPtrs:   true,
			Args:      []string{"--p_string=asdf12"},
			Usage: func(f *pflag.FlagSet) {
				f.PrintDefaults()
			},
		})
	}()
	assert.Nil(t, err)
	assert.Equal(t, "asdf12", *conf.PString)
	assert.Equal(t, slog.LevelDebug, *conf.PLogLevel)
	assert.Nil(t, conf.PInt)
	assert.Nil(t, conf.PInts)
	assert.Equal(t, []int{1, 3, 4}, *conf.PIntsDef)
	assert.Nil(t, conf.PStrings)
	assert.Equal(t, []string{"a", "b", "c"}, *conf.PStringsDef)

}

func TestNilPtrs_False(t *testing.T) {
	type TConf struct {
		PString     *string     `desc:"Pointer to string"`
		PLogLevel   *slog.Level `desc:"Pointer to log level" default:"debug"`
		PInt        *int        `desc:"Pointer to int"`
		PString2    *string     `desc:"Pointer to another string"`
		PInts       *[]int      `desc:"Pointer to int slice"`
		PIntsDef    *[]int      `desc:"Pointer to int slice" default:"1,3,4"`
		PStrings    *[]string   `desc:"Pointer to string slice"`
		PStringsDef *[]string   `desc:"Pointer to string slice" default:"a,b,c"`
	}

	var conf *TConf
	var err error = nil
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(r.(string))
			}
		}()
		conf = co.Configure[TConf](&co.Options{
			NoRecover: true,
			NilPtrs:   false,
			Args:      []string{"--p_string=asdf12"},
			Usage: func(f *pflag.FlagSet) {
				f.PrintDefaults()
			},
		})
	}()
	assert.Nil(t, err)
	assert.Equal(t, "asdf12", *conf.PString)
	assert.Equal(t, "", *conf.PString2)
	assert.Equal(t, slog.LevelDebug, *conf.PLogLevel)
	assert.Equal(t, 0, *conf.PInt)
	assert.Equal(t, []int{}, *conf.PInts)
	assert.Equal(t, []int{1, 3, 4}, *conf.PIntsDef)
	assert.Equal(t, []string{}, *conf.PStrings)
	assert.Equal(t, []string{"a", "b", "c"}, *conf.PStringsDef)

}
