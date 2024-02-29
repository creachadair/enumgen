// Program enumgen generates Go enumeration types. It is intended for use with
// the "go generate" tool.
//
// The generator reads a configuration file in YAML format (see gen.Config).
// To generate types, add:
//
//	//go:generate -command enumgen go run github.com/creachadair/enumgen@latest
//	//go:generate enumgen -config enums.yml -output generated.go
package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/creachadair/enumgen/gen"
)

var (
	configPath = flag.String("config", "", "Configuration file path")
	outputPath = flag.String("output", "", "Output file path (required)")
)

func main() {
	flag.Parse()
	if *outputPath == "" {
		log.Fatal("You must specify an -output file path")
	}

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Reading config: %v", err)
	}
	f, err := os.Create(*outputPath)
	if err != nil {
		log.Fatalf("Output: %v", err)
	}
	log.Printf("Generating %d enumerations for package %q", len(cfg.Enum), cfg.Package)
	if err := errors.Join(cfg.Generate(f), f.Close()); err != nil {
		log.Fatalf("Generate: %v", err)
	}
}

func loadConfig() (*gen.Config, error) {
	if *configPath == "" {
		log.Print("Loading configuration from package source")
		return gen.LoadPackage()
	} else if strings.HasSuffix(*configPath, ".go") {
		return gen.ConfigFromGoFile(*configPath)
	}
	return gen.ConfigFromYAML(*configPath)
}
