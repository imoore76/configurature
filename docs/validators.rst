============
Validators
============

.. |validator_link| raw:: html

    <a href="http://github.com/go-playground/validator" target="_blank">github.com/go-playground/validator</a>

.. |validator_docs| raw:: html

    <a href="https://pkg.go.dev/github.com/go-playground/validator/v10#readme-usage-and-documentation" target="_blank">its documentation</a>


Validators are specified in a field tag as ``validate:"validator1,validator2"``.
Some validators require an option and may be specified as ``validate:"validator1=foo,validator2=bar"``.
This functionality is provided by the excellent
|validator_link| package.
For more information, be sure to read |validator_docs|.

Validating Compound Types
============================

Some validators can act on both compound and simple values. For instance, given a field of type ``[]string`` the ``lt``
validator can specify the maximum number of elements in the slice or the maximum length of strings in the ``[]string``
slice. By default, validators will validate the field value - so if it is a slice, ``lt`` will apply to the number
of items in the slice. To validate items within a compound value, use ``dive``.

.. code-block:: go
    :caption: No more than 5 names can be specified. Each name must be at lest 3 chars long.

    type Config struct {
        Names []string `desc:"names" validate:"lt=6,dive,gte=3"`
    }

To validate map keys, you can use ``keys`` to specify that the following validators are applied to the map's keys and
``endkeys`` to specify that the remaining validators are to be applied to the map's values.

.. code-block:: go
    :caption: At least one name/age pair must be  specified. Names must be > 3 characters. Ages must be >= 18

    type Config struct {
        Ages map[string]int `desc:"Names and ages map" validate="required,keys,gt=3,endkeys,dive,gte=18"`
    }

Network Validators
==================

.. list-table::
    :header-rows: 1

    * - Validator
      - Description
    * - ``cidr``
      - Classless Inter-Domain Routing CIDR
    * - ``cidrv4``
      - Classless Inter-Domain Routing CIDRv4
    * - ``cidrv6``
      - Classless Inter-Domain Routing CIDRv6
    * - ``datauri``
      - Data URL
    * - ``fqdn``
      - Full Qualified Domain Name (FQDN)
    * - ``hostname``
      - Hostname RFC 952
    * - ``hostname_port``
      - ``host:port`` format  
    * - ``ip``
      - Internet Protocol Address IP
    * - ``ipv4``
      - Internet Protocol Address IPv4
    * - ``ipv6``
      - Internet Protocol Address IPv6
    * - ``mac``
      - Media Access Control Address MAC
    * - ``unix_addr``
      - Unix domain socket end point Address (path to socket file)
    * - ``uri``
      - URI String
    * - ``url``
      - URL String
    * - ``http_url``
      - HTTP URL String (``http`` and ``https`` will both validate)
    * - ``url_encoded``
      - URL Encoded

String Validators
=================

.. list-table::
    :header-rows: 1

    * - Validator
      - Description
      - Option
    * - ``alpha``
      - Alphabet Only
      -
    * - ``alphanum``
      - Alphanumeric
      -
    * - ``alphanumunicode``
      - Alphanumeric Unicode
      -
    * - ``alphaunicode``
      - Alphabet Unicode
      -
    * - ``ascii``
      - ASCII
      -
    * - ``boolean``
      - Boolean. Can be any one of ``1``, ``t``, ``T``, ``TRUE``, ``true``, ``True``, ``0``, ``f``, ``F``, ``FALSE``, ``false``, ``False``
      -
    * - ``contains=...``
      - Contains
      - the string that must be contained in the value
    * - ``containsany=...``
      - Contains Any
      - the string with the characters of which at least one must be in the value
    * - ``endsnotwith=...``
      - Does not end with
      - the string that the value must not end with
    * - ``endswith=...``
      - Ends With
      - the string that the value must end with
    * - ``excludes=...``
      - Excludes
      - the string that must not be contained in the value
    * - ``excludesall=...``
      - Excludes All
      - the string with the characters of which none may be in the value
    * - ``lowercase``
      - Lowercase
      -
    * - ``multibyte``
      - Multi-Byte Characters
      -
    * - ``not_blank``
      - The value must contain non-whitespace characters
      -
    * - ``number``
      - Number - e.g. ``8``, ``4.2`` ``-20``
      -
    * - ``numeric``
      - Must be a number or a string that can be parsed as a number
      -
    * - ``printascii``
      - Printable ASCII
      -
    * - ``regex=...``
      - The value must match the supplied regex
      - A valid regular expression. E.g. ``^\d{3}-\d{3}-\d{4}$``
    * - ``startsnotwith=...``
      - Does not start with
      - the string with which the value must not start
    * - ``startswith=...``
      - Starts With
      - the string with which the value must start
    * - ``uppercase``
      - Uppercase
      -

