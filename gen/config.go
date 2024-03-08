package gen

import (
	"bytes"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/creachadair/mds/mapset"
	yaml "gopkg.in/yaml.v3"
)

// LoadPackage reads and parses a combined YAML configuration from the Go files
// stored in the current working directory.
//
// An error is reported if the current working directory does not contain any
// Go source files with enumeration configurations in them, or if the files
// match multiple package names.
func LoadPackage() (*Config, error) {
	des, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}
	var cfg *Config
	for _, de := range des {
		if filepath.Ext(de.Name()) != ".go" || strings.HasSuffix(de.Name(), "_test.go") {
			continue
		}
		c, err := ConfigFromGoFile(de.Name())
		if errors.Is(err, errNoComment) {
			continue // OK, skip this file
		} else if err != nil {
			return nil, err
		}
		if cfg == nil {
			cfg = c
			continue
		} else if cfg.Package != c.Package {
			return nil, fmt.Errorf("file %q has package %q, want %q", de.Name(), c.Package, cfg.Package)
		}
		cfg.Enum = append(cfg.Enum, c.Enum...)
	}
	if cfg == nil || len(cfg.Enum) == 0 {
		return nil, errors.New("no matching .go files found")
	}
	return cfg, nil
}

// ConfigFromYAML reads and parses the YAML config file specified by path.
func ConfigFromYAML(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseConfig(f)
}

// ConfigFromGoFile reads and parses the Go file specified by path, and
// extracts a YAML config from each first comment block tagged enumgen:type
// found in the file.  An error results if no such comment is found.
func ConfigFromGoFile(path string) (*Config, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ConfigFromSource(path, src)
}

var errNoComment = errors.New("no config comment")

// ConfigFromSource parses a config from the text of a Go source file.
// The path is used to for diagnostics.
func ConfigFromSource(path string, text []byte) (*Config, error) {
	const flags = parser.ParseComments | parser.SkipObjectResolution
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, text, flags)
	if err != nil {
		return nil, err
	}

	type enumBlock struct {
		name string
		text []string
	}
	var enumBlocks []enumBlock
	for _, cg := range f.Comments {
		first := cg.List[0] // guaranteed to exist

		if rest, ok := strings.CutPrefix(first.Text, "/*enumgen:type"); ok {
			// Found a tagged comment group beginning with a block comment.
			name, rest, _ := strings.Cut(rest, "\n")
			enumBlocks = append(enumBlocks, enumBlock{
				name: strings.TrimSpace(name),
				text: []string{cleanMulti(rest)},
			})
		} else if rest, ok := strings.CutPrefix(first.Text, "//enumgen:type"); ok {
			enumBlocks = append(enumBlocks, enumBlock{
				name: strings.TrimSpace(rest), // must be validated later
			})
			// lines are filled below
		} else {
			continue // not a relevant comment block
		}

		// Run through the rest of the group accumulating comments.
		// Reaching this point, the latest block already has the name extracted.
		cur := &enumBlocks[len(enumBlocks)-1]
		for _, com := range cg.List[1:] {
			if rest, ok := strings.CutPrefix(com.Text, "//"); ok {
				cur.text = append(cur.text, cleanSingle(rest))
			} else if rest, ok := strings.CutPrefix(com.Text, "/*"); ok {
				cur.text = append(cur.text, cleanMulti(rest))
			}
		}
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package: %s\nenum:\n", f.Name.Name)
	for _, enum := range enumBlocks {
		fmt.Fprintf(&buf, "- type: %s\n", enum.name)
		fmt.Fprintln(&buf, indentLines("  ", enum.text))
	}
	if buf.Len() == 0 {
		return nil, fmt.Errorf("%w found in %q", errNoComment, path)
	}
	return ParseConfig(&buf)
}

// ParseConfig parses a YAML configuration text from r.
func ParseConfig(r io.Reader) (*Config, error) {
	dec := yaml.NewDecoder(r)
	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) checkValid() error {
	if c.Package == "" {
		return errors.New("package name not defined")
	}
	if len(c.Enum) == 0 {
		return errors.New("no enumerations defined")
	}
	enumSeen := mapset.New[string]()
	valueSeen := make(map[string]string)
	for i, e := range c.Enum {
		if e.Type == "" {
			return fmt.Errorf("enum %d: type name not defined", i+1)
		} else if enumSeen.Has(e.Type) {
			return fmt.Errorf("enum %d: duplicate type name %q", i+1, e.Type)
		}
		enumSeen.Add(e.Type)
		if len(e.Values) == 0 {
			return fmt.Errorf("enum %d: no enumerators defined", i+1)
		}
		if zero := e.Prefix + e.Zero; e.Zero != "" {
			if valueSeen[zero] != "" && valueSeen[zero] != e.Type {
				return fmt.Errorf("enum %q default %q duplicated in %q",
					e.Type, zero, valueSeen[zero])
			}
			valueSeen[zero] = e.Type
		}

		// It is OK for the zero enumerator to be duplicated in its own value
		// list, but other names must not be duplicated within the list. This map
		// keeps track of just the names in this group to prevent that.

		var thisName mapset.Set[string]
		for j, v := range e.Values {
			if v.Name == "" {
				return fmt.Errorf("enum %q value %d: name not defined", e.Type, j+1)
			} else if thisName.Has(v.Name) {
				return fmt.Errorf("enum %q value %d: name %q duplicated in %q", e.Type, j+1, v.Name, e.Type)
			}
			thisName.Add(v.Name)

			full := e.Prefix + v.Name
			if valueSeen[full] != "" {
				// If this enumerator is "my" zero value, it's OK to repeat it in
				// the values list to provide text and documentation.
				if valueSeen[full] != e.Type || e.Zero == "" || e.Zero != v.Name {
					return fmt.Errorf("enum %q value %d: name %q duplicated in %q",
						e.Type, j+1, full, valueSeen[full])
				}
			}
			valueSeen[full] = e.Type
		}
	}
	return nil
}

func indentLines(pfx string, text []string) string {
	var lines []string
	for _, t := range text {
		lines = append(lines, strings.Split(strings.TrimSuffix(t, "\n"), "\n")...)
	}
	for i := range lines {
		lines[i] = pfx + lines[i]
	}
	return strings.Join(lines, "\n")
}

func cleanSingle(s string) string {
	return strings.TrimSuffix(strings.TrimPrefix(s, " "), "\n")
}

func cleanMulti(s string) string {
	return strings.TrimSpace(strings.TrimSuffix(s, "*/"))
}
