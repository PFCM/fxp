// show-fix shows the representations of fixed point numbers, mostly for
// debugging conversions etc.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

var (
	typesFlag = flag.String("types", "", "comma separated list of `type names` to show. Leave empty to show all types")
	opsFlag   = flag.String("ops", "", "comma separated list of `operations` to show. Available operations are :"+strings.Join(opKeys, ", ")+". Defaults to all operations")
)

func main() {
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), help)
		fmt.Fprintln(flag.CommandLine.Output(), "\nOptional arguments:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if n := flag.NArg(); n < 1 || n > 2 {
		fail("Need exactly one or two arguments.")
	}

	types, err := parseTypes(*typesFlag)
	if err != nil {
		fail(err.Error())
	}
	ops, err := parseOps(*opsFlag)
	if err != nil {
		fail(err.Error())
	}

	a, err := parse(flag.Arg(0))
	if err != nil {
		fail(err.Error())
	}
	// Parse as 9 bits to cover the range of both signed and unsigned.
	w := tabwriter.NewWriter(os.Stdout, 11, 1, 1, ' ', 0)

	showConversions(w, types, a)

	if flag.NArg() == 2 {
		b, err := parse(flag.Arg(1))
		if err != nil {
			fail(err.Error())
		}
		fmt.Fprintln(w)
		showConversions(w, types, b)
		fmt.Fprintln(w)
		showOps(w, types, ops, a, b)
	}

	if err := w.Flush(); err != nil {
		fail(err.Error())
	}
}

func parseTypes(ts string) (map[string]bool, error) {
	all := make(map[string]bool)
	for _, t := range typeKeys {
		all[t] = true
	}
	if ts == "" {
		return all, nil
	}
	result := make(map[string]bool)
	for _, t := range strings.Split(ts, ",") {
		if !all[t] {
			return nil, fmt.Errorf("unknown type %q", t)
		}
		result[t] = true
	}
	return result, nil
}

func parseOps(os string) (map[string]bool, error) {
	all := make(map[string]bool)
	for _, o := range opKeys {
		all[o] = true
	}
	if os == "" {
		return all, nil
	}
	result := make(map[string]bool)
	for _, o := range strings.Split(os, ",") {
		if !all[o] {
			return nil, fmt.Errorf("unknown op %q", o)
		}
		result[o] = true
	}
	return result, nil
}

func parse(s string) (int64, error) {
	raw, err := strconv.ParseInt(s, 0, 9)
	if err != nil {
		return 0, err
	}
	// Validate the range.
	if raw < -128 {
		return 0, fmt.Errorf("%d doesn't fit in 8 bits", raw)
	}
	return raw, nil
}

func showConversions(w io.Writer, types map[string]bool, i int64) {
	for _, t := range typeKeys {
		if !types[t] {
			continue
		}
		for _, f := range conversions[t] {
			f(w, i)
		}
	}
}

func showOps(w io.Writer, types, showOps map[string]bool, a, b int64) {
	for _, t := range typeKeys {
		if !types[t] {
			continue
		}
		for _, o := range opKeys {
			if !showOps[o] {
				continue
			}
			for _, f := range ops[t][o] {
				f(w, a, b)
			}
		}
	}
}

func fail(reason string) {
	fmt.Fprintln(os.Stderr, reason)
	fmt.Fprintln(os.Stderr, help)
	os.Exit(1)
}

const help = `show-fix shows various fixed-point representations of the same
bit pattern.
Usage:
	show-fix [-types] num [num]

Where num is an integer literal in Go syntax. If a second number is provided,
also shows the results of various operations between them.
`
