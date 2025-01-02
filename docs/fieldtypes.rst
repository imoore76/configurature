============================
Field Types
============================

Supported field types
-----------------------
The following field types are supported

-  ``[]bool``
-  ``[]float32``
-  ``[]float64``
-  ``[]int32``
-  ``[]int64``
-  ``[]int``
-  ``[]net.IP``
-  ``[]string``
-  ``[]time.Duration``
-  ``[]uint8``
-  ``[]uint``
-  ``bool``
-  ``configurature.ConfigFile``
-  ``float32``
-  ``float64``
-  ``int16``
-  ``int32``
-  ``int64``
-  ``int8``
-  ``int``
-  ``map[string]int64``
-  ``map[string]int``
-  ``map[string]string``
-  ``net.IPMask``
-  ``net.IPNet``
-  ``net.IP``
-  ``slog.Level``
-  ``string``
-  ``time.Duration``
-  ``uint16``
-  ``uint32``
-  ``uint64``
-  ``uint8``
-  ``uint``

As well as pointers to all of these types. Configurature also
allows for adding `Custom Types <#custom-types>`__.

Custom Types
-------------

A custom Configurature field type specifies an exported type and how to interact with
it by satisfying Configurature’s ``Value`` interface.

.. code-block:: go
    :caption: Value interface

    type Value interface {
        String() string     // String representation of value
        Set(string) error   // Set the internal value from string
        Type() string       // Type name to string
    }

You may be writing a custom type to configure a Go struct field type
that is specific to your application or to a library used by your
application.

Custom Type Example
^^^^^^^^^^^^^^^^^^^^^
Here is the definition of an ``ThumbnailFile`` type that only accepts
certain image file types and file size limits.

.. code-block:: go
    :caption: thumbnail.go

    type (
        // go type to be set as config struct field types
        // type Config struct { FieldName ThumbnailFile `....` }
        ThumbnailFile string
    )

    // String value of type
    func (t *ThumbnailFile) String() string {
        return (string)(*t)
    }

    // Set is always called with a string and should return an error if the string
    // can not be converted to the underlying type
    func (t *ThumbnailFile) Set(v string) error {

        // This will fail if the file does not exist or there is any other error
        // accessing the file
        if st, err := os.Stat(v); err != nil {
            return err
        } else if st.Size() >= 5000000 {
            return errors.New("file must be less than 5MB")
        }
        // This will fail if the file is not of the supported type
        switch ext := path.Ext(strings.ToLower(v)); ext {
        case ".png", ".jpg", ".jpeg", ".gif":
            // ok
        default:
            return fmt.Errorf("file type \"%s\" not supported", ext)
        }
        *t = (ThumbnailFile)(v)
        return nil
    }

    // Name of the type
    func (i *ThumbnailFile) Type() string {
        return "Thumbnail"
    }

    func init() {

        // ThumbnailFile is the struct field type
        configurature.AddType[ThumbnailFile]()
    }


Add the type using Configurature's ``AddType()`` function as exemplified above.

The struct field type can be used in a Configurature struct like so:

.. code-block:: go

   type Config struct {
       ProductImage ThumbnailFile `desc:"Path to thumbnail for product"`
   }

This is just an example. In most cases a validator or a ``string`` field with an ``enum:"..."``
tag will satisfy the use case. However,
if a Configurature struct field uses an app specific type, you will need
to define a custom type, use a
:ref:`map value type<fieldtypes:Map Value Custom Types>`,
or use some translation logic to convert it.

Slice of Custom Types
--------------------------
In order to use a slice of custom types, you will need to define a
:ref:`custom type<fieldtypes:Custom Types>` for the slice element type. For example, 
if you want to use a slice of ``ThumbnailFile`` types, you will need to
:ref:`define a custom type<fieldtypes:Custom Type Example>` for
``ThumbnailFile``.

Then you can add the type to Configurature ``AddType[[]<CustomType>]()``. For example:

.. code-block:: go

    func init() {
        configurature.AddType(ThumbnailFile)
        configurature.AddType([]ThumbnailFile)
    }

.. important::

    The slice type
    must be added after the element type.

The struct field type can be used in a Configurature struct like so:

.. code-block:: go

   type Config struct {
       ProductImages []ThumbnailFile `desc:"Paths to thumbnails for product"`
   }

Slice types are specified in CSV format for the CLI and environment variables.

