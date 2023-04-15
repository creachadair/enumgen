// Package gen implements an enumeration generator.
//
// Enumerations are described by a configuration.  Each enumeration defines a
// type name and one or more enumerator values.
//
// # Type Structure
//
// The generated type is a struct containing an unexported string pointer.
// Enumerators of the type can be compared for equality by value, and can be
// used as map keys. The zero value represents an unknown (invalid) enumerator;
// the Valid method reports whether an enumerator is valid (i.e., non-zero).
//
// The String method returns a string representation for each enumerator, which
// defaults to the enumerator's base name.  The Enum method returns the name of
// the enumeration type.
//
// Enumerations generated by this package all satisfy this interface:
//
//	type EnumType interface {
//	   Enum() string   // return the enumeration type name
//	   String() string // return the string representation of an enumerator
//	   Valid() bool    // report whether the receiver is a valid enumerator
//	}
//
// Callers wishing to accept arbitrary enumerations may define this interface.
// It is not exported by the gen package to discourage inappropriate dependency
// on the code generator.
//
// # Configuration
//
// The gen.Config type defines a set of enumerations to generate in a single
// package. The general structure of a config in YAML is:
//
//	package: "name"         # the name of the output package (required)
//	enum:                   # a list of enumeration types to generate
//	  - type: "Name"        # the type name for this enum
//	    prefix: "x"         # (optional) prefix to append to each enumerator name
//	    zero: "Bad"         # (optional) name of zero enumerator
//	    doc: "text"         # (optional) documentation comment for the enum type
//	    val-doc: "text"     # (optional) aggregate documentation for the values
//	    flag-value: true    # (optional) define flag.Value methods for the type
//	    text-marshal: true  # (optional) define text marshaling methods for the type
//	    values:
//	      - name: "A"       # the name of the first enumerator (required)
//	        doc: "text"     # (optional) documentation for this enumerator
//	        text: "aaa"     # (optional) string text for the enumerator
package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"strings"
)

// A Config specifies a collection of enumerations in a single package.
type Config struct {
	Package string  // package name for the generated file (required)
	Enum    []*Enum // enumerations to generate (at least one is required)
}

// An Enum defines an enumeration type.
//
// The generated type for an enumeration is a struct with an unexported pointer
// to the string representation of the enumerator. This representation allows
// cheap pointer comparisons, and users of the type outside the package cannot
// create new non-zero values of the type. The zero value is explicitly defined
// as the "unknown" value for an enumeration.
type Enum struct {
	Type   string   // enumeration type name (required)
	Values []*Value // the enumeration values (required)

	// If set, this prefix is prepended to each enumerator's variable name.
	// Otherwise, the variable name matches the Name field of the value.
	Prefix string

	// If set, this text is added as a doc comment for the enumeration.
	// Multiple lines are OK. The text should not contain comment markers.
	Doc string

	// If set, a variable is defined for the zero value with this name.
	// Typically a name like "Unknown" or "Invalid" makes sense.
	// Otherwise, no variable is defined for the zero value; the caller can
	// still construct a zero value explicitly if needed.
	Zero string

	// If set, this text is inserted at the top of the var block in the
	// generated code for the enumerator values.
	ValDoc string `yaml:"val-doc"`

	// If true, generate methods to implement flag.Value for the type.
	FlagValue bool `yaml:"flag-value"`

	// If true, implement encoding.TextMarshaler for the type.
	TextMarshal bool `yaml:"text-marshal"`
}

// A Value defines a single enumerator.
type Value struct {
	Name string // enumerator name (required)

	// If set, this text is added as a doc comment for the enumerator value.  If
	// it is a single line, it is added as a line comment; otherwise it is
	// placed before the enumerator. The text should not contain comment markers.
	// The placeholder {name} will be replaced with the final generated name of
	// the enumerator.
	Doc string

	// If set, this text is used as the string representation of the value.
	// Otherwise, the Name field is used.
	Text string
}

// Generate generates the enumerations defined by c into w as Go source text.
//
// If there is an error formatting the generated code, the unformatted code is
// still written to w before reporting the error. The caller should NOT use the
// output in case of error. Any error means there is a bug in the generator,
// and the output is written only to support debugging.
func (c *Config) Generate(w io.Writer) error {
	if err := c.checkValid(); err != nil {
		return err
	}

	var buf bytes.Buffer
	fmt.Fprint(&buf, "// Code generated by enumgen. DO NOT EDIT.\n\n")
	fmt.Fprintf(&buf, "package %s\n", c.Package)

	// If we are generating any flag or text marshaler values, import the "fmt"
	// package used by the generated code for error reporting.
	for _, e := range c.Enum {
		if e.FlagValue || e.TextMarshal {
			fmt.Fprintln(&buf, `import "fmt"`)
			break
		}
	}

	for _, e := range c.Enum {
		fmt.Fprintln(&buf)
		if err := e.generate(&buf); err != nil {
			return fmt.Errorf("enum %q: %w", e.Type, err)
		}
	}

	// Format the generated source. If this fails, write the unformatted source
	// to the output before reporting an error so the caller can debug.
	src, err := format.Source(buf.Bytes())
	if err != nil {
		w.Write(buf.Bytes())
		return fmt.Errorf("go format: %w", err)
	}
	_, err = w.Write(src)
	return err
}

