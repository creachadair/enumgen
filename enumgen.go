// Program enumgen generates Go enumeration types. It is intended for use with
// the "go generate" tool.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/creachadair/enumgen/gen"
)

var (
	configPath = flag.String("config", "", "Configuration file path (required)")
	outputPath = flag.String("output", "", "Output file path (required)")
)

func main() {
	flag.Parse()
	switch {
	case *configPath == "":
		log.Fatal("You must specify a -config file path")
	case *outputPath == "":
		log.Fatal("You must specify an -output file path")
	}
	cfg, err := gen.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Reading config: %v", err)
	}
	f, err := os.Create(*outputPath)
	if err != nil {
		log.Fatalf("Output: %v", err)
	}
	err = cfg.Generate(f)
	cerr := f.Close()
	if err != nil {
		log.Fatalf("Generate: %v", err)
	} else if cerr != nil {
		log.Fatalf("Close output: %v", err)
	}
}
