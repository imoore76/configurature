============================
Struct Field Types
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

.. raw:: html

   <!-- TOC -->

Custom Types
-------------

A custom Configurature type specifies the type of struct field it is
for, and how to interact with it by satisfying Configurature’s ``Value``
interface.

.. code-block:: go
    :caption: Value interface

    type Value interface {
        String() string
        Set(string) error
        Type() string
        Interface() interface{}
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

    func init() {

        // ThumbnailFile is the struct field type
        // thumbnailValue is the struct type created above to implement the Value interface
        configurature.AddType[ThumbnailFile, thumbnailValue]()
    }


Add the type using Configurature's ``AddType()`` function as exemplified above.

The struct field type can be used in a Configurature struct like so:

.. code:: go

   type Config struct {
       ProductImage ThumbnailFile `desc:"Path to thumbnail for product"`
   }

This is just an example. In most cases a validator or a ``string`` field with an ``enum:"..."``
tag will satisfy the use case. However,
if a Configurature struct field uses an app specific type, you will need
to define a custom type or use a `map value type <#map-value-types>`__
in order to use it or use some translation to convert it to its type.

Map Value Types
--------------------------

Map value types are custom types that are used to map strings to a
custom set of values. Use ``AddMapValueType[T any](string, map[string]T)``
(usually in an ``init()`` function)
to create and register these types with configurature.

The type argument ``[T any]`` is the custom value type and can usually be
omitted because it is inferred
from the map value type. The string argument will be the name of the type
in ``Usage()`` text and will default to the type's name.
The map argument is the string -> value map.

Map Value Type Examples
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Here are some real-world examples.

Log Level
##########

This is all the code
required to implement the ``slog.Level`` custom type in Configurature:

.. code:: go

    func init() {
        configurature.AddMapValueType("Level", map[string]slog.Level{
            "debug": slog.LevelDebug,
            "info":  slog.LevelInfo,
            "warn":  slog.LevelWarn,
            "error": slog.LevelError,
        })
    }

Defining this in a config struct looks like

.. code:: go

   type Config struct {
       LogLevel slog.Level `desc:"Log level of app" default:"info"`
   }

Usage text looks like

::

   --log_level Level   Log level (debug|info|warn|error) (default debug)

.. important::
    
    Since map keys are not ordered, the order of these options will be
    randomized. If you want them to appear in the same order every time you
    may use the ``enum:"..."`` :ref:`tag<usage:tags>` to specify it.

.. code:: go

   type Config struct {
       LogLevel slog.Level `desc:"Log level of app" enum:"debug,info,warn,error" default:"info"`
   }

Color
##########
.. warning::

    The type used in ``AddMapValueType`` can not be a type that
    is already handled
    by Configurature (common types like string, int, etc.). If you want to
    reuse an existing type, you will have to create a new one that derives from
    the existing type. E.g. ``type Color string`` below.

.. code:: go

    type Color string

    func init() {
        configurature.AddMapValueType("", map[string]Color{
            "red":   "#ff0000",
            "blue":  "#0000ff",
            "green": "#00ff00",
        })
    }

This can be specified on a config struct using the ``Color`` type.

.. code:: go

    type Config struct {
        Background Color `desc:"Color of the background" default:"red"`
        Text       Color `desc:"Color of text" default:"blue"`
    }


Delay
############

.. important::
    
    Since the ``time.Duration`` type is already supported by Configurature,
    a derived type is created in Go.

.. code:: go

    type Delay time.Duration

    func init() {
        configurature.AddMapValueType("", map[string]Delay{
            "short":  Delay(1 * time.Minute),
            "medium": Delay(5 * time.Minute),
            "long":   Delay(10 * time.Minute),
        })
    }

.. note::

    In some cases, you may need to cast the value to the type. For example,
    ``Delay(1 * time.Minute)`` etc. above.

.. code:: go

   type Config struct {
        WaitTime   Delay `desc:"Delay time" default:"medium"`
        Background Color `desc:"Color of the background" default:"red"`
        Text       Color `desc:"Color of text" default:"blue"`
   }
