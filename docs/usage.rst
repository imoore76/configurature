================
Usage
================

Basic Usage and Examples
================================

Declare your configuration and run ``Configure[]()``

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

The tags ``desc`` add descriptions to the fields and are required.
``default`` specifies a default value.

.. seealso::

    See :ref:`Tags<usage:tags>` for a complete list of tags.

Help text
---------------------------

Running this app with ``--help`` displays the app usage:

.. code-block:: shell

    user@host $ myapp --help
    Command usage:
    -h, --help               show help and exit
        --listen_ip ip       IP address on which to listen (default 127.0.0.1)
        --listen_port uint   port on which to listen (default 8080)


Specifying Config Values
==========================
The following methods can be used to set configuration values in order of 
precedence.


Default Value
---------------------------
The default value (if specified) in the struct field's
``default`` tag will be used
if no other value is specified.

.. code-block:: go

    type AppConfig struct {
        ListenIP   net.IP `desc:"IP address on which to listen" default:"127.0.0.1"`
        ListenPort uint   `desc:"port on which to listen" default:"8080"`
    }


.. code-block:: shell

    user@host ~$ myapp
    IP: 127.0.0.1
    Port: 8080


Without a default value, the zero value of the type will be used.

CLI Flags
---------------------------
Flags can be specified on the command line. Short flags are also supported
by adding a ``short:"x"`` tag.

.. code-block:: go

    type AppConfig struct {
        ListenIP   net.IP `desc:"IP address on which to listen" short:"i" default:"127.0.0.1"`
        ListenPort uint   `desc:"port on which to listen" default:"8080"`
    }

.. code-block:: shell

    user@host ~$ myapp -i 2.2.2.2
    IP: 2.2.2.2
    Port: 8080

    user@host ~$ myapp --listen_ip 0.0.0.0
    IP: 0.0.0.0
    Port: 8080

Environment variables
---------------------------
You can also use environment variables in the form of uppercase arguments
prefixed with the ``EnvPrefix`` option.

.. code-block:: go

    type AppConfig struct {
        ListenIP   net.IP `desc:"IP address on which to listen" short:"i" default:"127.0.0.1"`
        ListenPort uint   `desc:"port on which to listen" default:"8080"`
    }

    conf := co.Configure[AppConfig](&co.Options{
        EnvPrefix: "MYAPP_",
    })

.. code-block:: shell

    user@host ~$ MYAPP_LISTEN_PORT=443 myapp --listen_ip 0.0.0.0
    IP: 0.0.0.0
    Port: 443


.. seealso::
    
    See the ``--print_env_template`` flag documented in :ref:`Using env files<usage:using .env files>`.
    This can be used to list all the environment variables that can be set.

Using .env files
---------------------------

Configurature itself does not provide anything that
parses ``.env`` files since there are many other tools
that have filled this space. Here are a couple favorites:

* godotenv - https://github.com/joho/godotenv
* direnv - https://direnv.net/

Since Configurature automatically uses environment variables,
there is no further action needed after loading environment variables from a ``.env`` file.
To create a template ``.env`` file, you can run your app with a ``--print_env_template`` flag.
This flag is hidden so won't appear in your app's Usage() text when using ``--help``.

.. code-block:: shell

    user@host ~$ myapp --print_env_template

    # IP address on which to listen (default 127.0.0.1)
    APP_LISTEN_IP="127.0.0.1"

    # port on which to listen (default 8080)
    APP_LISTEN_PORT="8080"

Copy and paste, or redirect its output to a ``.env`` file,
then edit as needed.

Config Files
---------------------------
A config file can be added by adding a special ``ConfigFile`` field to the
configuration struct. ``ConfigFile`` is part of the ``configurature`` package.

.. code-block:: go

    type AppConfig struct {
        ListenIP   net.IP        `desc:"IP address on which to listen" default:"127.0.0.1"`
        ListenPort uint          `desc:"port on which to listen" default:"8080"`
        Conf       co.ConfigFile `desc:"configuration file" short:"c"`
    }

The example above also adds a ``short`` tag to specify a short version of the option.

.. code-block:: shell

    user@host ~$ myapp -h
    Command usage:
    -c, --conf configFile    configuration file
    -h, --help               show help and exit
        --listen_ip ip       IP address on which to listen (default 127.0.0.1)
        --listen_port uint   port on which to listen (default 8080)


Supported configuration file formats are yaml and json 
(determined by file extension). Adding the configuration options
to a ``conf.yaml`` file looks like

.. code-block:: yaml

    # conf.yaml
    listen_ip: 0.0.0.0
    listen_port: 80

.. code-block:: shell

    user@host ~$ myapp -c conf.yaml
    IP: 0.0.0.0
    Port: 80

Resolution order of values is command line, environment variable, and
finally configuration file.

Generating Config Files
---------------------------

