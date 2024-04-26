// show-fix shows the representations of fixed point numbers, mostly for
// debugging conversions etc.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/pfcm/fxp/fix"
	"golang.org/x/exp/maps"
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
	for _, t := range conversionKeys {
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

var conversions = map[string][]func(w io.Writer, i int64){
	"U08": []func(w io.Writer, i int64){
		func(w io.Writer, i int64) {
			showUnsigned(w, "U08", fix.U08(i))
		},
		func(w io.Writer, i int64) {
			showUnsigned(w, "U08ToU17", fix.U08ToU17(fix.U08(i)))
		},
		func(w io.Writer, i int64) {
			showUnsigned(w, "U08ToU71", fix.U08ToU71(fix.U08(i)))
		},
		func(w io.Writer, i int64) {
			showSigned(w, "U08ToS17", fix.U08ToS17(fix.U08(i)))
		},
	},
	"U17": []func(w io.Writer, i int64){
		func(w io.Writer, i int64) {
			showUnsigned(w, "U17", fix.U17(i))
		},
		func(w io.Writer, i int64) {
			showUnsigned(w, "U17ToU08", fix.U17ToU08(fix.U17(i)))
		},
		func(w io.Writer, i int64) {
			showUnsigned(w, "U17ToU71", fix.U17ToU71(fix.U17(i)))
		},
		func(w io.Writer, i int64) {
			showSigned(w, "U17ToS17", fix.U17ToS17(fix.U17(i)))
		},
	},
	"U71": []func(w io.Writer, i int64){
		func(w io.Writer, i int64) {
			showUnsigned(w, "U71", fix.U71(i))
		},
		func(w io.Writer, i int64) {
			showUnsigned(w, "U71ToU08", fix.U71ToU08(fix.U71(i)))
		},
		func(w io.Writer, i int64) {
			showUnsigned(w, "U71ToU17", fix.U71ToU17(fix.U71(i)))
		},
		func(w io.Writer, i int64) {
			showSigned(w, "U71ToS17", fix.U71ToS17(fix.U71(i)))
		},
	},
	"S17": []func(w io.Writer, i int64){
		func(w io.Writer, i int64) {
			showSigned(w, "S17", fix.S17(i))
		},
		func(w io.Writer, i int64) {
			showUnsigned(w, "S17ToU08", fix.S17ToU08(fix.S17(i)))
		},
		func(w io.Writer, i int64) {
			showUnsigned(w, "S17ToU17", fix.S17ToU17(fix.S17(i)))
		},
		func(w io.Writer, i int64) {
			showUnsigned(w, "S17ToU17", fix.S17ToU17(fix.S17(i)))
		},
	},
}

var conversionKeys = func() []string {
	keys := maps.Keys(conversions)
	sort.Strings(keys)
	return keys
}()

func showConversions(w io.Writer, types map[string]bool, i int64) {
	for _, t := range conversionKeys {
		if !types[t] {
			continue
		}
		for _, f := range conversions[t] {
			f(w, i)
		}
	}
}

var ops = map[string]map[string][]func(io.Writer, int64, int64){
	"U08": {
		"SAdd": []func(io.Writer, int64, int64){
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U08 + U08",
					fix.U08(i).SAdd(fix.U08(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U08 + U17",
					fix.U08(i).SAddU17(fix.U17(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U08 + U71",
					fix.U08(i).SAddU71(fix.U71(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U08 + S17",
					fix.U08(i).SAddS17(fix.S17(j)))
			},
		},
		"SSub": []func(io.Writer, int64, int64){
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U08 - U08",
					fix.U08(i).SSub(fix.U08(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U08 - U17",
					fix.U08(i).SSubU17(fix.U17(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U08 - U71",
					fix.U08(i).SSubU71(fix.U71(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U08 - S17",
					fix.U08(i).SSubS17(fix.S17(j)))
			},
		},
		"SMul": []func(io.Writer, int64, int64){
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U08 * U08",
					fix.U08(i).SMul(fix.U08(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U08 * U17",
					fix.U08(i).SMulU17(fix.U17(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U08 * U71",
					fix.U08(i).SMulU71(fix.U71(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U08 * S17",
					fix.U08(i).SMulS17(fix.S17(j)))
			},
		},
	},
	"U17": {
		"SAdd": []func(io.Writer, int64, int64){
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U17 + U08",
					fix.U17(i).SAddU08(fix.U08(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U17 + U17",
					fix.U17(i).SAdd(fix.U17(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U17 + U71",
					fix.U17(i).SAddU71(fix.U71(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U17 + S17",
					fix.U17(i).SAddS17(fix.S17(j)))
			},
		},
		"SSub": []func(io.Writer, int64, int64){
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U17 - U08",
					fix.U17(i).SSubU08(fix.U08(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U17 - U17",
					fix.U17(i).SSub(fix.U17(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U17 - U71",
					fix.U17(i).SSubU71(fix.U71(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U17 - S17",
					fix.U17(i).SSubS17(fix.S17(j)))
			},
		},
		"SMul": []func(io.Writer, int64, int64){
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U17 * U08",
					fix.U17(i).SMulU08(fix.U08(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U17 * U17",
					fix.U17(i).SMul(fix.U17(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U17 * U71",
					fix.U17(i).SMulU71(fix.U71(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U17 * S17",
					fix.U17(i).SMulS17(fix.S17(j)))
			},
		},
	},
	"U71": {
		"SAdd": []func(io.Writer, int64, int64){
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U71 + U08",
					fix.U71(i).SAddU08(fix.U08(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U71 + U17",
					fix.U71(i).SAddU17(fix.U17(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U71 + U71",
					fix.U71(i).SAdd(fix.U71(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U71 + S17",
					fix.U71(i).SAddS17(fix.S17(j)))
			},
		},
		"SSub": []func(io.Writer, int64, int64){
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U71 - U08",
					fix.U71(i).SSubU08(fix.U08(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U71 - U17",
					fix.U71(i).SSubU17(fix.U17(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U71 - U71",
					fix.U71(i).SSub(fix.U71(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U71 - S17",
					fix.U71(i).SSubS17(fix.S17(j)))
			},
		},
		"SMul": []func(io.Writer, int64, int64){
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U71 * U08",
					fix.U71(i).SMulU08(fix.U08(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U71 * U17",
					fix.U71(i).SMulU17(fix.U17(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U71 * U71",
					fix.U71(i).SMul(fix.U71(j)))
			},
			func(w io.Writer, i, j int64) {
				showUnsigned(w, "U71 * S17",
					fix.U71(i).SMulS17(fix.S17(j)))
			},
		},
	},
	"S17": {
		"SAdd": []func(io.Writer, int64, int64){
			func(w io.Writer, i, j int64) {
				showSigned(w, "S17 + U08",
					fix.S17(i).SAddU08(fix.U08(j)))
			},
			func(w io.Writer, i, j int64) {
				showSigned(w, "S17 + U17",
					fix.S17(i).SAddU17(fix.U17(j)))
			},
			func(w io.Writer, i, j int64) {
				showSigned(w, "S17 + U71",
					fix.S17(i).SAddU71(fix.U71(j)))
			},
			func(w io.Writer, i, j int64) {
				showSigned(w, "S17 + S17",
					fix.S17(i).SAdd(fix.S17(j)))
			},
		},
		"SSub": []func(io.Writer, int64, int64){
			func(w io.Writer, i, j int64) {
				showSigned(w, "S17 - U08",
					fix.S17(i).SSubU08(fix.U08(j)))
			},
			func(w io.Writer, i, j int64) {
				showSigned(w, "S17 - U17",
					fix.S17(i).SSubU17(fix.U17(j)))
			},
			func(w io.Writer, i, j int64) {
				showSigned(w, "S17 - U71",
					fix.S17(i).SSubU71(fix.U71(j)))
			},
			func(w io.Writer, i, j int64) {
				showSigned(w, "S17 - S17",
					fix.S17(i).SSub(fix.S17(j)))
			},
		},
		"SMul": []func(io.Writer, int64, int64){
			func(w io.Writer, i, j int64) {
				showSigned(w, "S17 * U08",
					fix.S17(i).SMulU08(fix.U08(j)))
			},
			func(w io.Writer, i, j int64) {
				showSigned(w, "S17 * U17",
					fix.S17(i).SMulU17(fix.U17(j)))
			},
			func(w io.Writer, i, j int64) {
				showSigned(w, "S17 * U71",
					fix.S17(i).SMulU71(fix.U71(j)))
			},
			func(w io.Writer, i, j int64) {
				showSigned(w, "S17 * S17",
					fix.S17(i).SMul(fix.S17(j)))
			},
		},
	},
}

var opKeys = []string{"SAdd", "SSub", "SMul"}

func showOps(w io.Writer, types, showOps map[string]bool, a, b int64) {
	for _, t := range conversionKeys {
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

func showUnsigned[U ~uint8](w io.Writer, name string, u U) {
	fmt.Fprintf(w, "%s:\t%v\t%d\t0x%02x\t0b%08b\n", name, u, u, uint8(u), u)
}

func showSigned[S ~int8](w io.Writer, name string, s S) {
	sign := ""
	if s < 0 {
		sign = "-"
	}
	fmt.Fprintf(w, "%s:\t%v\t%d\t%s0x%02x\t0b%08b\n",
		name, s, s, sign, abs(int(s)), uint8(s))
}

func abs(i int) int { return max(i, -i) }

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