// generate generates the enumeration defined by e into w.
func (e *Enum) generate(w io.Writer) error {
	if doc := formatDoc(e.Doc); doc != "" {
		fmt.Fprintln(w, doc)
	}

	// Generate the enumeration type.
	fmt.Fprintf(w, "type %s struct { *string }\n", e.Type)

	// Generate the Enum, String, and Valid methods.
	fmt.Fprintf(w, `
// Enum returns the name of the enumeration type for %[1]s.
func (%[1]s) Enum() string { return %[1]q }

// String returns the string representation of %[1]s v.
func (v %[1]s) String() string {
   if v.string == nil {
      return %[2]q
   }
   return *v.string
}

// Valid reports whether v is a valid %[1]s value.
func (v %[1]s) Valid() bool { return v.string != nil }
`, e.Type, "<invalid>")

	// Extract the label strings for the defined enumerators.
	labels := make([]string, len(e.Values))
	for i, v := range e.Values {
		if v.Text != "" {
			labels[i] = v.Text
		} else {
			labels[i] = v.Name
		}
	}
	strs := fmt.Sprintf("_str_%s", e.Type)

	// If requested, emit flag.Value methods.
	if e.FlagValue {
		fmt.Fprintf(w, `
// Set implements part of the flag.Value interface for %[1]s.
// A value must equal the string representation of an enumerator.
func (v *%[1]s) Set(s string) error {
   for i, opt := range %[2]s {
      if opt == s {
         v.string = &%[2]s[i]
         return nil
      }
   }
   return fmt.Errorf("invalid value for %[1]s: %%q", s)
}
`, e.Type, strs)
	}

	// If requested, emit text marshaling methods.
	if e.TextMarshal {
		fmt.Fprintf(w, `
// MarshalText encodes the value of the %[1]s enumerator as text.
// It satisfies the encoding.TextMarshaler interface.
func (v %[1]s) MarshalText() ([]byte, error) { return []byte(v.String()), nil }
`, e.Type)
		fmt.Fprintf(w, `
// UnarshalText decodes the value of the %[1]s enumerator from a string.
// It reports an error if data does not encode a known enumerator.
// An empty slice decodes to the invalid (zero) value.
// This method satisfies the encoding.TextUnmarshaler interface.
func (v *%[1]s) UnmarshalText(data []byte) error {
   text := string(data)
   if text == "" || text == (%[1]s{}).String() {
      v.string = nil
      return nil
   }
   for i, opt := range %[2]s {
      if opt == text {
         v.string = &%[2]s[i]
         return nil
      }
   }
   return fmt.Errorf("invalid value for %[1]s: %%q", text)
}
`, e.Type, strs)
	}

	// Generate the enumerators and string values.
	if doc := formatDoc(e.ValDoc); doc != "" {
		fmt.Fprintln(w, doc)
	}
	fmt.Fprintln(w, "var (")
	fmt.Fprintf(w, "\t%s = []string{", strs)
	for _, label := range labels {
		fmt.Fprintf(w, "%q,", label)
	}
	fmt.Fprint(w, "}\n\n")

	if e.Zero != "" {
		fmt.Fprintf(w, "\t%s%s = %s{}\n", e.Prefix, e.Zero, e.Type)
	}
	for i, v := range e.Values {
		fullName := e.Prefix + v.Name
		doc := formatDoc(injectName(v.Doc, fullName))
		multiline := strings.Contains(doc, "\n")
		if doc != "" && multiline {
			fmt.Fprintf(w, "\t%s\n", doc)
		}
		fmt.Fprintf(w, "\t%s = %s{&%s[%d]}", fullName, e.Type, strs, i)
		if doc != "" {
			if multiline {
				fmt.Fprintln(w) // extra space after documented enumerator
			} else {
				fmt.Fprint(w, "\t", doc)
			}
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, ")")
	return nil
}

// formatDoc reformats a doc string into Go line comments. Line breaks in the
// input are preserved. If s == "", the result is also empty.
func formatDoc(s string) string {
	if s == "" {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i, line := range lines {
		lines[i] = "// " + strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
}

// injectName replaces "{name}" markers in s with the specified name.
func injectName(s, name string) string {
	return strings.ReplaceAll(s, "{name}", name)
}
