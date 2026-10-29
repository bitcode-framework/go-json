// Command embed-io shows how to expose I/O modules to embedded programs.
//
// It mirrors the "With I/O Modules" section of the README. Two things differ
// from the naive snippet, both required for the program to actually run:
//
//   - I/O imports are resolved by CompileFile, not Compile. Compile skips
//     import resolution, so a program with "import": {"fs": "io:fs"} compiles
//     but fails at run time with "variable 'fs' not defined".
//   - Filesystem writes are disabled by DefaultSecurityConfig. Enable them and
//     scope them to a directory with FS.AllowedPaths.
//
//	go run ./examples/go/embed-io
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	goio "github.com/bitcode-framework/go-json/io"
	gojson "github.com/bitcode-framework/go-json/runtime"
	"github.com/bitcode-framework/go-json/stdlib"
)

const program = `{
  "name": "write_and_read",
  "import": {"fs": "io:fs"},
  "steps": [
    {"let": "payload", "expr": "{message: 'written by go-json', items: [1, 2, 3]}"},
    {"call": "fs.write", "with": ["input.target", "toJSON(payload)"]},
    {"let": "raw", "call": "fs.read", "with": ["input.target"]},
    {"let": "parsed", "expr": "fromJSON(raw)"},
    {"return": "parsed"}
  ]
}`

func main() {
	dir, err := os.MkdirTemp(".", "embed-io-out-")
	if err != nil {
		log.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	absDir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("abs dir: %v", err)
	}

	programPath := filepath.Join(absDir, "program.json")
	if err := os.WriteFile(programPath, []byte(program), 0o600); err != nil {
		log.Fatalf("write program: %v", err)
	}

	// Writes are disabled by default; allow them for this directory only.
	sec := goio.DefaultSecurityConfig()
	sec.FS.AllowWrite = true
	sec.FS.AllowedPaths = []string{absDir}

	reg := stdlib.DefaultRegistry()
	rt := gojson.NewRuntime(
		gojson.WithStdlib(reg.All()),
		gojson.WithStdlibEnv(reg.EnvVars()),
		gojson.WithIOSecurity(sec),
		gojson.WithIO(goio.FS(sec)),
	)
	defer rt.Close()

	compiled, err := rt.CompileFile(programPath)
	if err != nil {
		log.Fatalf("compile: %v", err)
	}

	target := filepath.Join(absDir, "result.json")
	result, err := rt.Execute(compiled, map[string]any{"target": target})
	if err != nil {
		log.Fatalf("execute: %v", err)
	}

	out, err := json.MarshalIndent(result.Value, "", "  ")
	if err != nil {
		log.Fatalf("marshal: %v", err)
	}
	fmt.Println(string(out))
}
