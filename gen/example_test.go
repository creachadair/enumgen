package gen_test

import (
	"log"
	"os"

	"github.com/creachadair/enumgen/gen"
)

const input = `package example
/*enumgen:type Example

# package example inferred from the declaration above.
# The type name assigned by the enumgen: comment.

doc: Example is an example enumeration.
values:
  - name: Good
    doc: upsides
  - name: Bad
    doc: downsides
  - name: Ugly
    doc: what it says on the tin
*/
`

func Example() {
	cfg, err := gen.ConfigFromSource("example.go", []byte(input))
	if err != nil {
		log.Fatalf("Parse: %v", err)
	}
	if err := cfg.Generate(os.Stdout); err != nil {
		log.Fatalf("Generate: %v", err)
	}
	// Output:
	// // Code generated by enumgen. DO NOT EDIT.
	//
	// package example
	//
	// // Example is an example enumeration.
	// type Example struct{ _Example uint8 }
	//
	// // Enum returns the name of the enumeration type for Example.
	// func (Example) Enum() string { return "Example" }
	//
	// // String returns the string representation of Example v.
	// func (v Example) String() string { return _str_Example[v._Example] }
	//
	// // Valid reports whether v is a valid non-zero Example value.
	// func (v Example) Valid() bool { return v._Example > 0 && int(v._Example) < len(_str_Example) }
	//
	// // Index returns the integer index of Example v.
	// func (v Example) Index() int { return int(v._Example) }
	//
	// var (
	// 	_str_Example = []string{"<invalid>", "Good", "Bad", "Ugly"}
	//
	// 	Good = Example{1} // upsides
	// 	Bad  = Example{2} // downsides
	// 	Ugly = Example{3} // what it says on the tin
	// )
}
