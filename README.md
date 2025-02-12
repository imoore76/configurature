# Configurature

<p align="center">
<img src="logo.png" height="384" width="384" alt="configurature logo" />
</p>

Configurature is a Go library that provides declarative app configuration using structs.
Configuration values can be specified (in value precedence order) on the command line,
using environment variables, and/or in a config file (yaml or json).

Configuration structs can be composed in a way that your application's entry points do not
need to be aware of the structure of other packages' configurations in order to initialize them.

See the complete documentation at [http://configurature-docs.readthedocs.io](http://configurature-docs.readthedocs.io).

## Basic Usage

Basic usage consists of defining your configuration structs and running `configurature.Configure()`.

```go
package main

import (
	"fmt"
	"net"

	co "github.com/imoore76/configurature"
)

// DB config struct. Could also reside in your application's "db" package.
type DBConfig struct {
	Host     string `desc:"Database host (required)" validate:"required"`
	Port     int    `desc:"Database port" default:"5432" validate:"gte=1024"`
	User     string `desc:"Database user" default:"postgres"`
	Password string `desc:"Database password" default:"postgres"`
}

// Network server config struct. Could also reside in a different package.
type ServerConfig struct {
	ListenIP   net.IP `desc:"IP address on which to listen" default:"127.0.0.1"`
	ListenPort uint   `desc:"port on which to listen" default:"8080"`
}

type Config struct {
	ServerConfig               // Embedded struct
	DB           DBConfig      // Sub-config
	LogLevel     string        `desc:"Log level" default:"info"` // direct field
	Conf         co.ConfigFile `desc:"Configuration file" short:"c"`
}

func main() {

	conf := co.Configure[Config](&co.Options{
		EnvPrefix: "MYAPP_",
	})

	fmt.Printf("DB Host: %s\n", conf.DB.Host)
	fmt.Printf("Log level: %s\n", conf.LogLevel)
	fmt.Printf("IP: %s\n", conf.ListenIP)
	fmt.Printf("Port: %d\n", conf.ListenPort)
}
```

Running this app with `--help` displays the app usage:

```
user@host $ myapp --help
Command usage:
  -c, --conf configFile      Configuration file
      --db_host string       Database host (required)
      --db_password string   Database password (default "postgres")
      --db_port int          Database port (default 5432)
      --db_user string       Database user (default "postgres")
  -h, --help                 show help and exit
      --listen_ip ip         IP address on which to listen (default 127.0.0.1)
      --listen_port uint     port on which to listen (default 8080)
      --log_level string     Log level (default "info")
```

CLI option and environment variable example:
```
user@host $ MYAPP_LISTEN_IP=0.0.0.0 myapp --listen_port 80 --db_host localhost
DB Host: localhost
Log level: info
IP: 0.0.0.0
Port: 80
```

Configuration values can be specified on the command line, using environment variables, and/or in a config file.

Configurature also supports

* Custom types
* Validation
* Nested configurations

See the complete documentation at [http://configurature-docs.readthedocs.io](http://configurature-docs.readthedocs.io).

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for details.                           

## License 

Apache 2.0; see [`LICENSE`](LICENSE) for details.                      

## Disclaimer                                                                   

This project is not an official Google project. It is not supported by Google and Google specifically
disclaims all warranties as to its quality, merchantability, or fitness for a particular purpose.

