package gen

import (
	"errors"
	"fmt"
	"io"
	"os"

	yaml "gopkg.in/yaml.v3"
)

// LoadConfig reads and parses a YAML configuration from path.
func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseConfig(f)
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
	enumSeen := make(map[string]bool)
	valueSeen := make(map[string]string)
	for i, e := range c.Enum {
		if e.Type == "" {
			return fmt.Errorf("enum %d: type name not defined", i+1)
		} else if enumSeen[e.Type] {
			return fmt.Errorf("enum %d: duplicate type name %q", i+1, e.Type)
		}
		enumSeen[e.Type] = true
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
