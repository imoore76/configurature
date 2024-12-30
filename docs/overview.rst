=============
Overview
=============

Configurature is a Go library that provides declarative app configuration using structs.
Configuration values can be specified (in value precedence order) on the command line,
using environment variables, and/or in a config file (yaml or json).

Configuration structs can be composed in a way that your application's entry points do not
need to be aware of the structure of other packages' configurations in order to initialize them.

Basic Example
=============

.. code-block:: go
    :caption: main.go

    package main

    import (
        "fmt"
        "net"

        co "github.com/imoore76/configurature"
    )

    type AppConfig struct {
        ListenIP   net.IP `desc:"IP address on which to listen" default:"127.0.0.1"`
        ListenPort uint   `desc:"port on which to listen" default:"8080"`
    }

    func main() {

        conf := co.Configure[AppConfig](nil)

        fmt.Printf("IP: %s\n", conf.ListenIP)
        fmt.Printf("Port: %d\n", conf.ListenPort)
    }


.. code-block:: shell
    :caption: defaults

    $ go run main.go
    IP: 127.0.0.1
    Port: 8080

.. code-block:: shell
    :caption: specify values on the command line

    $ go run main.go --listen_ip=0.0.0.0 --listen_port=22
    IP: 0.0.0.0
    Port: 22

.. code-block:: shell
    :caption: help text

    $ go run main.go --help
    Command usage:
        -h, --help           show help and exit
        --listen_ip ip       IP address on which to listen (default 127.0.0.1)
        --listen_port uint   port on which to listen (default 8080) 

.. code-block:: shell
    :caption: validation

    $ go run main.go --listen_ip=5
    invalid argument "5" for "--listen_ip" flag: failed to parse IP: "5"
    Command usage:
        -h, --help           show help and exit
        --listen_ip ip       IP address on which to listen (default 127.0.0.1)
        --listen_port uint   port on which to listen (default 8080) 

Installation
=============

.. code-block:: go

    // anywhere in your application
    include "github.com/imoore76/configurature"

then

.. code-block:: bash

    $ go mod tidy