String Format Validators
========================

.. list-table::
    :header-rows: 1

    * - Validator
      - Description
    * - ``base64``
      - Base64 String
    * - ``btc_addr``
      - Bitcoin Address
    * - ``btc_addr_bech32``
      - Bitcoin Bech32 Address (segwit)
    * - ``mongodb``
      - MongoDB ObjectID
    * - ``cron``
      - Cron schedule string
    * - ``spicedb``
      - SpiceDb ObjectID/Permission/Type
    * - ``datetime``
      - Datetime
    * - ``e164``
      - e164 formatted phone number
    * - ``email``
      - E-mail String
    * - ``eth_addr``
      - Ethereum Address
    * - ``hexadecimal``
      - Hexadecimal String
    * - ``hexcolor``
      - Hexcolor String
    * - ``hsl``
      - HSL String
    * - ``hsla``
      - HSLA String
    * - ``html``
      - HTML Tags
    * - ``html_encoded``
      - HTML Encoded
    * - ``isbn``
      - International Standard Book Number
    * - ``isbn10``
      - International Standard Book Number 10
    * - ``isbn13``
      - International Standard Book Number 13
    * - ``issn``
      - International Standard Serial Number
    * - ``iso3166_1_alpha2``
      - Two-letter country code (ISO 3166-1 alpha-2)
    * - ``iso3166_1_alpha3``
      - Three-letter country code (ISO 3166-1 alpha-3)
    * - ``iso3166_1_alpha_numeric``
      - Numeric country code (ISO 3166-1 numeric)
    * - ``iso3166_2``
      - Country subdivision code (ISO 3166-2)
    * - ``iso4217``
      - Currency code (ISO 4217)
    * - ``json``
      - JSON
    * - ``latitude``
      - Latitude
    * - ``longitude``
      - Longitude
    * - ``postcode_iso3166_alpha2``
      - Postcode
    * - ``postcode_iso3166_alpha2_field``
      - Postcode
    * - ``rgb``
      - RGB String
    * - ``rgba``
      - RGBA String
    * - ``ssn``
      - Social Security Number SSN
    * - ``timezone``
      - Timezone
    * - ``uuid``
      - Universally Unique Identifier UUID
    * - ``uuid3``
      - Universally Unique Identifier UUID v3
    * - ``uuid4``
      - Universally Unique Identifier UUID v4
    * - ``uuid5``
      - Universally Unique Identifier UUID v5
    * - ``md5``
      - MD5 hash
    * - ``sha256``
      - SHA256 hash
    * - ``sha384``
      - SHA384 hash
    * - ``sha512``
      - SHA512 hash
    * - ``ripemd128``
      - RIPEMD-128 hash
    * - ``ripemd160``
      - RIPEMD-160 hash
    * - ``tiger128``
      - TIGER128 hash
    * - ``tiger160``
      - TIGER160 hash
    * - ``tiger192``
      - TIGER192 hash
    * - ``semver``
      - Semantic Versioning 2.0.0
    * - ``ulid``
      - Universally Unique Lexicographically Sortable Identifier ULID
    * - ``cve``
      - Common Vulnerabilities and Exposures Identifier (CVE id)

Comparision Validators
======================

.. list-table::
    :header-rows: 1

    * - Validator
      - Description
    * - ``gt=...``
      - Greater than
    * - ``gte=...``
      - Greater than or equal
    * - ``lt=...``
      - Less Than
    * - ``lte=...``
      - Less Than or Equal
    * - ``ne=...``
      - Not Equal

.. note::

    For numeric values, the comparison is treated as numeric. For string values, the length of the string is used 
    for comparison. For slices, maps, and arrays, the number of items is used for comparison.

Miscellaneous Validators
==============================

.. list-table::
    :header-rows: 1

    * - Validator
      - Description
      - Option
    * - ``dir``
      - Existing Directory
      -
    * - ``dirpath``
      - Directory Path
      -
    * - ``file``
      - Existing File
      -
    * - ``filepath``
      - File Path
      -
    * - ``image``
      - File path to a valid image
      -
    * - ``len=...``
      - Exact length
      - Exact length of string, slice, array, or map.
    * - ``max=...``
      - Maximum length
      - Max length of string, slice, array, or map.
    * - ``min=...``
      - Minimum length
      - Min length of string, slice, array, or map.
    * - ``required``
      - A value must be supplied
      -