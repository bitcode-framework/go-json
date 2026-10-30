//go:build js && wasm

// Command wasm-playground exposes go-json to the browser for the live demo at
// https://bitcode-framework.github.io/go-json/
//
// The JS side calls goJsonRun(sourceJSON) and receives a JSON string:
//
//	{"ok":true,"value":...,"steps":N,"duration":...}
//	{"ok":false,"error":"..."}
package main

import (
	"encoding/json"
	"syscall/js"
	"time"

	gojson "github.com/bitcode-framework/go-json/runtime"
	"github.com/bitcode-framework/go-json/stdlib"
)

type result struct {
	OK       bool   `json:"ok"`
	Value    any    `json:"value,omitempty"`
	Steps    int    `json:"steps,omitempty"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

func respond(r result) string {
	b, err := json.Marshal(r)
	if err != nil {
		return `{"ok":false,"error":"failed to encode result"}`
	}
	return string(b)
}

func main() {
	reg := stdlib.DefaultRegistry()
	rt := gojson.NewRuntime(
		gojson.WithStdlib(reg.All()),
		gojson.WithStdlibEnv(reg.EnvVars()),
		gojson.WithoutIO(),
	)
	defer rt.Close()

	run := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) < 1 {
			return respond(result{Error: "missing program source"})
		}

		source := args[0].String()
		start := time.Now()

		program, err := rt.Compile([]byte(source))
		if err != nil {
			return respond(result{Error: err.Error()})
		}

		exec, err := rt.Execute(program, nil)
		if err != nil {
			return respond(result{Error: err.Error()})
		}

		return respond(result{
			OK:       true,
			Value:    exec.Value,
			Steps:    exec.Steps,
			Duration: time.Since(start).String(),
		})
	})
	defer run.Release()

	// check validates a program without executing it.
	check := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) < 1 {
			return respond(result{Error: "missing program source"})
		}
		if _, err := rt.Compile([]byte(args[0].String())); err != nil {
			return respond(result{Error: err.Error()})
		}
		return respond(result{OK: true, Value: "valid"})
	})
	defer check.Release()

	js.Global().Set("goJsonRun", run)
	js.Global().Set("goJsonCheck", check)
	js.Global().Set("goJsonReady", js.ValueOf(true))

	select {}
}
