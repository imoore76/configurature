Combining Configurations
=========================

A feature of Configurature is that you can combine configurations from
multiple packages of your application into a single configuration and
initialize it from a single location (e.g. ``main()``). Given the
following project layout

::

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

Typically each of these packages has its own configuration requirements.
The server package is in charge of passing the ``ipcSocketFile`` to the
``ipc`` package instantiation and the ``rpcTimeout`` to the ``rpc``
package. The ``main`` package needs to pass config to ``server``. They
can be passed around as distinct function arguments, which may be some
variation of:

.. code:: go

   s := server.New(listenIp, listenPort, ipcSocketFile, rpcTimeout)

This can become unwieldy as the number of configuration options grows.
It also requires that packages be aware of implementation details of
other packages. If the ``ipc`` package is changed to require a
``ipcTimeout`` configuration option, then the ``server`` package must be
updated to pass this new configuration option.

Configurature provides a way to avoid this scenario. So that
adding a configuration option to the ``ipc`` package is as easy as
adding a single field to its configuration struct. Neither the
``server`` or ``main`` packages need to be made aware of the new
configuration or any change in configuration at all.

Nested Config
-------------

You can nest configurations in configuration structs so that they are
distinct fields in a parent struct.

.. code:: go

   // main.go
   type Config struct {
       ConfigFile co.ConfigFile `desc:"configuration file" short:"c"`
       Logging    log.Config
       Server     server.Config
   }

.. code:: go

   // server.go
   type Config struct {
       ListenIP   net.IP `desc:"IP address on which to listen" default:"127.0.0.1"`
       ListenPort uint   `desc:"port on which to listen" default:"8080"`
       IPC        ipc.Config
       RPC        rpc.Config
   }

Now ``server`` and ``log`` are initialized in ``main()`` with:

.. code:: go

       conf := co.Configure[Config](&co.Options{
           EnvPrefix: "APP_",
           Args:      os.Args[1:],
       })

       log.SetupLogging(&conf.Logging)
       s := server.New(&conf.Server)
       s.Run()

``ipc`` and ``rpc`` can be similarly initialized from the ``server``
package without needing to be aware of their configuration details.
Changes in configuration do not result in changes to function or method
calls.

Nested configuration does result in nested value specification. E.g.
``SocketFile`` in the ``ipc`` configuration becomes a
``--server_ipc_socket_file`` flag, ``APP_SERVER_IPC_SOCKET_FILE``
environment variable, and in a configuration file:

.. code:: yaml

   # conf.yaml
   server:
     ipc:
       socket_file: /tmp/server-ipc.sock

You can name sub configs and even give them empty names.

.. code:: go

   // main.go
   type Config struct {
       ConfigFile co.ConfigFile `desc:"configuration file" short:"c"`
       Logging    log.Config
       Server     server.Config `name:""`
   }

.. code:: go

   // server.go
   type Config struct {
       ListenIP   net.IP     `desc:"IP address on which to listen" default:"127.0.0.1"`
       ListenPort uint       `desc:"port on which to listen" default:"8080"`
       IPC        ipc.Config `name:"other"`
       RPC        rpc.Config
   }

``ListenIP`` is now specified using ``--listen_ip`` instead of
``--server_listen_ip`` because its name is empty. IPC configuration is
prefixed with ``other_``. E.g. ``--other_socket_file`` from the command
line. Naming applies to environment variables and config file structure
as well.

Flat Config
-----------

You can also include other config structs as anonymous fields.

.. code:: go

   // main.go
   type Config struct {
       ConfigFile co.ConfigFile `desc:"configuration file" short:"c"`
       log.LogConfig
       server.ServerConfig
   }

.. code:: go

   // server.go
   type ServerConfig struct {
       ListenIP   net.IP `desc:"IP address on which to listen" default:"127.0.0.1"`
       ListenPort uint   `desc:"port on which to listen" default:"8080"`
       ipc.IPCConfig
       rpc.RPCConfig
   }

Now ``server`` and ``log`` are initialized in ``main()`` with:

.. code:: go

       conf := co.Configure[Config](&co.Options{
           EnvPrefix: "APP_",
           Args:      os.Args[1:],
       })

       log.SetupLogging(&conf.LogConfig)
       s := server.New(&conf.ServerConfig)
       s.Run()

A downside of this is that the names config structs from each package
must be unique and there can not be duplicate field names between
structs. This is usually fine for small projects. A hybrid approach can
also be used where ``server`` is a flat config including ``ipc`` and
``rpc`` configs anonymously and ``main``\ s config contains concrete
fields that hold configurations for other packages.

Configuration structs included anonymously result in flat value
specification. E.g. ``SocketFile`` in the ``ipc`` configuration becomes
a ``--socket_file`` flag, ``APP_SOCKET_FILE`` environment variable. You
may want to rename the field name in this case to ``IpcSocketFile`` or
just let the field’s ``desc`` provide context for what “socket file”
refers to.

Mixed
-----

You are free to mix and match in a way that makes sense for your
project. ``Configure()`` will ``panic()`` if there are duplicate field names or
short flag names instead of quietly resulting in unintended
configuration.


Using Get
-------------

You can also use ``configurature.Get[T]()`` from anywhere in your app as long as
``Configure[T]()`` has previously been called.

.. code:: go

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

Then anywhere else in our code, you can call ``Get[T]()`` where T is the
type of config you want to retrieve from the top-level configuration.

.. code:: go

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
