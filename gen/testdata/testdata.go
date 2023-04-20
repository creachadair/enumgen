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

//enumgen:type Color
// doc: |
//   A Color is a source of joy for all who behold it.
// flag-value: true
// constructor: true
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
