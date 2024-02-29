// Code generated by enumgen. DO NOT EDIT.

package testdata

import (
	"fmt"
	"strings"
)

// An enumeration defined in a Go file.
type E4 struct{ _E4 uint8 }

// Enum returns the name of the enumeration type for E4.
func (E4) Enum() string { return "E4" }

// String returns the string representation of E4 v.
func (v E4) String() string { return _str_E4[v._E4] }

// Valid reports whether v is a valid non-zero E4 value.
func (v E4) Valid() bool { return v._E4 > 0 && int(v._E4) < len(_str_E4) }

// Index returns the integer index of E4 v.
func (v E4) Index() int { return int(v._E4) }

var (
	_str_E4 = []string{"<invalid>", "P", "D", "Q"}

	E4_P = E4{1}
	E4_D = E4{2}
	E4_Q = E4{3}
)

// A Size denotes the size of a t-shirt.
type Size struct{ _Size uint8 }

// Enum returns the name of the enumeration type for Size.
func (Size) Enum() string { return "Size" }

// String returns the string representation of Size v.
func (v Size) String() string { return _str_Size[v._Size] }

// Valid reports whether v is a valid non-zero Size value.
func (v Size) Valid() bool { return v._Size > 0 && int(v._Size) < len(_str_Size) }

// Index returns the integer index of Size v.
func (v Size) Index() int { return _idx_Size[v._Size] }

var (
	_str_Size = []string{"<invalid>", "Small", "Medium", "Large", "XLarge"}
	_idx_Size = []int{0, 1, 2, 4, 10}

	Small  = Size{1}
	Medium = Size{2}
	Large  = Size{3}
	XLarge = Size{4}
)

// A Color is a source of joy for all who behold it.
type Color struct{ _Color uint8 }

// Enum returns the name of the enumeration type for Color.
func (Color) Enum() string { return "Color" }

// String returns the string representation of Color v.
func (v Color) String() string { return _str_Color[v._Color] }

// Valid reports whether v is a valid non-zero Color value.
func (v Color) Valid() bool { return v._Color > 0 && int(v._Color) < len(_str_Color) }

// Index returns the integer index of Color v.
func (v Color) Index() int { return int(v._Color) }

// NewColor returns the first enumerator of Color whose string is a
// case-insensitive match for s. If no enumerator matches, it returns the
// zero enumerator.
func NewColor(s string) Color {
	for i, opt := range _str_Color[1:] {
		if strings.EqualFold(opt, s) {
			return Color{uint8(i + 1)}
		}
	}
	return Color{0}
}

// Set implements part of the flag.Value interface for Color.
// A value must equal the string representation of an enumerator.
func (v *Color) Set(s string) error {
	if e := NewColor(s); e.Valid() {
		*v = e
		return nil
	}
	return fmt.Errorf("invalid value for Color: %q", s)
}

// The names of the colours supported here.
var (
	_str_Color = []string{"<invalid>", "fire-engine-red", "scummy-green", "azure-sky-blue"}

	Red   = Color{1} // Red is the colour of my true love's eyes.
	Green = Color{2} // Green is the colour of my true love's blood.
	Blue  = Color{3}
)
