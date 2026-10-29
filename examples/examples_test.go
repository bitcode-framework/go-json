package runtime_test

import (
	"os"
	"path/filepath"
	"testing"

	goio "github.com/bitcode-framework/go-json/io"
	"github.com/bitcode-framework/go-json/lang"
	gojson "github.com/bitcode-framework/go-json/runtime"
	"github.com/bitcode-framework/go-json/stdlib"
)

// examplesDir holds the runnable programs referenced by README.md.
const examplesDir = "../examples"

// newRuntime builds a runtime wired the same way the CLI does, so an example
// that passes here also runs via `go-json run`.
func newRuntime(t *testing.T, opts ...gojson.Option) *gojson.Runtime {
	t.Helper()

	reg := stdlib.DefaultRegistry()
	base := []gojson.Option{
		gojson.WithStdlib(reg.All()),
		gojson.WithStdlibEnv(reg.EnvVars()),
	}
	rt := gojson.NewRuntime(append(base, opts...)...)
	t.Cleanup(func() { _ = rt.Close() })
	return rt
}

// examplePath resolves an example file relative to the package directory,
// regardless of the current working directory (some tests chdir to the
// package root so fs.* paths in the examples resolve the way the README
// documents).
func examplePath(name string) string {
	path := filepath.Join(examplesDir, name)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return filepath.Join("examples", name)
}

// runExample compiles and executes an example file, failing the test on error.
func runExample(t *testing.T, name string, input map[string]any, opts ...gojson.Option) *lang.ExecutionResult {
	t.Helper()

	path := examplePath(name)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("example %s is missing: %v", name, err)
	}

	rt := newRuntime(t, opts...)
	program, err := rt.CompileFile(path)
	if err != nil {
		t.Fatalf("compile %s: %v", name, err)
	}

	result, err := rt.Execute(program, input)
	if err != nil {
		t.Fatalf("execute %s: %v", name, err)
	}
	return result
}

// numericEqual compares two numeric values regardless of their concrete Go
// type. The engine returns ints for integer arithmetic but float64 for
// JSON-sourced numbers, so examples cannot assert an exact Go type.
func numericEqual(got, want any) bool {
	toFloat := func(v any) (float64, bool) {
		switch n := v.(type) {
		case int:
			return float64(n), true
		case int64:
			return float64(n), true
		case float64:
			return n, true
		case float32:
			return float64(n), true
		default:
			return 0, false
		}
	}
	g, okG := toFloat(got)
	w, okW := toFloat(want)
	return okG && okW && g == w
}

// valueMatches reports whether got matches want, comparing numbers
// numerically and everything else by equality.
func valueMatches(got, want any) bool {
	if _, ok := want.(int); ok {
		return numericEqual(got, want)
	}
	if _, ok := want.(float64); ok {
		return numericEqual(got, want)
	}
	return got == want
}

// TestExamples_Run guards every program referenced by README.md.
//
// These are the exact files the documentation tells users to run, so a broken
// example is a broken README. Each case previously failed in a different way:
// the CLI zeroed MaxSteps, and several snippets used response shapes or
// signatures the engine does not accept.
func TestExamples_Run(t *testing.T) {
	tests := []struct {
		name  string
		file  string
		input map[string]any
		opts  []gojson.Option
		want  any
	}{
		{
			name: "hello world",
			file: "01-hello-world.json",
			want: "Hello, World!",
		},
		{
			name:  "input and string building",
			file:  "02-greet-input.json",
			input: map[string]any{"name": "Alice", "age": 30},
			want:  "Hello, Alice! You are 30 years old.",
		},
		{
			name: "functions and recursion",
			file: "03-factorial.json",
			want: 3628800,
		},
		{
			name:  "jsonc comments and trailing commas",
			file:  "04-discount.jsonc",
			input: map[string]any{"price": 100, "tier": "gold"},
			want:  float64(80),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := runExample(t, tc.file, tc.input, tc.opts...)
			if !valueMatches(result.Value, tc.want) {
				t.Errorf("%s: got %v (%T), want %v (%T)",
					tc.file, result.Value, result.Value, tc.want, tc.want)
			}
		})
	}
}

// TestExamples_DiscountTiers checks every branch of the JSONC example.
func TestExamples_DiscountTiers(t *testing.T) {
	cases := map[string]float64{
		"gold":   80,
		"silver": 90,
		"basic":  100,
	}

	for tier, want := range cases {
		t.Run(tier, func(t *testing.T) {
			result := runExample(t, "04-discount.jsonc", map[string]any{
				"price": 100,
				"tier":  tier,
			})
			if !valueMatches(result.Value, want) {
				t.Errorf("tier %q: got %v, want %v", tier, result.Value, want)
			}
		})
	}
}

// TestExamples_IOModule runs the fs example against its bundled data file.
//
// The example reads its data by a path relative to the working directory (that
// is how fs.* resolves paths), so the test runs from the package root — the
// same place the README tells users to invoke the CLI.
func TestExamples_IOModule(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(".."); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	sec := goio.DefaultSecurityConfig()

	result := runExample(t, "06-io-modules.jsonc", nil,
		gojson.WithIOSecurity(sec),
		gojson.WithIO(goio.FS(sec)),
	)

	summary, ok := result.Value.(map[string]any)
	if !ok {
		t.Fatalf("expected a map result, got %T", result.Value)
	}
	if got := summary["orders"]; !numericEqual(got, 3) {
		t.Errorf("orders = %v, want 3", got)
	}
	if got := summary["total"]; !numericEqual(got, 500) {
		t.Errorf("total = %v, want 500", got)
	}
}

// TestExamples_CompileOnly covers examples that cannot execute in unit tests
// because they need external pieces (a live server, script runtimes).
func TestExamples_CompileOnly(t *testing.T) {
	files := []string{
		"05-web-server.json",
		"07-script-plugins/script_pipeline.json",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			path := examplePath(file)
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("example %s is missing: %v", file, err)
			}

			rt := newRuntime(t)
			if _, err := rt.CompileFile(path); err != nil {
				t.Errorf("compile %s: %v", file, err)
			}
		})
	}
}
