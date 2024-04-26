// gen-fix uses fix/gen to generate most of the fix package.
package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pfcm/fxp/fix/gen"
)

var (
	dirFlag = flag.String("dir", "", "directory in which to write output")

	genOpsFlag   = flag.Bool("ops", true, "whether or not to generate tests/benchmarks for the hand-written saturating arithmetic in sat.go")
	genTypesFlag = flag.Bool("types", true, "whether or not to generate all of the various fixed-point types")
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	log.SetPrefix("gen-fix: ")

	if *genOpsFlag {
		if err := genOps(); err != nil {
			log.Fatal(err)
		}
	}

	if *genTypesFlag {
		if err := genTypes(); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("All done")
}

func genOps() error {
	log.Println("Generating op tests/benchmarks")
	if err := write("sat_bench_test.go", gen.GenOpsBenchmarks); err != nil {
		return err
	}
	return write("sat_test.go", gen.GenOpsTests)
}

func genTypes() error {
	type generator interface {
		Typename() string
		Gen() ([]byte, error)
		GenTest() ([]byte, error)
	}
	var types []generator
	for _, u := range gen.Unsigneds {
		types = append(types, u)
	}
	for _, s := range gen.Signeds {
		types = append(types, s)
	}

	for _, g := range types {
		log.Printf("Generating %v", g.Typename())
		name := strings.ToLower(g.Typename())

		implName := name + ".go"
		if err := write(implName, g.Gen); err != nil {
			return err
		}

		testName := name + "_test.go"
		if err := write(testName, g.GenTest); err != nil {
			return err
		}
	}
	return write("pairs.go", gen.GenPairs)
}

func write(filename string, f func() ([]byte, error)) error {
	b, err := f()
	if err != nil {
		return err
	}
	path := filepath.Join(*dirFlag, filename)
	if err := os.WriteFile(path, b, 0666); err != nil {
		return err
	}
	log.Printf("Wrote %q", path)
	return nil
}