You can use the internal hidden flag ``--print_yaml_template``
to generate a YAML template file. Specifying values along
with using the flag uses those values for the configuration
file output.

.. code-block:: shell

    user@host ~$ myapp --listen_ip 0.0.0.0 --print_yaml_template
    # Generated from
    # [--listen_ip 0.0.0.0 --print_yaml_template]

    # IP address on which to listen (default 127.0.0.1)
    listen_ip: 0.0.0.0

    # port on which to listen (default 8080)
    listen_port: 8080


Copy and paste, or redirect its output to a ``.yaml`` file,
and edit as needed.

Options
=======================
``Configure[]()`` can be called with an ``Options`` pointer as input.

.. code-block:: go

    // Config options
    type Options struct {
        EnvPrefix         string              // Prefix for environment variables
        Args              []string            // Arguments to parse
        NilPtrs           bool                // Leave pointers set to nil if values aren't specified
        Usage             func(*flag.FlagSet) // Usage function called when configuration is incorrect or for --help
        NoRecover         bool                // Don't recover from panic
        ShowInternalFlags bool                // Show hidden internal flags
        NoShortHelp       bool                // Don't add "h" as a short help flag
    }

If not specified, the defaults are:

.. code-block:: go

    Options{
        EnvPrefix:         "",
        Args:              os.Args[1:],
        NilPtrs:           false,
        Usage:             // internally composed
        NoRecover:         false,
        ShowInternalFlags: false,
        NoShortHelp:       false,
    }


EnvPrefix
-----------------------

The prefix to use when checking for configuration values specified as
environment variables. 

.. important::

    If not set, Configurature will not use environment variables.


Args
-----------------------

The string of arguments to parse. Typically this would be set to,
and defaults to ``os.Args[1:]`` (the first element of ``os.Args``
is the name of the command being run, so is not included).


NilPtrs
-----------------------

When using config field pointers, leave them set to ``nil`` rather than a zero
value if no value is specified and no default is set.

.. code-block:: go

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

This would result in ``ListenPort is 8080`` since no value was specified and no default was provided for MaxConns. With the default behavior, MaxConns would be a pointer to a zero value (`0` for ints).

NoRecover
-----------------------

The default behavior is to handle a ``panic()`` that occurs
during parsing, print its message to ``os.Stderr`` and exit.

.. code-block:: go

    if r := recover(); r != nil {
        fmt.Fprintf(os.Stderr, "error parsing configuration: %s\n", r)
        os.Exit(1)
    }

You can disable this behavior by setting ``NoRecover`` to true.
You may then handle panics in your app however you'd like.

.. code-block:: go

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

ShowInternalFlags
----------------------

Show internal flags ``--print_env_template`` and ``--print_yaml_template`` in
Usage() text.


Usage
----------------------

Create and specify your own handler for command usage help. This occurs when incorrect configuration
flags are supplied or for ``--help``.
The default ``Usage`` function is 

.. code-block:: go

    func(f *pflag.FlagSet) {
        fmt.Println("Command usage:")
        fmt.Println(f.FlagUsages())
        os.Exit(0)
    }

NoShortHelp
----------------------
Do not add ``-h`` as a short flag for help. This may be useful if there is a field that you want to use
``-h`` for.

.. code-block:: go

    type Config struct {
        HangTime time.Duration `desc:"Time to hang" short:"h" default:"1m"`
    }

    func main() {

        conf := co.Configure[Config](&co.Options{
            NoShortHelp: true,
        })

        fmt.Println(conf.HangTime)
    }

.. code-block:: shell
    
    user@host ~$ myapp --help
    Command usage:
    -h, --hang_time duration   Time to hang (default 1m0s)
        --help                 show help and exit

    user@host ~$ myapp -h 2h
    2h0m0s


Tags
=======================

The following struct tags are used by Configurature:

.. list-table::
    :header-rows: 1

    * - Tag
      - Description
    * - ``desc:"..."``
      - Provides a description for the field shown in the ``Usage`` message
    * - ``default:"..."``
      - Specifies the default value for the field
    * - ``short:"..."``
      - Specifies the short version of the flag
    * - ``name:"..."``
      - Specifies an alternate name for the field instead of using the struct field name converted to snake case
    * - ``hidden:""``
      - Hides the field from the "Usage" help displayed. It can still be specified on the cli, environment variable, or config file.
    * - ``ignore:""``
      - Makes Configurature completely ignore the struct field
    * - ``enum:"x,y,z"``
      - Only the values ``x``, ``y``, or ``z`` are valid for this field. This will automatically add the values to the help text of the field.
    * - ``validate:"..."``
      - Specifies `validators <#validators>`_ for this field separated by ``,``

.. note::
    
    Looking for a tag to indicate a field is required?
    Use the ``required`` or ``not_blank`` :doc:`validator </validators>`.
