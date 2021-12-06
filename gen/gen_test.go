package gen_test

import (
	"bytes"
	"crypto/sha256"
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
	for _, path := range []string{"gen.go", "testdata/gentest.yml"} {
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

	t.Run("E3", func(t *testing.T) {
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
		{`"X" conflicts with default`, &gen.Config{
			Package: "foo",
			Enum: []*gen.Enum{
				{Type: "bar", Zero: "X", Values: []*gen.Value{{Name: "X"}}},
			},
		}},
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
