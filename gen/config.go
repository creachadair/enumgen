package gen

import (
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

// LoadConfig reads and parses a YAML configuration from path.
//
// If the filename of path ends in ".go", it calls ConfigFromSource on the
// file; otherwise the file must be a standalone YAML file and is parsed by
// ConfigFromFile.
func LoadConfig(path string) (*Config, error) {
	if filepath.Ext(path) == ".go" {
		return ConfigFromSource(path)
	}
	return ConfigFromFile(path)
}

// ConfigFromFile reads and parses the YAML config file specified by path.
func ConfigFromFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseConfig(f)
}

// ConfigFromSource reads and parses the Go file specified by path, and
// extracts a YAML config from the first comment block tagged enumgen:config
// found in the file.  An error results if no such comment is found.
func ConfigFromSource(path string) (*Config, error) {
	const flags = parser.ParseComments | parser.SkipObjectResolution
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, flags)
	if err != nil {
		return nil, err
	}
	var configText []string
	for _, cg := range f.Comments {
		first := cg.List[0] // guaranteed to exist

		if strings.HasPrefix(first.Text, "/*enumgen:config") {
			// Found a tagged comment group beginning with a block comment.
			clean := strings.TrimSpace(strings.TrimSuffix(first.Text, "*/"))
			lines := strings.Split(clean, "\n")
			configText = append(configText, lines[1:]...) // discard the tag
		} else if !strings.HasPrefix(first.Text, "//enumgen:config") {
			continue
		}

		// Run through the rest of the group accumulating comments.  Skip the
		// first one, which was either the tag comment, or has already been added
		// to the collection.
		for _, com := range cg.List[1:] {
			if strings.HasPrefix(com.Text, "//") {
				configText = append(configText, strings.TrimPrefix(com.Text, "// "))
			} else {
				clean := strings.TrimSuffix(strings.TrimPrefix(com.Text, "/*"), "*/")
				configText = append(configText, strings.TrimSpace(clean))
			}
		}
		break
	}
	if len(configText) != 0 {
		input := strings.NewReader(strings.Join(configText, "\n"))
		cfg, err := ParseConfig(input)
		if err == nil && cfg.Package == "" {
			cfg.Package = f.Name.Name
		}
		return cfg, err
	}
	return nil, fmt.Errorf("no config comment found in %q", path)
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
		if zero := e.Prefix + e.Zero; zero != "" {
			if valueSeen[zero] != "" {
				return fmt.Errorf("enum %q default %q duplicated in %q",
					e.Type, zero, valueSeen[zero])
			}
			valueSeen[zero] = e.Type
		}
		for j, v := range e.Values {
			if v.Name == "" {
				return fmt.Errorf("enum %q value %d: name not defined", e.Type, j+1)
			} else if e.Zero != "" && v.Name == e.Zero {
				return fmt.Errorf("enum %q value %d: name %q conflicts with default", e.Type, j+1, v.Name)
			} else if full := e.Prefix + v.Name; valueSeen[full] != "" {
				return fmt.Errorf("enum %q value %d: name %q duplicated in %q",
					e.Type, j+1, full, valueSeen[full])
			} else {
				valueSeen[full] = e.Type
			}
		}
	}
	return nil
}
