package testdata

//go:generate ./gen.sh

/*enumgen:type E4

# package testdata inferred from the declaration above.
# type name assigned by the enumgen: comment.
doc: An enumeration defined in a Go file.
prefix: E4_
values:
  - name: P
  - name: D
  - name: Q
*/

/*enumgen:type E5
doc: Another enumeration defined in a Go file.
lowercase-text: true
zero: Unfruited
values:
 - name: Unfruited
   text: fruitless  # not modified
 - name: Apple
 - name: Pear
 - name: Plum
 - name: Cherry
   text: UNMODIFIED # not lowercased
*/

/*enumgen:type Size

doc: "A {name} denotes the size of a t-shirt."
from-index: true
values:
  - name: Small
    index: 1

  - name: Medium  # index is 2

  - name: Large
    index: 4

  - name: XLarge
    index: 10
*/

//enumgen:type Color
// doc: |
//   A Color is a source of joy for all who behold it.
// flag-value: true
// constructor: true
// val-doc: The names of the colours supported here.
// values:
//   - name: Red
//     doc: "{name} is the colour of my true love's eyes."
//     text: fire-engine-red
//
//   - name: Green
//     doc: "{name} is the colour of my true love's blood."
//     text: scummy-green
//
//   - name: Blue
//     text: azure-sky-blue
