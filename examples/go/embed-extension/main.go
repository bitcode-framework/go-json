// Command embed-extension shows how to expose your own Go functions to a
// go-json program through the extension API.
//
// It mirrors the "With Custom Extensions" section of the README. As with I/O
// imports, "ext:" imports are resolved by CompileFile, so the program is
// written to a temporary file first.
//
//	go run ./examples/go/embed-extension
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	gojson "github.com/bitcode-framework/go-json/runtime"
	"github.com/bitcode-framework/go-json/stdlib"
)

const program = `{
  "name": "use_extension",
  "import": {"myapp": "ext:myapp"},
  "steps": [
    {"let": "user", "call": "myapp.getUser", "args": ["u-1"]},
    {"let": "greeting", "expr": "'Hello, ' + user.name + '!'"},
    {"return": "greeting"}
  ]
}`

// getUser is the host function exposed to go-json.
//
// Extension functions must match the native signature func(...any) (any, error)
// — the same shape I/O modules use. A plain Go signature such as
// func(id string) (map[string]any, error) compiles but fails at run time with
// "'myapp.getUser' is not a function".
func getUser(args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("getUser: id is required")
	}
	id, _ := args[0].(string)
	return map[string]any{"id": id, "name": "Alice"}, nil
}

func main() {
	dir, err := os.MkdirTemp(".", "embed-ext-out-")
	if err != nil {
		log.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	programPath := filepath.Join(dir, "program.json")
	if err := os.WriteFile(programPath, []byte(program), 0o600); err != nil {
		log.Fatalf("write program: %v", err)
	}

	reg := stdlib.DefaultRegistry()
	rt := gojson.NewRuntime(
		gojson.WithStdlib(reg.All()),
		gojson.WithStdlibEnv(reg.EnvVars()),
		gojson.WithoutIO(),
		gojson.WithExtension("myapp", gojson.Extension{
			Functions: map[string]any{
				"getUser": getUser,
			},
		}),
	)
	defer rt.Close()

	compiled, err := rt.CompileFile(programPath)
	if err != nil {
		log.Fatalf("compile: %v", err)
	}

	result, err := rt.Execute(compiled, nil)
	if err != nil {
		log.Fatalf("execute: %v", err)
	}

	out, err := json.MarshalIndent(result.Value, "", "  ")
	if err != nil {
		log.Fatalf("marshal: %v", err)
	}
	fmt.Println(string(out))
}
