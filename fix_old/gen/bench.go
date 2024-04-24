package gen

import (
	"fmt"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/transform"
)

var ops = []op{{
	Name:      "usadd",
	InputType: "uint8",
	Implementations: []string{
		"usadd",
		"usaddpre",
		"usaddpost",
		"usaddpostbranchless",
	},
	Canon: `min(a+b, 255)`,
}, {
	Name:      "ussub",
	InputType: "uint8",
	Implementations: []string{
		"ussub",
		"ussubbranch",
		"ussubbranchless",
		"ussubmin",
	},
	Canon: `max(a-b, 0)`,
}, {
	Name:      "usmul",
	InputType: "uint8",
	FracInput: true,
	Implementations: []string{
		"usmul",
		"usmulbig",
		"usmulbigmin",
		"usmulbigbranchless",
		"usmulbits",
		// TODO: table based?
	},
	Canon: `min((a*b) >> f, 255)`,
}, {
	Name:      "ssadd",
	InputType: "int8",
	Implementations: []string{
		"ssadd",
		"ssaddbranch",
		"ssaddbranchless",
		"ssaddbig",
	},
	Canon: `min(max(a+b, -128), 127)`,
}, {
	Name:      "sssub",
	InputType: "int8",
	Implementations: []string{
		"sssub",
		"sssubbig",
		"sssubadd",
		"sssubaddlessbranch",
		"sssubdirect",
	},
	Canon: `min(max(a-b, -128), 127)`,
}, {
	Name:      "ssmul",
	InputType: "int8",
	FracInput: true,
	Implementations: []string{
		"ssmul",
		"ssmulbig",
		"ssmulbigbranch",
	},
	Canon: `min(max((a*b) >> f, -128), 127)`,
}}

type op struct {
	Name            string
	InputType       string // "uint8" or "int8"
	FracInput       bool
	Implementations []string
	Canon           string
}

func (o op) MaxName() int {
	var l int
	for _, i := range o.Implementations {
		l = max(l, len(i))
	}
	return l
}

func GenOpsBenchmarks() ([]byte, error) {
	return execAndFmt(benchTmpl, ops)
}

func GenOpsTests() ([]byte, error) {
	return execAndFmt(testTmpl, ops)
}

var benchTmpl = template.Must(template.New("benchmarks").Funcs(template.FuncMap{
	"benchname": func(n string) (string, error) {
		s, _, err := transform.String(cases.Title(language.English), n)
		if err != nil {
			return "", err
		}
		return "Benchmark" + s, nil
	},
	"padto": func(s string, n int) string {
		// :)
		return strings.ReplaceAll(fmt.Sprintf("%-*s", n, s), " ", "_")
	},
}).Parse(`
// Code generated by by github.com/pfcm/fxp/fix/gen DO NOT EDIT.

package fix

import (
	"math/rand"
	"testing"
)

{{range $op := .}}
func {{benchname .Name}}(b *testing.B) {
	b.Run("random", func(b *testing.B) {
		const n = 1 << 16
		var as, bs [n]{{.InputType}}
		{{if .FracInput -}}
		var fs [n]uint8
		{{end -}}
		for i := range as {
			as[i] = {{.InputType}}(i)
			bs[i] = {{.InputType}}(i)
			{{if .FracInput}}fs[i] = uint8(i){{end -}}
		}
		shuffleSlice(as[:])
		shuffleSlice(bs[:])
		{{if .FracInput -}}
		shuffleSlice(fs[:])
		{{end -}}
		{{range .Implementations -}}
		b.Run("{{padto . $op.MaxName}}", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				j := i%n
				x := {{.}}(as[j], bs[j]{{if $op.FracInput}}, fs[j]{{end}})
				_ = x
			}
		})
		{{end}}
	})
	b.Run("small", func(b *testing.B) {
		{{range .Implementations -}}
		b.Run("{{padto . $op.MaxName}}", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				x := {{.}}(1, 2{{if $op.FracInput}}, 3{{end}})
				_ = x
			}
		})
		{{end}}
	})
	b.Run("big", func(b *testing.B) {
		{{range .Implementations -}}
		b.Run("{{padto . $op.MaxName}}", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				x := {{.}}(126, 127{{if $op.FracInput}}, 6{{end}})
				_ = x
			}
		})
		{{end}}
	})
}
{{end}}

func shuffleSlice[T any](ts []T) {
	rand.Shuffle(len(ts), func(i, j int) { ts[i], ts[j] = ts[j], ts[i] })
}
`))

var testTmpl = template.Must(template.New("tests").Funcs(template.FuncMap{
	"testname": func(n string) (string, error) {
		s, _, err := transform.String(cases.Title(language.English), n)
		if err != nil {
			return "", err
		}
		return "Test" + s, nil
	},
	"loop": func(t, v string) (string, error) {
		var start, end int
		switch t {
		case "uint8":
			start = 0
			end = 256
		case "int8":
			start = -128
			end = 127
		default:
			return "", fmt.Errorf("unknown type for loop: %q", t)
		}
		return fmt.Sprintf("for %s := %d; %s < %d; %s++", v, start, v, end, v), nil
	},
}).Parse(`
// Code generated by by github.com/pfcm/fxp/fix/gen DO NOT EDIT.

package fix

import (
	"testing"
)

{{range $op := .}}
func {{testname .Name}}(t *testing.T) {
	{{range .Implementations -}}
	t.Run("{{.}}", func(t *testing.T) {
	{{loop $op.InputType "a"}} {
		{{loop $op.InputType "b"}} {
		{{if $op.FracInput -}}
		 for f := 0; f < 8; f++ {
		{{end -}}
			want := {{$op.Canon}}
			got := {{.}}({{$op.InputType}}(a), {{$op.InputType}}(b){{if $op.FracInput}}, uint8(f){{end}})
			if {{$op.InputType}}(want) != got {
				t.Errorf("{{.}}(%x, %x{{if $op.FracInput}}, %x{{end}}) = %x, want: %x", {{$op.InputType}}(a), {{$op.InputType}}(b){{if $op.FracInput}}, f{{end}}, got, want)
			}
		{{if $op.FracInput}}}{{end -}}
		}
	}
	})
	{{end}}
}
{{end}}
`))
