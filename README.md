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
	Host     string `validate:"required"` // this field is required
	Port     int    `desc:"Database port" default:"5432"`
	User     string `desc:"Database user" default:"postgres"`
	Password string `desc:"Database password" default:"postgres"`
}

// Network server config struct. Could also reside in a different package.
type ServerConfig struct {
	ServerName  string `name:"hostname"` // rename this field in the config
	ReadTimeout int    // no struct tags are required
	ListenIP    net.IP `default:"127.0.0.1"`
	ListenPort  uint   `default:"8080"`
}

type Config struct {
	ServerConfig                  // Embedded struct
	CalculatedField string        `ignore:"true"` // ignore this field
	DB              DBConfig      // Sub-config in `DB` struct
	LogLevel        string        `desc:"Log level" default:"info"` // direct field
	Conf            co.ConfigFile `desc:"Configuration file" short:"c"`
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
Command usage:
  -c, --conf configFile      Configuration file
      --db_host string       db host
      --db_password string   Database password (default "postgres")
      --db_port int          Database port (default 5432)
      --db_user string       Database user (default "postgres")
  -h, --help                 show help and exit
      --hostname string      hostname
      --listen_ip ip         listen ip (default 127.0.0.1)
      --listen_port uint     listen port (default 8080)
      --log_level string     Log level (default "info")
      --read_timeout int     read timeout
```

CLI option and environment variable example:
```shell
user@host $ MYAPP_LISTEN_IP=0.0.0.0 myapp --listen_port 80 --db_host localhost
DB Host: localhost
Log level: info
IP: 0.0.0.0
Port: 80
```

Example config yaml file:
```yaml
hostname: "myapp.example.com"

listen_ip: 0.0.0.0
listen_port: 80

log_level: info

db:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
```

Configuration values can be specified on the command line, using environment variables, and/or in a config file.

Configurature also supports

* Custom types
* Validation
* Nested configurations

See the complete documentation at [http://configurature-docs.readthedocs.io](http://configurature-docs.readthedocs.io).


## Templates

Print config file template:
```shell
user@host $ myapp --print_yaml_template
# Generated with
# [--print_yaml_template]

# hostname
hostname: ""

# read timeout
read_timeout: 0

# listen ip
listen_ip: 127.0.0.1

# listen port
listen_port: 8080

db:

  # db host
  host: ""

  # Database port
  port: 5432

  # Database user
  user: postgres

  # Database password
  password: postgres

# Log level
log_level: info
```

Print environment variable template:

```shell
user@host $ myapp --print_env_template
# Generated with
# [--print_env_template]

# Configuration file
MYAPP_CONF=""

# db host
MYAPP_DB_HOST=""

# Database password
MYAPP_DB_PASSWORD="postgres"

# Database port
MYAPP_DB_PORT="5432"

# Database user
MYAPP_DB_USER="postgres"

# hostname
MYAPP_HOSTNAME=""

# listen ip
MYAPP_LISTEN_IP="127.0.0.1"

# listen port
MYAPP_LISTEN_PORT="8080"

# Log level
MYAPP_LOG_LEVEL="info"

# read timeout
MYAPP_READ_TIMEOUT="0"
```

Templates use existing values.

```shell
user@host $ MYAPP_HOSTNAME=server1 myapp --print_yaml_template
# Generated with
# [--print_yaml_template]

# hostname
hostname: server1

# read timeout
read_timeout: 0

# etc...
```

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for details.                           

## License 

Apache 2.0; see [`LICENSE`](LICENSE) for details.                      

## Disclaimer                                                                   

This project is not an official Google project. It is not supported by Google and Google specifically
disclaims all warranties as to its quality, merchantability, or fitness for a particular purpose.

