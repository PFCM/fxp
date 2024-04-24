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
		if err := genOps(*dirFlag); err != nil {
			log.Fatal(err)
		}
	}

	if *genTypesFlag {
		if err := genTypes(*dirFlag); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("All done")
}

func genOps(dir string) error {
	log.Println("Generating op benchmarks")

	path := filepath.Join(dir, "sat_bench_test.go")
	b, err := gen.GenOpsBenchmarks()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0666); err != nil {
		return err
	}
	log.Printf("Wrote %s", path)

	path = filepath.Join(dir, "sat_test.go")
	b, err = gen.GenOpsTests()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0666); err != nil {
		return err
	}
	log.Printf("Write %s", path)

	return nil
}

func genTypes(dir string) error {
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

		prefix := filepath.Join(*dirFlag, strings.ToLower(g.Typename()))
		path := prefix + ".go"
		src, err := g.Gen()
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, src, 0666); err != nil {
			return err
		}
		log.Printf("Wrote %q", path)

		testPath := prefix + "_test.go"
		src, err = g.GenTest()
		if err != nil {
			return err
		}
		if err := os.WriteFile(testPath, src, 0666); err != nil {
			return err
		}
		log.Printf("Wrote %q", testPath)
	}
	return nil
}