.. code-block:: shell

    $ my_app --product_images "images/side.jpg,images/top.jpg,images/bottom.jpg"

In configuration files, arrays are used.

.. code-block:: yaml
    :caption: config.yaml

    product_images:
      - images/side.jpg
      - images/top.jpg
      - images/bottom.jpg


Map Value Custom Types
--------------------------

Map value types are custom types that are used to map strings to a
custom set of values. Use ``AddMapValueType(typeName string, keys []string, values []T)``
(usually in an ``init()`` function) to create and register these types
with configurature.

Its arguments are

* ``typeName`` - The name of the type in ``Usage()`` text. Defaults to the type's name.
* ``keys`` - The slice of keys to use for the map. A slice is used so that order is preserved.
* ``values`` - The slice of values that correspond with the keys.

See examples that follow.

Log Level Example
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

This is all the code
required to implement the ``slog.Level`` custom type in Configurature:

.. code-block:: go
    :caption: Register type

    func init() {
        configurature.AddMapValueType("",
            []string{
                "debug",
                "info",
                "warn",
                "error",
            },
            []slog.Level{
                slog.LevelDebug,
                slog.LevelInfo,
                slog.LevelWarn,
                slog.LevelError,
            },
        )
    }

Defining this in a config struct looks like

.. code-block:: go
    :caption: Define config

    type Config struct {
        LogLevel slog.Level `desc:"Log level of app" default:"info"`
    }

.. code-block:: shell
    :caption: Usage text

    --log_level Level   Log level (debug|info|warn|error) (default debug)

Color Example
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
.. warning::

    The type used in ``AddMapValueType`` can not be a type that
    is already handled
    by Configurature (common types like string, int, etc.). If you want to
    reuse an existing type, you will have to create a new one that derives from
    the existing type. E.g. ``type Color string`` below.

.. code-block:: go
    :caption: Register type

    type Color string

    func init() {
        configurature.AddMapValueType("",
            []string{
                "red",
                "blue",
                "green",
            },
            []Color{
                "#ff0000",
                "#0000ff",
                "#00ff00",
            },
        )
    }

This can be specified on a config struct using the ``Color`` type.

.. code-block:: go
    :caption: Define config

    type Config struct {
        Background Color `desc:"Color of the background" default:"red"`
        Text       Color `desc:"Color of text" default:"blue"`
    }


Delay Example
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. note::

    In some cases, you may need to cast the value to the type. For example,
    ``Delay(1 * time.Minute)`` etc. below.


.. code-block:: go
    :caption: Register type

    // time.Duration is already registered. Use a `Delay` derived type
    type Delay time.Duration

    func init() {
        configurature.AddMapValueType("",
            []string{
                "short",
                "medium",
                "long",
            },
            []Delay{
                Delay(1 * time.Minute),
                Delay(5 * time.Minute),
                Delay(10 * time.Minute),
            },
        )
    }

.. code-block:: go
    :caption: Define config

    type Config struct {
        WaitTime   Delay `desc:"Delay time" default:"medium"`
    }

Type Name Example
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. note::

    If the value type name is too long or otherwise undesirable, a 
    different type name can be specified. Below ``"Cluster"`` is specified
    so that the full
    type name ``"DeploymentClusterIdentifier"`` is not used.

.. code-block:: go
    :caption: Register type

    func init() {
        // "DeploymentClusterIdentifier" is a bit verbose.
        // Use "Cluster" instead.
        configurature.AddMapValueType("Cluster",
            []string{
                "dev",
                "staging",
                "prod",
            },
            []myapp.DeploymentClusterIdentifier{
                // dev
                uuid.FromString("ab0c39a5-9678-48d2-a534-7ba049988ae3"),
                // staging
                uuid.FromString("5ffeda13-9b47-404e-9985-12b7d77c9f5d"),
                // prod
                uuid.FromString("727953da-075d-4520-9aaf-9447e88f1677"),
            },
        )
    }

.. code-block:: go
    :caption: Define config

    type Config struct {
        DeployTo  DeploymentClusterIdentifier `desc:"Cluster in which to deploy" default:"dev"`
    }

.. code-block:: shell
    :caption: Help text

    $ myapp --help
    Command usage:
        --deploy_to Cluster   Cluster in which to deploy (dev|staging|prod) (default dev)
    -h, --help                show help and exit

