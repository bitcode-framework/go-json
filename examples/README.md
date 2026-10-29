# Examples

Every program here is referenced from the top-level [README](../README.md) and
is run by [`examples_test.go`](examples_test.go), so the documentation cannot
drift away from working code.

## CLI examples

Run from `packages/go-json` with the `go-json` binary.

| File | Demonstrates | Command |
|------|--------------|---------|
| `01-hello-world.json` | `let` + `return` | `go-json run examples/01-hello-world.json` |
| `02-greet-input.json` | `input` block, string building | `go-json run examples/02-greet-input.json --input '{"name":"Alice","age":30}'` |
| `03-factorial.json` | functions, recursion | `go-json run examples/03-factorial.json` |
| `04-discount.jsonc` | JSONC comments, trailing commas | `go-json run examples/04-discount.jsonc --input '{"price":100,"tier":"gold"}'` |
| `05-web-server.json` | HTTP server + routes | `go-json serve examples/05-web-server.json` |
| `06-io-modules.jsonc` | `io:fs` import, read + aggregate | `go-json run examples/06-io-modules.jsonc --io fs` |
| `07-script-plugins/` | `script:` imports (needs runtimes) | see below |

Flags may come before or after the program path:

```bash
go-json run examples/02-greet-input.json --input '{"name":"Alice","age":30}'
go-json run --input '{"name":"Alice","age":30}' examples/02-greet-input.json
```

`02` and `04` also ship `.input.json` files, which is more convenient on shells
that mangle quotes:

```bash
go-json run examples/02-greet-input.json --input-file examples/02-greet-input.input.json
```

### Script plugins

`script:` imports need a script runtime, which lives in the separate
[`go-json-runtimes`](../../go-json-runtimes) module. The `go-json` binary built
from that module has every runtime wired in:

```bash
# one-line install (Linux / macOS / Git Bash)
curl -fsSL https://raw.githubusercontent.com/bitcode-framework/go-json-runtimes/main/install.sh | sh

# or, with a Go toolchain
go install github.com/bitcode-framework/go-json-runtimes/cmd/go-json@latest
```

Then:

```bash
cd packages/go-json/examples/07-script-plugins
go-json run script_pipeline.json
```

The pipeline calls `predict` in `plugins/predict.js`, then feeds that result to
`format` in `helpers/format.js`.

Two things to know when writing scripts:

- **`args` passes literals, `with` passes expressions.** `"args": ["predict", [1,2,3]]`
  sends the list itself; `"args": ["format", "prediction"]` would send the
  *string* `"prediction"`. Use `"with": ["'format'", "prediction"]` to pass the
  variable's value.
- **Go values reach the script as `params`.** The runtime injects `params.args`
  (the arguments) and `params.__function` (the function name). Return the result
  from the script body; the runtime picks it up as the return value.

Python (`.py`) and Node (`.ts`, `.mjs`) runtimes are registered but their direct
`Execute` is not implemented yet, so they return an explicit error.

## Go embedding examples

| Directory | Demonstrates |
|-----------|--------------|
| `go/embed-hello` | `NewRuntime`, `CompileFile`, `Execute` |
| `go/embed-io` | exposing I/O modules, scoping filesystem writes |
| `go/embed-extension` | exposing your own Go functions via `ext:` |

```bash
cd packages/go-json
go run ./examples/go/embed-hello
go run ./examples/go/embed-io
go run ./examples/go/embed-extension
```

Notes that the snippets in the README do not make obvious:

- **Use `CompileFile` for anything with `import`.** `Compile` skips import
  resolution, so an `io:`/`ext:`/`script:` import compiles but fails at run time
  with `variable 'fs' not defined`.
- **Filesystem writes are off by default.** `DefaultSecurityConfig` sets
  `FS.AllowWrite: false`; enable it and scope it with `FS.AllowedPaths`.
- **Extension functions use the native signature.** They must be
  `func(...any) (any, error)`, not a plain Go signature.

## Keeping examples honest

```bash
cd packages/go-json
go test ./examples/ -v
```

The suite executes `01`–`04` and `06`, checks the expected values, and
compile-checks `05` and the script pipeline (which need a server or a runtime).
