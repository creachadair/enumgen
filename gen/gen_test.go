package gen_test

import (
	"bytes"
	"crypto/sha256"
	"encoding"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/creachadair/enumgen/gen"
	"github.com/creachadair/enumgen/gen/testdata"
)

type enumType interface {
	Enum() string
	Index() int
	String() string
	Valid() bool
}

func check(t *testing.T, v enumType, ok bool, want string) {
	t.Helper()
	if got := v.Valid(); got != ok {
		t.Errorf("(%v).Valid: got %v, want %v", v, got, ok)
	}
	if got := v.String(); got != want {
		t.Errorf("(%v).String: got %q want %q", v, got, want)
	}
}

func checkGenerated(t *testing.T) {
	t.Helper()

	// Verify that the generator package and the testdata config match the hash
	// embedded in the generated test data.
	h := sha256.New()
	for _, path := range []string{"gen.go", "config.go", "testdata/gentest.yml", "testdata/testdata.go"} {
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("Reading input: %v", err)
		}
		defer f.Close()
		if _, err := io.Copy(h, f); err != nil {
			t.Fatalf("Hashing %q: %v", path, err)
		}
	}

	actual := fmt.Sprintf("%x", h.Sum(nil))
	if actual != testdata.GeneratorHash {
		t.Log(`-- WARNING
Either the code generator or the test data (or both) have changed.
To update the test data to match, run:
   go generate ./gen/testdata
`)
		t.Fatalf("Got hash %q, want %q", testdata.GeneratorHash, actual)
	}
}

func TestEnums(t *testing.T) {
	checkGenerated(t)

	t.Run("E1", func(t *testing.T) {
		var zero testdata.E1
		check(t, zero, false, "<invalid>")
		check(t, testdata.A, true, "alpha")
		check(t, testdata.B, true, "bravo")
		check(t, testdata.C, true, "C")
	})

	t.Run("E2", func(t *testing.T) {
		var zero testdata.E2
		check(t, zero, false, "<invalid>")
		check(t, testdata.E2_Invalid, false, "<invalid>")
		check(t, testdata.E2_A, true, "A")
		check(t, testdata.E2_B, true, "B")
	})

	t.Run("E4", func(t *testing.T) {
		var zero testdata.E4
		check(t, zero, false, "<invalid>")
		check(t, testdata.E4_P, true, "P")
		check(t, testdata.E4_D, true, "D")
		check(t, testdata.E4_Q, true, "Q")
	})

	t.Run("Count", func(t *testing.T) {
		var zero testdata.Count
		check(t, zero, false, "zilch")
		check(t, testdata.Zero, false, "zilch")
		check(t, testdata.One, true, "lonely")
		check(t, testdata.Two, true, "tango")
	})

	t.Run("E1Index", func(t *testing.T) {
		var zero testdata.E1
		for i, e := range []testdata.E1{zero, testdata.A, testdata.B, testdata.C} {
			if got := e.Index(); got != i {
				t.Errorf("Index for %v: got %d, want %d", e, got, i)
			}
		}
	})

	t.Run("E2Map", func(t *testing.T) {
		// Verify that enumerators work as map keys.
		m := map[testdata.E2]bool{
			testdata.E2_Invalid: true,
			testdata.E2_A:       true,
		}
		if !m[testdata.E2_Invalid] {
			t.Error("Invalid missing")
		}
		if !m[testdata.E2_A] {
			t.Error("A missing")
		}
		if m[testdata.E2_B] {
			t.Error("B found")
		}
	})

	t.Run("E3Flag", func(t *testing.T) {
		var target testdata.E3
		check(t, target, false, "<invalid>")

		var _ flag.Value = &target

		if err := target.Set("foo"); err != nil {
			t.Errorf("Set foo: %v", err)
		} else if target != testdata.X {
			t.Errorf("Set foo: got %v, want %v", target, testdata.X)
		}
		if err := target.Set("bar"); err != nil {
			t.Errorf("Set bar: %v", err)
		} else if target != testdata.Y {
			t.Errorf("Set bar: got %v, want %v", target, testdata.Y)
		}
		if err := target.Set("baz"); err == nil {
			t.Error("Set baz did not report an error")
		} else if target != testdata.Y {
			t.Errorf("After set baz: got %v, want %v", target, testdata.Y)
		}
	})

	t.Run("E3Text", func(t *testing.T) {
		var _ encoding.TextMarshaler = testdata.E3{}
		var _ encoding.TextUnmarshaler = (*testdata.E3)(nil)

		t.Run("Good", func(t *testing.T) {
			var target testdata.E3

			bits, err := json.Marshal(testdata.X)
			if err != nil {
				t.Fatalf("Marshal %v failed: %v", testdata.X, err)
			}

			if err := json.Unmarshal(bits, &target); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			want := testdata.X.String()
			if got := target.String(); got != want {
				t.Errorf("Decoded value: got %q, want %q", got, want)
			}
		})

		t.Run("Zero", func(t *testing.T) {
			var target testdata.E3

			// An empty string should decode to the zero enumerator.
			if err := json.Unmarshal([]byte(`""`), &target); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			if target.Valid() {
				t.Error("Decoded empty incorrectly reports valid")
			}

			// The <invalid> label should decode to the zero enumerator.
			bits, err := json.Marshal(testdata.E3{})
			if err != nil {
				t.Fatalf("Marshal %v failed: %v", testdata.E3{}, err)
			}
			if err := json.Unmarshal(bits, &target); err != nil {
				t.Fatalf("Unmarshal failed; %v", err)
			}
			if target.Valid() {
				t.Errorf("Decoded %s incorrectly reports valid", string(bits))
			}
		})

		t.Run("Bad", func(t *testing.T) {
			var target testdata.E3

			const bad = "nonesuch"
			if err := json.Unmarshal([]byte(`"`+bad+`"`), &target); err == nil {
				t.Errorf("Unmarshal: got %v, want error", target)
			} else {
				t.Logf("Decoding %q correctly failed: %v", bad, err)
			}
		})
	})

	t.Run("E3FromIndex", func(t *testing.T) {
		var zero testdata.E3
		tests := []struct {
			input int
			want  testdata.E3
		}{
			{0, zero},
			{1, testdata.X},
			{2, testdata.Y},
			{3, zero},
			{-1, zero},
		}
		for _, tc := range tests {
			if got := testdata.E3FromIndex(tc.input); got != tc.want {
				t.Errorf("E3FromIndex(%d): got %v, want %v", tc.input, got, tc.want)
			}
		}
	})

	t.Run("SizeFromIndex", func(t *testing.T) {
		var zero testdata.Size
		tests := []struct {
			input int
			want  testdata.Size
		}{
			{0, zero},
			{1, testdata.Small},
			{2, testdata.Medium},
			{3, zero},
			{4, testdata.Large},
			{5, zero},
			{10, testdata.XLarge},
			{50, zero},
			{-1, zero},
		}
		for _, tc := range tests {
			if got := testdata.SizeFromIndex(tc.input); got != tc.want {
				t.Errorf("E3FromIndex(%d): got %v, want %v", tc.input, got, tc.want)
			}
		}
	})

	t.Run("ColorFlag", func(t *testing.T) {
		const redText = "fire-engine-red"
		color := testdata.Red
		if got := color.String(); got != redText {
			t.Errorf("Red: got %q, want %q", got, redText)
		}
		var _ flag.Value = &color
	})
}

