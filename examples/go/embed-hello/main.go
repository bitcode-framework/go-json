// Command embed-hello embeds go-json in a Go application.
//
// It mirrors the "Embedding in Go" section of the README and runs the
// hello.json program shipped next to it.
//
//	go run ./examples/go/embed-hello
package main

import (
	"fmt"
	"log"
	"path/filepath"

	gojson "github.com/bitcode-framework/go-json/runtime"
	"github.com/bitcode-framework/go-json/stdlib"
)

func main() {
	reg := stdlib.DefaultRegistry()
	rt := gojson.NewRuntime(
		gojson.WithStdlib(reg.All()),
		gojson.WithStdlibEnv(reg.EnvVars()),
	)
	defer rt.Close()

	path := filepath.Join("examples", "01-hello-world.json")
	program, err := rt.CompileFile(path)
	if err != nil {
		log.Fatalf("compile %s: %v", path, err)
	}

	result, err := rt.Execute(program, map[string]any{})
	if err != nil {
		log.Fatalf("execute: %v", err)
	}

	fmt.Println(result.Value)
	fmt.Printf("completed in %d steps in %v\n", result.Steps, result.Duration)
}
