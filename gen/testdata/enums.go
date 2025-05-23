// Code generated by enumgen. DO NOT EDIT.

package testdata

import (
	"fmt"
	"strings"
)

type E1 struct{ _E1 uint8 }

// Enum returns the name of the enumeration type for E1.
func (E1) Enum() string { return "E1" }

// String returns the string representation of E1 v.
func (v E1) String() string { return _str_E1[v._E1] }

// Valid reports whether v is a valid non-zero E1 value.
func (v E1) Valid() bool { return v._E1 > 0 && int(v._E1) < len(_str_E1) }

// Index returns the integer index of E1 v.
func (v E1) Index() int { return int(v._E1) }

var (
	_str_E1 = []string{"<invalid>", "alpha", "bravo", "C"}

	A = E1{1}
	B = E1{2}
	C = E1{3}
)

type E2 struct{ _E2 uint8 }

// Enum returns the name of the enumeration type for E2.
func (E2) Enum() string { return "E2" }

// String returns the string representation of E2 v.
func (v E2) String() string { return _str_E2[v._E2] }

// Valid reports whether v is a valid non-zero E2 value.
func (v E2) Valid() bool { return v._E2 > 0 && int(v._E2) < len(_str_E2) }

// Index returns the integer index of E2 v.
func (v E2) Index() int { return int(v._E2) }

var (
	_str_E2 = []string{"<invalid>", "A", "B"}

	E2_Invalid = E2{0}
	E2_A       = E2{1}
	E2_B       = E2{2}
)

type E3 struct{ _E3 uint8 }

// Enum returns the name of the enumeration type for E3.
func (E3) Enum() string { return "E3" }

// String returns the string representation of E3 v.
func (v E3) String() string { return _str_E3[v._E3] }

// Valid reports whether v is a valid non-zero E3 value.
func (v E3) Valid() bool { return v._E3 > 0 && int(v._E3) < len(_str_E3) }

// Index returns the integer index of E3 v.
func (v E3) Index() int { return int(v._E3) }

// newE3 returns the first enumerator of E3 whose string is a
// case-insensitive match for s. If no enumerator matches, it returns the
// zero enumerator.
func newE3(s string) E3 {
	for i, opt := range _str_E3[1:] {
		if strings.EqualFold(opt, s) {
			return E3{uint8(i + 1)}
		}
	}
	return E3{0}
}

// E3FromIndex returns the first enumerator of E3 whose index equals v.
// If no enumerator matches, it returns the zero enumerator.
func E3FromIndex(v int) E3 {
	var zero E3
	if v <= 0 || v >= len(_str_E3) {
		return zero
	}
	return E3{uint8(v)}
}

// Set implements part of the flag.Value interface for E3.
// A value must equal the string representation of an enumerator.
func (v *E3) Set(s string) error {
	if e := newE3(s); e.Valid() {
		*v = e
		return nil
	}
	return fmt.Errorf("invalid value for E3: %q", s)
}

// MarshalText encodes the value of the E3 enumerator as text.
// It satisfies the encoding.TextMarshaler interface.
func (v E3) MarshalText() ([]byte, error) { return []byte(v.String()), nil }

// UnarshalText decodes the value of the E3 enumerator from a string.
// It reports an error if data does not encode a known enumerator.
// An empty slice decodes to the zero value.
// This method satisfies the encoding.TextUnmarshaler interface.
func (v *E3) UnmarshalText(data []byte) error {
	*v = E3{}
	text := string(data)
	if text == "" || text == _str_E3[0] {
		return nil
	}
	for i, opt := range _str_E3[1:] {
		if opt == text {
			v._E3 = uint8(i + 1)
			return nil
		}
	}
	return fmt.Errorf("invalid value for E3: %q", text)
}

var (
	_str_E3 = []string{"<invalid>", "foo", "bar"}

	X = E3{1}
	Y = E3{2}
)

type Count struct{ _Count uint8 }

// Enum returns the name of the enumeration type for Count.
func (Count) Enum() string { return "Count" }

// String returns the string representation of Count v.
func (v Count) String() string { return _str_Count[v._Count] }

// Valid reports whether v is a valid non-zero Count value.
func (v Count) Valid() bool { return v._Count > 0 && int(v._Count) < len(_str_Count) }

// Index returns the integer index of Count v.
func (v Count) Index() int { return int(v._Count) }

var (
	_str_Count = []string{"zilch", "lonely", "tango"}

	Zero = Count{0} // Nothing to see here
	One  = Count{1} // The very loneliest
	Two  = Count{2}
)

// GeneratorHash is used by the tests to verify that the testdata
// package is updated when the code generator changes.
const GeneratorHash = "080e5622e6545d8782c5cb2e16072467dd74296780714fd2499d8bc1f82c611a"