func TestErrors(t *testing.T) {
	tests := []struct {
		desc   string
		config *gen.Config
	}{
		{"package name not defined", &gen.Config{}},
		{"no enumerations defined", &gen.Config{Package: "foo"}},
		{"type name not defined", &gen.Config{
			Package: "foo",
			Enum:    []*gen.Enum{{}},
		}},
		{"no enumerators defined", &gen.Config{
			Package: "foo",
			Enum:    []*gen.Enum{{Type: "bar"}},
		}},
		{"name not defined", &gen.Config{
			Package: "foo",
			Enum: []*gen.Enum{{Type: "bar", Values: []*gen.Value{
				{},
			}}},
		}},

		// Check for enumerator duplication within an enum.
		{`name "baz" duplicated in "bar"`, &gen.Config{
			Package: "foo",
			Enum: []*gen.Enum{{Type: "bar", Values: []*gen.Value{
				{Name: "baz"}, {Name: "baz"},
			}}},
		}},

		// Check for duplicate enum names.
		{`duplicate type name "bar"`, &gen.Config{
			Package: "foo",
			Enum: []*gen.Enum{
				{Type: "bar", Values: []*gen.Value{{Name: "baz"}}},
				{Type: "bar", Values: []*gen.Value{{Name: "quux"}}},
			},
		}},

		// Check for enumerator duplication across enums.
		{`name "baz" duplicated in "bar"`, &gen.Config{
			Package: "foo",
			Enum: []*gen.Enum{
				{Type: "bar", Values: []*gen.Value{{Name: "baz"}}},
				{Type: "zut", Values: []*gen.Value{{Name: "baz"}}},
			},
		}},

		// Check that name collisions due to prefix addition are caught.
		{`name "AX" duplicated in "bar"`, &gen.Config{
			Package: "foo",
			Enum: []*gen.Enum{
				{Type: "bar", Prefix: "A", Values: []*gen.Value{{Name: "X"}}},
				{Type: "baz", Values: []*gen.Value{{Name: "AX"}}},
			},
		}},

		// Check for name collisions with default (zero) enumerators.
		{`name "X" duplicated in "bar"`, &gen.Config{
			Package: "foo",
			Enum: []*gen.Enum{
				{Type: "bar", Zero: "X", Values: []*gen.Value{{Name: "Y"}}},
				{Type: "baz", Values: []*gen.Value{{Name: "X"}}},
			},
		}},
		{`default "Y" duplicated in "bar"`, &gen.Config{
			Package: "foo",
			Enum: []*gen.Enum{
				{Type: "bar", Values: []*gen.Value{{Name: "Y"}}},
				{Type: "baz", Zero: "Y", Values: []*gen.Value{{Name: "X"}}},
			},
		}},
		{`default "Y" duplicated in "bar"`, &gen.Config{
			Package: "foo",
			Enum: []*gen.Enum{
				{Type: "bar", Zero: "Y", Values: []*gen.Value{{Name: "X"}}},
				{Type: "baz", Zero: "Y", Values: []*gen.Value{{Name: "Z"}}},
			},
		}},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			var buf bytes.Buffer
			err := test.config.Generate(&buf)
			if err == nil {
				t.Errorf("Test %s: expected error, got\n%s", test.desc, buf.String())
			} else if !strings.Contains(err.Error(), test.desc) {
				t.Errorf("Test %s: error does not match: %v", test.desc, err)
			}
		})
	}
}
