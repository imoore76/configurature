# Configurature

<p align="center">
<img src="logo.png" height="384" width="384" alt="configurature logo" />
</p>

Configurature is a Go library that provides declarative app configuration using structs.
Configuration values can be specified (in value precedence order) on the command line, using environment variables, and/or in a config file (yaml or json).

Configuration structs can be composed in a way that your application's entry points do not need to be aware of the structure of other packages' configurations in order to initialize them.

---

Table of Contents

<!-- TOC -->

- [Usage](#usage)
  * [Basic](#basic)
  * [Using .env Files](#using-env-files)
  * [Using Validators](#using-validators)
  * [Options](#options)
    + [EnvPrefix](#envprefix)
    + [Args](#args)
    + [NilPtrs](#nilptrs)
    + [NoRecover](#norecover)
    + [Usage](#usage-1)
    + [ShowInternalFlags](#showinternalflags)
  * [Combining Package Configurations](#combining-package-configurations)
    + [Nested Config](#nested-config)
    + [Flat Config](#flat-config)
    + [Mixed](#mixed)
    + [Get](#get)
- [Tags](#tags)
- [Validators](#validators)
  * [Compound Types](#compound-types)
  * [Network](#network)
  * [Strings](#strings)
  * [Format](#format)
  * [Comparisons](#comparisons)
  * [Other](#other)
- [Supported Types](#supported-types)
  * [Custom Types](#custom-types)
    + [Map Vaule Types](#map-value-types)
- [License](#license)
- [Contributing](#contributing)
- [Disclaimer](#disclaimer)

<!-- TOC end -->

<!-- TOC --><a name="usage"></a>

## Usage

Basic usage consists of defining your configuration structs and running `configurature.Configure()`. This section is meant to be read serially where each section builds on the previous one. Read on.

<!-- TOC --><a name="basic"></a>

### Basic

Declare your configuration and run Configure.

```go
package main

import (
    "fmt"
    "net"
    "os"

    co "github.com/imoore76/configurature"
)

type AppConfig struct {
    ListenIP   net.IP `desc:"IP address on which to listen" default:"127.0.0.1"`
    ListenPort uint   `desc:"port on which to listen" default:"8080"`
}

func main() {

    conf := co.Configure[AppConfig](&co.Options{
        EnvPrefix: "APP_",
        Args:      os.Args[1:],
    })

    fmt.Printf("IP: %s\n", conf.ListenIP)
    fmt.Printf("Port: %d\n", conf.ListenPort)
}
```

The tags `desc` add descriptions to the fields and are required. `default` specifies a default value.

Running this app with `--help` displays the app usage:

```
user@host $ myapp --help
Command usage:
  -h, --help               show help and exit
      --listen_ip ip       IP address on which to listen (default 127.0.0.1)
      --listen_port uint   port on which to listen (default 8080)
```

These flags can be specified on the command line.

```
user@host ~$ myapp --listen_ip 0.0.0.0
IP: 0.0.0.0
Port: 8080
```

You can also use environment variables in the form of uppercase arguments prefixed with the `EnvPrefix` option.

```
user@host ~$ APP_LISTEN_PORT=443 myapp --listen_ip 0.0.0.0
IP: 0.0.0.0
Port: 443
```

A config file can be added by adding a special `ConfigFile` field to the configuration struct. `ConfigFile` is part of the `configurature` package.

```go
type AppConfig struct {
    ListenIP   net.IP        `desc:"IP address on which to listen" default:"127.0.0.1"`
    ListenPort uint          `desc:"port on which to listen" default:"8080"`
    Conf       co.ConfigFile `desc:"configuration file" short:"c"`
}
```

The example above also adds a `short` tag to specify a short version of the option.

```shell
user@host ~$ myapp -h
Command usage:
  -c, --conf configFile    configuration file
  -h, --help               show help and exit
      --listen_ip ip       IP address on which to listen (default 127.0.0.1)
      --listen_port uint   port on which to listen (default 8080)
```

Supported configuration file formats are yaml and json (determined by file extension). Adding the configuration options to a `conf.yaml` file looks like

```yaml
# conf.yaml
listen_ip: 0.0.0.0
listen_port: 80
```

```shell
user@host ~$ myapp -c conf.yaml
IP: 0.0.0.0
Port: 80
```

Resolution order of values is command line, environment variable, and finally configuration file.

### Generating Config Files

You can use the internal hidden flag `--print_yaml_template` to generate a YAML template file. Specifying values along with using the flag uses those values for the configuration file output.

```shell
user@host ~$ myapp --listen_ip 0.0.0.0 --print_yaml_template
# Generated from
# [--listen_ip 0.0.0.0 --print_yaml_template]

# IP address on which to listen (default 127.0.0.1)
listen_ip: 0.0.0.0

# port on which to listen (default 8080)
listen_port: 8080
```

Copy and paste, or redirect its output to a `.yaml` file, and edit as needed.

<!-- TOC --><a name="using-env-files"></a>

### Using .env files

Configurature itself does not provide anything that parses `.env` files since there are many other tools that have filled this space. Here are a couple of my favorites:

* godotenv - https://github.com/joho/godotenv
* direnv - https://direnv.net/

Since Configurature automatically uses environment variables, there is no further action needed after loading environment variables from a `.env` file. To create a template `.env` file, you can run your app with a `--print_env_template` flag. This flag is hidden so won't appear in your app's Usage() text when using `--help`.

```shell
user@host ~$ myapp --print_env_template

# IP address on which to listen (default 127.0.0.1)
APP_LISTEN_IP="127.0.0.1"

# port on which to listen (default 8080)
APP_LISTEN_PORT="8080"
```

Copy and paste, or redirect its output to a `.env` file, then edit as needed.

<!-- TOC --><a name="using-validators"></a>

### Using Validators

Validators provide validation of configuration input and are specified in the `validators` tag. The `ListenPort` can be changed so that it must fall within the server application port range of 1024 - 49151.

```go
type AppConfig struct {
    ListenIP   net.IP        `desc:"IP address on which to listen" default:"127.0.0.1"`
    ListenPort uint          `desc:"port on which to listen" default:"8080" validate:"gt=1024,lt=49151"`
    Conf       co.ConfigFile `desc:"configuration file" short:"c"`
}
```

Re-running the app with a `listen_port` of `80` results in

```
user@host ~$ myapp --listen_port 80
error parsing configuration: validation failed; listen_port must be greater than 1024
```

See the [Validators](#validators) section for more info.

<!-- TOC --><a name="options"></a>

### Options

`Configure[]()` expects an `Options` pointer as input.

```go
// Config options
type Options struct {
	EnvPrefix         string              // Prefix for environment variables
	Args              []string            // Arguments to parse
	NilPtrs           bool                // Leave pointers set to nil if values aren't specified
	Usage             func(*flag.FlagSet) // Usage function called when configuration is incorrect or for --help
	NoRecover         bool                // Don't recover from panic
	ShowInternalFlags bool                // Show hidden internal flags
}
```

<!-- TOC --><a name="envprefix"></a>

#### EnvPrefix

The prefix to use when checking for configuration values specified as environment variables.

<!-- TOC --><a name="args"></a>

#### Args

The string of arguments to parse. Typically this would be set to `os.Args[1:]`. The first element is the name of the command, so is not included.

<!-- TOC --><a name="nilptrs"></a>

#### NilPtrs

When using config field pointers, leave them set to `nil` rather than a zero value if no value is specified and no default is set.

```go
type Config struct {
    MaxConns   *int `desc:"Max number of connections"`
    ListenPort *int `desc:"Port on which to listen" default:"8080"`
}

conf := co.Configure[Config](&co.Opts{
    NilPtrs: true,
})

if conf.MaxConns != nil {
    fmt.Printf("MaxConns is %d\n", *conf.MaxConns)
}
fmt.Printf("ListenPort is %d\n", *conf.ListenPort)
```

This would result in `ListenPort is 8080` since no value was specified and no default was provided for MaxConns. With the default behavior, MaxConns would be a pointer to a zero value (`0` for ints).

<!-- TOC --><a name="norecover"></a>

#### NoRecover

The default behavior is to handle a `panic()` that occurs during parsing, print its message to `os.Stderr` and exit.

```go
    if r := recover(); r != nil {
        fmt.Fprintf(os.Stderr, "error parsing configuration: %s\n", r)
        os.Exit(1)
    }
```

You can disable this behavior by setting `NoRecover` to true. You may then handle panics in your app however you'd like.

```go
var err error = nil
var conf *AppConfig
func() {
    defer func() {
        if r := recover(); r != nil {
            err = errors.New(r.(string))
        }
    }()
    conf = co.Configure[AppConfig](&co.Options{
        EnvPrefix: "APP_",
        Args:      os.Args[1:],
        NoRecover: true,
    })

}()

if err != nil {
    // custom handling
}
```

#### ShowInternalFlags

Show internal flags `--print_env_template` and `--print_yaml_template` in Usage() text.

<!-- TOC --><a name="usage-1"></a>

#### Usage

Create and specify your own handler for command usage help. This occurs when incorrect configuration flags are supplied or for `--help`.
The default `Usage` function is 

```go
func(f *pflag.FlagSet) {
    fmt.Println("Command usage:")
    fmt.Println(f.FlagUsages())
    os.Exit(0)
}
```

<!-- TOC --><a name="combining-package-configurations"></a>

### Combining Package Configurations

A feature of Configurature is that you can combine configurations from multiple packages of your application into a single configuration and initialize it from a single location (e.g. `main()`).
Given the following project layout

```
├── cmd
│   └── main.go
├── db
│   └── db.go
├── log
│   └── log.go
└── server
    ├── ipc
    │   └── ipc.go
    ├── rpc
    │   └── rpc.go
    └── server.go
```

Typically each of these packages has its own configuration requirements. The server package is in charge of passing the `ipcSocketFile` to the `ipc` package instantiation and the `rpcTimeout` to the `rpc` package. The `main` package needs to pass config to `server`. They can be passed around as distinct function arguments, which may be some variation of:

```go
s := server.New(listenIp, listenPort, ipcSocketFile, rpcTimeout)
```

Configurature provides a mechanism to avoid this scenario. So that adding a configuration option to the `ipc` package is as easy as adding a single field to its configuration struct. Neither the `server` or `main` packages need to be made aware of the new configuration or any change in configuration at all.

<!-- TOC --><a name="nested-config"></a>

#### Nested Config

You can nest configurations in configuration structs so that they are distinct fields in a parent struct.

```go
// main.go
type Config struct {
    ConfigFile co.ConfigFile `desc:"configuration file" short:"c"`
    Logging    log.Config
    Server     server.Config
}
```

```go
// server.go
type Config struct {
    ListenIP   net.IP `desc:"IP address on which to listen" default:"127.0.0.1"`
    ListenPort uint   `desc:"port on which to listen" default:"8080"`
    IPC        ipc.Config
    RPC        rpc.Config
}
```

Now `server` and `log` are initialized in `main()` with:

```go
    conf := co.Configure[Config](&co.Options{
        EnvPrefix: "APP_",
        Args:      os.Args[1:],
    })

    log.SetupLogging(&conf.Logging)
    s := server.New(&conf.Server)
    s.Run()
```

`ipc` and `rpc` can be similarly initialized from the `server` package without needing to be aware of their configuration details. Changes in configuration do not result in changes to function or method calls.

Nested configuration does result in nested value specification. E.g. `SocketFile` in the `ipc` configuration becomes a `--server_ipc_socket_file` flag, `APP_SERVER_IPC_SOCKET_FILE` environment variable, and in a  configuration file:

```yaml
# conf.yaml
server:
  ipc:
    socket_file: /tmp/server-ipc.sock
```

You can name sub configs and even give them empty names.

```go
// main.go
type Config struct {
    ConfigFile co.ConfigFile `desc:"configuration file" short:"c"`
    Logging    log.Config
    Server     server.Config `name:""`
}
```

```go
// server.go
type Config struct {
    ListenIP   net.IP     `desc:"IP address on which to listen" default:"127.0.0.1"`
    ListenPort uint       `desc:"port on which to listen" default:"8080"`
    IPC        ipc.Config `name:"other"`
    RPC        rpc.Config
}
```

`ListenIP` is now specified using `--listen_ip` instead of `--server_listen_ip` because its name is empty. IPC configuration is prefixed with `other_`. E.g. `--other_socket_file` from the command line. Naming applies to environment variables and config file structure as well.

<!-- TOC --><a name="flat-config"></a>

#### Flat Config

You can also include other config structs as anonymous fields.

```go
// main.go
type Config struct {
    ConfigFile co.ConfigFile `desc:"configuration file" short:"c"`
    log.LogConfig
    server.ServerConfig
}
```

```go
// server.go
type ServerConfig struct {
    ListenIP   net.IP `desc:"IP address on which to listen" default:"127.0.0.1"`
    ListenPort uint   `desc:"port on which to listen" default:"8080"`
    ipc.IPCConfig
    rpc.RPCConfig
}
```

Now `server` and `log` are initialized in `main()` with:

```go
    conf := co.Configure[Config](&co.Options{
        EnvPrefix: "APP_",
        Args:      os.Args[1:],
    })

    log.SetupLogging(&conf.LogConfig)
    s := server.New(&conf.ServerConfig)
    s.Run()
```

A downside of this is that the names config structs from each package must be unique and there can not be duplicate field names between structs. This is usually fine for small projects. A hybrid approach can also be used where `server` is a flat config including `ipc` and `rpc` configs anonymously and `main`s config contains concrete fields that hold configurations for other packages.

Configuration structs included anonymously result in flat value specification. E.g. `SocketFile` in the `ipc` configuration becomes a `--socket_file` flag, `APP_SOCKET_FILE` environment variable. You may want to rename the field name in this case to `IpcSocketFile` or just let the field's `desc` provide context for what "socket file" refers to.

<!-- TOC --><a name="mixed"></a>

#### Mixed

You are free to mix and match in a way that makes sense for your project. `Config()` will panic() if there are duplicate field names or short flag names instead of quietly resulting in unintended configuration.

<!-- TOC --><a name="get"></a>

#### Get

You can also use `Get[T]()` from anywhere in your app as long as `Configure[T]()` has been called.

```go
// main.go
type Config struct {
    BuriedComponentConfig bc.Config
    StoreConfig           store.Config
}

func main() {
    conf := co.Configure[Config](&co.Options{
        EnvPrefix: "APP_",
        Args:      os.Args[1:],
    })
    // ...
}
```

Then anywhere else in our code, you can call `Get[T]()` where T is the type of config you want to retrieve from the top-level configuration.

```go
// buried_component.go

type Config struct {
    MyInt int    `desc:"integer config item"`
    MyStr string `desc:"string config item"`
}

// buried_component needs its Config struct
func New() {
    if conf, err := co.Get[Config](); err != nil {
        // handle err
    } else {
        // Initialize a new BuriedComponent with conf
    }
}

// components need store Config to initialize their own stores
func doSomethingWithNewStore() (err error) {
    sConf, err := co.Get[store.Config]()
    if err != nil {
        return fmt.Errorf("error getting store config: %w", err)
    }

    store := store.New(sConf)

    // do something with store
}
```

<!-- TOC --><a name="tags"></a>

## Tags

The following tags are used by Configurature:

| Tag                | Description                                                                                                                      |
|:------------------:|:-------------------------------------------------------------------------------------------------------------------------------- |
| `desc:"..."`       | Provides a description for the field shown in the `Usage` message.                                                               |
| `default:"..."`    | Specifies the default value for the field                                                                                        |
| `short:"..."`      | Specifies the short version of the flag                                                                                          |
| `name:"..."`       | Specifies an alternate name for the field instead of using the struct field name converted to snake case                         |
| `hidden:""`        | Hides the field from the "Usage" help displayed. It can still be specified on the cli, environment variable, or config file.     |
| `ignore:""`        | Makes Configurature completely ignore the struct field                                                                           |
| `enum:"x,y,z"`     | Only the values `x`, `y`, or `z` are valid for this field. This will automatically add the values to the help text of the field. |
| `validate:"..."` | Specifies [validators](#validators) for this field separated by `,`                                                              |

<!-- TOC --><a name="validators"></a>

## Validators

Validators are specified in a field tag as `validate:"validator1,validator2"`. Some validators require an option and may be specified as `validators:"validator1=foo,validator2=bar"`. This functionality is provided by the excellent [github.com/go-playground/validator](http://github.com/go-playground/validator) package. Fore more information, be sure to read its documentation at https://pkg.go.dev/github.com/go-playground/validator/v10#readme-usage-and-documentation.

### Compound Types

Some validators can act on compound and scalar values. For instance, given a field of type `[]string` the `lt` validator can specify the maximum number of elements in the slice or the maximum length of strings in the `[]string` slice. By default, validators will validate the field value - so if it is a slice, `lt` will apply to the number of items in the slice. To validate items within a compound value, use `dive`.

```go
type Config struct {
    // No more than 5 names can be specified. Each name must be at lest 3 chars long.
    Names []string `desc:"names" validate:"lt=6,dive,gte=3"`
}
```

To validate map keys, you can use `keys` to specify that the following validators are applied to the map's keys and `endkeys` to specify that the remaining validators are to be applied to the map's values.

```go
type Config struct {
    // Ages must be specified. Names in the map must be > 3 characters. Ages in the map must be >= 18
    Ages map[string]int `desc:"Names and ages map" validate="required,keys,gt=3,endkeys,dive,gte=18"`
}
```

### Network
| Validator | Description |
| :---: | :-- |
| `cidr` | Classless Inter-Domain Routing CIDR | 
| `cidrv4` | Classless Inter-Domain Routing CIDRv4 | 
| `cidrv6` | Classless Inter-Domain Routing CIDRv6 | 
| `datauri` | Data URL | 
| `fqdn` | Full Qualified Domain Name (FQDN) | 
| `hostname` | Hostname RFC 952 | 
| `hostname_port` | `host:port` format | 
| `ip` | Internet Protocol Address IP | 
| `ipv4` | Internet Protocol Address IPv4 | 
| `ipv6` | Internet Protocol Address IPv6 | 
| `mac` | Media Access Control Address MAC | 
| `unix_addr` | Unix domain socket end point Address (path to socke tfile) | 
| `uri` | URI String | 
| `url` | URL String | 
| `http_url` | HTTP URL String (`http` and `https` will both validate)| 
| `url_encoded` | URL Encoded | 

### Strings
| Validator | Description | Option |
| :---: | :-- | :-- |
| `alpha` | Alpha Only | |
| `alphanum` | Alphanumeric |  |
| `alphanumunicode` | Alphanumeric Unicode |  |
| `alphaunicode` | Alpha Unicode |  |
| `ascii` | ASCII |  |
| `boolean` | Boolean. Can be any one of `1`, `t`, `T`, `TRUE`, `true`, `True`, `0`, `f`, `F`, `FALSE`, `false`, `False`|  |
| `contains=...` | Contains | the string that must be contained in the value |
| `containsany=...` | Contains Any | the string with the characters of which at least one must be in the value |
| `endsnotwith=...` | Ends Not With | the string that the value must not end with |
| `endswith=...` | Ends With | the string that the value not end with |
| `excludes=...` | Excludes | the string that must not be contained in the value |
| `excludesall=...` | Excludes All | the string with the characters of which none may be in the value |
| `lowercase` | Lowercase |  |
| `multibyte` | Multi-Byte Characters |  |
| `not_blank`   | The value must contain non-whitespace characters |  |
| `number` | Number - e.g. `8`, `4.2` `-20` |  |
| `numeric` | Must be a number or a string that can be parsed as a number |  |
| `printascii` | Printable ASCII |  |
| `regex=...`   | The value must match the supplied regex  | A valid regular expression.<br />E.g. `^\d{3}-\d{3}-\d{4}$` |
| `startsnotwith=...` | Starts Not With |  the string with which the value must not start|
| `startswith=...` | Starts With |  the string with which the value must start |
| `uppercase` | Uppercase |  |

### Format
| Validator | Description |
| :---: | :-- |
| `base64` | Base64 String | 
| `btc_addr` | Bitcoin Address | 
| `btc_addr_bech32` | Bitcoin Bech32 Address (segwit) | 
| `mongodb` | MongoDB ObjectID | 
| `cron` | Cron | 
| `spicedb` | SpiceDb ObjectID/Permission/Type | 
| `datetime` | Datetime | 
| `e164` | e164 formatted phone number | 
| `email` | E-mail String | 
| `eth_addr` | Ethereum Address | 
| `hexadecimal` | Hexadecimal String | 
| `hexcolor` | Hexcolor String | 
| `hsl` | HSL String | 
| `hsla` | HSLA String | 
| `html` | HTML Tags | 
| `html_encoded` | HTML Encoded | 
| `isbn` | International Standard Book Number | 
| `isbn10` | International Standard Book Number 10 | 
| `isbn13` | International Standard Book Number 13 | 
| `issn` | International Standard Serial Number | 
| `iso3166_1_alpha2` | Two-letter country code (ISO 3166-1 alpha-2) | 
| `iso3166_1_alpha3` | Three-letter country code (ISO 3166-1 alpha-3) | 
| `iso3166_1_alpha_numeric` | Numeric country code (ISO 3166-1 numeric) | 
| `iso3166_2` | Country subdivision code (ISO 3166-2) | 
| `iso4217` | Currency code (ISO 4217) | 
| `json` | JSON | 
| `latitude` | Latitude | 
| `longitude` | Longitude | 
| `postcode_iso3166_alpha2` | Postcode | 
| `postcode_iso3166_alpha2_field` | Postcode | 
| `rgb` | RGB String | 
| `rgba` | RGBA String | 
| `ssn` | Social Security Number SSN | 
| `timezone` | Timezone | 
| `uuid` | Universally Unique Identifier UUID | 
| `uuid3` | Universally Unique Identifier UUID v3 | 
| `uuid4` | Universally Unique Identifier UUID v4 | 
| `uuid5` | Universally Unique Identifier UUID v5 | 
| `md5` | MD5 hash | 
| `sha256` | SHA256 hash | 
| `sha384` | SHA384 hash | 
| `sha512` | SHA512 hash | 
| `ripemd128` | RIPEMD-128 hash | 
| `ripemd128` | RIPEMD-160 hash | 
| `tiger128` | TIGER128 hash | 
| `tiger160` | TIGER160 hash | 
| `tiger192` | TIGER192 hash | 
| `semver` | Semantic Versioning 2.0.0 | 
| `ulid` | Universally Unique Lexicographically Sortable Identifier ULID | 
| `cve` | Common Vulnerabilities and Exposures Identifier (CVE id) | 

### Comparisons
| Validator | Description | Option |
| :---: | :-- | :-- |
| `gt=...` | Greater than | depends on type - see note | 
| `gte=...` | Greater than or equal | depends on type - see note | 
| `lt=...` | Less Than | depends on type - see note | 
| `lte=...` | Less Than or Equal |  depends on type - see note| 
| `ne=...` | Not Equal | depends on type - see note |


**Note** For numeric values, the comparison is treated as numeric. For string values, the length of the string is used for comparison. For slices, maps, and arrays, the number of items is used for comparison.

### Other
| Validator | Description | Option |
| :---: | :-- | :-- |
| `dir` | Existing Directory | |
| `dirpath` | Directory Path | |
| `file` | Existing File | |
| `filepath` | File Path | |
| `image` | File path to a valid image | |
| `len=...` | Exact length | Exact length of string, slice, array, or map. | 
| `max=...` | Maximum length | Max length of string, slice, array, or map. | 
| `min=...` | Minimum length | Min length of string, slice, array, or map. |  
| `required` | A value must be supplied | | 


<!-- TOC --><a name="supported-types"></a>

## Supported Types

The following field types are supported

* `[]bool`
* `[]float32`
* `[]float64`
* `[]int32`
* `[]int64`
* `[]int`
* `[]net.IP`
* `[]string`
* `[]time.Duration`
* `[]uint8`
* `[]uint`
* `bool`
* `configurature.ConfigFile`
* `float32`
* `float64`
* `int16`
* `int32`
* `int64`
* `int8`
* `int`
* `map[string]int64`
* `map[string]int`
* `map[string]string`
* `net.IPMask`
* `net.IPNet`
* `net.IP`
* `slog.Level`
* `string`
* `time.Duration`
* `uint16`
* `uint32`
* `uint64`
* `uint8`
* `uint`

As well as pointers to all of these types. See also: [Custom Types](#custom-types)

<!-- TOC --><a name="custom-types"></a>

### Custom Types

A custom Cofigurature type specifies the type of struct field it is for, and how to interact with it by satisfying Configurature's `Value` interface.

```go
type Value interface {
    String() string
    Set(string) error
    Type() string
    Interface() interface{}
}
```

You may be writing a custom type to configure a Go struct field type that is specific to your application or to a library used by your application.

Here is the definition of an `ThumbnailFile` type that only accepts certain image file types and file size limits.

```go
type (
	// go type to be set as config struct field types
	// type Config struct { FieldName ThumbnailFile `....` }
	ThumbnailFile string

	// type for Value interface
	thumbnailValue ThumbnailFile
)

// String value of type
func (t *thumbnailValue) String() string {
	return (string)(*t)
}

// Set is always called with a string and should return an error if the string
// can not be converted to the underlying type
func (t *thumbnailValue) Set(v string) error {

	// This will fail if the file does not exist or there is any other error
	// accessing the file
	if st, err := os.Stat(v); err != nil {
		return err
	} else if st.Size() > 5000000 {
		return errors.New("file must be less than 5MB")
	}
	// This will fail if the file is not of the supported type
	switch ext := path.Ext(strings.ToLower(v)); ext {
	case ".png", ".jpg", ".jpeg", ".gif":
		// ok
	default:
		return fmt.Errorf("file type \"%s\" not supported", ext)
	}
	*t = (thumbnailValue)(v)
	return nil
}

// Name of the type
func (i *thumbnailValue) Type() string {
	return "Thumbnail"
}

// Return the value of the type converted to its field value type
func (t *thumbnailValue) Interface() interface{} {
	return ThumbnailFile(*t)
}

```

Add the type using `AddType()`

```go
// ThumbnailFile is the struct field type
// thumbnailValue is the struct type created above to implement the Value interface
AddType[ThumbnailFile, thumbnailValue]()
```

Specify as struct field type

```go
type Config struct {
    ProductImage ThumbnailFile `desc:"Path to thumbnail for product"`
}
```

This is just an example. In most cases a validator will satisfy what you want to do, or a `string` field with an `enum:"..."` tag. However, if a Configurature struct field uses an app specific type, you will need to define a custom type or use a [map value type](#map-value-types) in order to use it or use some translation to convert it to its type.

<a name="map-value-types"></a>

#### Map Value Types

The function `AddMapValueType` can be used to create and register string to value map types with Configurature. This is all the code required to implement the `slog.Level` custom type in Configurature:

```go
AddMapValueType[slog.Level]("Level", map[string]slog.Level{
    "debug": slog.LevelDebug,
    "info":  slog.LevelInfo,
    "warn":  slog.LevelWarn,
    "error": slog.LevelError,
})
```

The type parameter is the struct field type, the first argument is the name of the type as it will appear in Usage() text, and the second is the string -> value map. If you specify a blank string for the type name, it will default to the name of the type minus its package name.

Defining this in a config struct looks like

```go
type Config struct {
    LogLevel slog.Level `desc:"Log level of app" default:"info"`
}
```

Usage text looks like

```
--log_level Level   Log level (debug|info|warn|error) (default debug)
```

Note that this adds the available options to the map from the help text. Since map keys are not ordered, the order of these options will be randomized. If you want them to appear in the same order every time you may use the `enum:"..."` [tag](#tags) to specify it.

The type used in `AddMapValueType` must not be a type already handled by Configurature (common types like string, int, etc.). If you want to do this, you will have to create a new type.

```go
type Color string

AddMapValueType[Color]("", map[string]Color{
    "red":   "#ff0000",
    "blue":  "#0000ff",
    "green": "#00ff00",
})
```

This can be specified on a config struct using the `Color` type.

```go
type Config struct {
    Background Color `desc:"Color of the background" default:"red"`
    Text       Color `desc:"Color of text" default:"blue"`
}
```

And the Usage() text looks like

```
--background Color   Color of the background (red|blue|green) (default red)
--text       Color   Color of text (red|blue|green) (default blue)
```

Note that these are not case sensitive. `--text Blue` will be treated as `--text blue`.

<!-- TOC --><a name="license"></a>

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for details.                           

## License 

Apache 2.0; see [`LICENSE`](LICENSE) for details.                      

## Disclaimer                                                                   

This project is not an official Google project. It is not supported by
Google and Google specifically disclaims all warranties as to its quality,      
merchantability, or fitness for a particular purpose.

