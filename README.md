# enumgen

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=white)](https://pkg.go.dev/github.com/creachadair/enumgen)
[![Go Report Card](https://goreportcard.com/badge/github.com/creachadair/enumgen)](https://goreportcard.com/report/github.com/creachadair/enumgen)

Program `enumgen` is a command-line tool to generate Go enumeration types.

## Usage

This gool is intended for use with the "[go generate][gogen]" command.

The generator reads a configuration file in YAML format (see [`gen.Config`][gc]).
To generate types, add a rule like this:

```go
//go:generate -command enumgen go run github.com/creachadair/enumgen@latest
//go:generate enumgen -config enums.yml -output generated.go
```

## Type Structure

The generated type for an enumeration is a struct with an unexported pointer to
the string representation of the enumerator. This representation allows cheap
value comparisons, enumerators can be used as map keys, and users of the type
outside the package cannot create new non-zero values of the type. The zero
value is explicitly defined as the "unknown" value for an enumeration.

The generated type exports three methods:

- The `Enum` method returns the name of the generated type.

- The `Valid` method reports whether an enumerator is valid (non-zero).

- The `String` method returns a string representation for each enumerator,
  which defaults to the enumerator's base name.

## Configuration

The [`gen.Config`][gc] type defines a set of enumerations to generate in a
single package. The general structure of a config in YAML follows this example

```yaml
   package: "name"     # the name of the output package (required)
   enum:               # a list of enumeration types to generate
     - type: "Name"    # the type name for this enum
       prefix: "x"     # (optional) prefix to append to each enumerator name
       zero: "Bad"     # (optional) name of zero enumerator
       doc: "text"     # (optional) documentation comment for the enum type
       val-doc: "text" # (optional) aggregate documentation for the values
       values:
         - name: "A"   # the name of the first enumerator (required)
           doc: "text" # (optional) documentation for this enumerator
           text: "aaa" # (optional) string text for the enumerator
```

[gogen]: https://go.dev/blog/generate
[gc]: https://godoc.org/github.com/creachadair/enumgen/gen#Config
