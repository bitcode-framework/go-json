package cli

import (
	"flag"
	"reflect"
	"testing"
)

func TestReorderArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "flags already first",
			args: []string{"--input", "{}", "program.json"},
			want: []string{"--input", "{}", "program.json"},
		},
		{
			name: "flags after the path",
			args: []string{"program.json", "--input", "{}"},
			want: []string{"--input", "{}", "program.json"},
		},
		{
			name: "multiple flags after the path",
			args: []string{"program.json", "--input", "{}", "--timeout", "5s"},
			want: []string{"--input", "{}", "--timeout", "5s", "program.json"},
		},
		{
			name: "attached value is not split",
			args: []string{"program.json", "--input={}"},
			want: []string{"--input={}", "program.json"},
		},
		{
			name: "bool flag takes no value",
			args: []string{"program.json", "--trace"},
			want: []string{"--trace", "program.json"},
		},
		{
			name: "bool flag followed by a path",
			args: []string{"--trace", "program.json"},
			want: []string{"--trace", "program.json"},
		},
		{
			name: "single dash value",
			args: []string{"--input-file", "-", "program.json"},
			want: []string{"--input-file", "-", "program.json"},
		},
		{
			name: "terminator stops flag parsing",
			args: []string{"--trace", "--", "--not-a-flag"},
			want: []string{"--trace", "--not-a-flag"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.String("input", "", "")
			fs.String("input-file", "", "")
			fs.String("timeout", "", "")
			fs.Bool("trace", false, "")

			got := reorderArgs(fs, tc.args)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("reorderArgs(%q) = %q, want %q", tc.args, got, tc.want)
			}
		})
	}
}

// TestReorderArgs_FlagAfterPathParses is the regression guard for the bug where
// `go-json run program.json --input '{...}'` silently ignored --input: Go's
// flag package stops at the first positional argument.
func TestReorderArgs_FlagAfterPathParses(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	input := fs.String("input", "", "")
	fs.Bool("trace", false, "")

	parseFlags(fs, []string{"program.json", "--input", `{"name":"Alice"}`, "--trace"})

	if *input != `{"name":"Alice"}` {
		t.Errorf("input = %q, want the value written after the path", *input)
	}
	if fs.NArg() != 1 || fs.Arg(0) != "program.json" {
		t.Errorf("positional args = %v, want [program.json]", fs.Args())
	}
	if !fs.Lookup("trace").Value.(interface{ IsBoolFlag() bool }).IsBoolFlag() {
		t.Error("expected --trace to be registered as a bool flag")
	}
}
