# go-json

A general-purpose programming language written in JSON/JSONC, embeddable in Go applications.

**Write programs in JSON. Run them anywhere Go runs.**

```json
{
  "name": "hello",
  "go_json": "1",
  "steps": [
    {"let": "name", "expr": "input.name ?? 'World'"},
    {"return": "'Hello, ' + name + '!'"}
  ]
}
```

```bash
$ go-json run hello.json --input '{"name": "Alice"}'
"Hello, Alice!"
```

## Why go-json?

go-json fills a gap no existing tool covers: a **general-purpose JSON programming language** that is embeddable in Go applications.

| Tool | What It Does | What It Can't Do |
|------|-------------|-----------------|
| jsonnet | Data templating | No side effects, no I/O, no control flow |
| jq | JSON transformation | Single-purpose, no general programming |
| AWS Step Functions | Cloud workflows | AWS-locked, no local execution |
| Node-RED | IoT/integration | Node.js-dependent, visual-only |
| CEL / expr-lang | Expression evaluation | Expressions only, no statements or functions |

go-json is a **complete programming language** — variables, functions, loops, structs, imports, error handling, parallel execution, I/O, and a built-in web server — all expressed as valid JSON.

### Use Cases

- **No-code/low-code engine** — Build visual programming UIs where the output is JSON
- **Business rules & automation** — Define rules that non-developers can read and modify
- **Embeddable scripting** — Add user-programmable logic to your Go application
- **Workflow orchestration** — Multi-step processes with branching, loops, and parallel execution
- **API servers** — Declarative HTTP routing with middleware, auth, and templates
- **Code generation** — Write once in JSON, generate Go, JavaScript, or Python

## Installation

### One-line install (no Go required)

Downloads a prebuilt binary for your OS and architecture, verifies its
checksum, and puts it on your PATH.

**Linux / macOS:**

```bash
curl -fsSL https://raw.githubusercontent.com/bitcode-framework/go-json-runtimes/main/install.sh | sh
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/bitcode-framework/go-json-runtimes/main/install.ps1 | iex
```

Git Bash, MSYS2, and Cygwin on Windows also work with the `install.sh` one-liner.

Pin a version or choose the install directory:

```bash
GO_JSON_VERSION=v0.1.4 GO_JSON_BIN_DIR="$HOME/bin" \
  curl -fsSL https://raw.githubusercontent.com/bitcode-framework/go-json-runtimes/main/install.sh | sh
```

```powershell
& ([scriptblock]::Create((irm https://raw.githubusercontent.com/bitcode-framework/go-json-runtimes/main/install.ps1))) -Version v0.1.4 -BinDir C:\tools
```

The installer falls back to `go install` when no prebuilt binary matches your
platform and a Go toolchain is present.

### With Go

```bash
# Full build: includes the script runtimes (goja, quickjs, yaegi, wasm)
go install github.com/bitcode-framework/go-json-runtimes/cmd/go-json@latest

# Lean build: core language only, zero extra dependencies
go install github.com/bitcode-framework/go-json/cmd/go-json-lite@latest
```

Two binaries exist because `go-json-runtimes` depends on `go-json`. Core cannot
import the runtimes without creating a module cycle, and core deliberately keeps
zero dependencies. Both are called `go-json` at run time when installed from the
runtimes module — that one is the recommended default because `script:` imports
work out of the box.

### As a Go library

```bash
go get github.com/bitcode-framework/go-json@latest
```

**Requirements:** Go 1.24+ (only for the Go install path; the one-line
installer needs no toolchain).

## Quick Start

Every snippet below has a runnable counterpart in [`examples/`](examples/) —
clone the repo and run them directly.

### 1. Hello World

Create `hello.json`:

```json
{
  "name": "hello",
  "go_json": "1",
  "steps": [
    {"let": "greeting", "value": "Hello, World!"},
    {"return": "greeting"}
  ]
}
```

```bash
go-json run hello.json
```

Runnable: [`examples/01-hello-world.json`](examples/01-hello-world.json)

### 2. With Input

Create `greet.json`:

```json
{
  "name": "greet",
  "input": {"name": "string", "age": "int"},
  "steps": [
    {"let": "msg", "expr": "'Hello, ' + input.name + '! You are ' + string(input.age) + ' years old.'"},
    {"return": "msg"}
  ]
}
```

```bash
go-json run greet.json --input '{"name": "Alice", "age": 30}'
```

Runnable: [`examples/02-greet-input.json`](examples/02-greet-input.json)
(with input file [`02-greet-input.input.json`](examples/02-greet-input.input.json))

Flags work before or after the program path.

### 3. Functions & Recursion

```json
{
  "name": "math_demo",
  "functions": {
    "factorial": {
      "params": {"n": "int"},
      "returns": "int",
      "steps": [
        {"if": "n <= 1", "then": [{"return": 1}]},
        {"let": "sub", "call": "factorial", "with": {"n": "n - 1"}},
        {"return": "n * sub"}
      ]
    }
  },
  "steps": [
    {"let": "result", "call": "factorial", "with": {"n": "10"}},
    {"return": "result"}
  ]
}
```

Runnable: [`examples/03-factorial.json`](examples/03-factorial.json)

### 4. Web Server

```json
{
  "name": "my_api",
  "server": {"port": 3000},
  "functions": {
    "listUsers": {
      "params": {"request": "map"},
      "steps": [
        {"return": {"status": 200, "body": [
          {"id": 1, "name": "Alice"},
          {"id": 2, "name": "Bob"}
        ]}}
      ]
    }
  },
  "routes": [
    {"method": "GET", "path": "/api/users", "handler": "listUsers"}
  ]
}
```

```bash
go-json serve api.json
# Server running at http://localhost:3000
# Swagger UI at http://localhost:3000/docs
```

Runnable: [`examples/05-web-server.json`](examples/05-web-server.json)

Handlers return the response fields at the **top level** (`status`, `body`),
not nested under `value`.

### 5. I/O Modules

```json
{
  "name": "read_and_parse",
  "go_json": "1",
  "import": {"fs": "io:fs"},
  "steps": [
    {"let": "raw", "call": "fs.read", "args": ["data.json"]},
    {"let": "parsed", "expr": "fromJSON(raw)"},
    {"let": "total", "expr": "reduce(parsed.items, #acc + #.amount, 0)"},
    {"let": "summary", "expr": "{orders: len(parsed.items), total: total}"},
    {"return": "summary"}
  ]
}
```

```bash
go-json run read_and_parse.json --io fs
```

Runnable: [`examples/06-io-modules.jsonc`](examples/06-io-modules.jsonc)
(with data [`06-io-modules.data.json`](examples/06-io-modules.data.json))

Three ways to call any function — choose based on your needs:

```json
{"call": "http.get", "args": ["https://api.com"]}
{"call": "http.get", "with": ["url"]}
{"let": "r", "expr": "http.get('https://api.com')"}
```

`args` passes **literals**; `with` passes **expressions**. `"args": ["data.json"]`
sends that path, while `"with": ["path"]` sends the variable's value. Stdlib
functions such as `fromJSON` are reached through `expr`, not `call`.

### 6. JSONC Support (Comments!)

go-json supports JSONC — JSON with comments and trailing commas:

```jsonc
{
  // Calculate discount based on customer tier
  "name": "discount_calculator",
  "go_json": "1",
  "functions": {
    "calculateDiscount": {
      "params": {
        "price": "float",
        "tier": "string",
      }, // trailing comma OK
      "returns": "float",
      "steps": [
        {"_c": "Gold customers get 20% off"},
        {"if": "tier == 'gold'", "then": [
          {"return": "price * 0.8"}
        ]},
        /* Silver customers get 10% off */
        {"if": "tier == 'silver'", "then": [
          {"return": "price * 0.9"}
        ]},
        {"return": "price"}
      ]
    }
  },
  "steps": [
    {"let": "final_price", "call": "calculateDiscount", "with": {
      "price": "input.price",
      "tier": "input.tier"
    }},
    {"return": "final_price"}
  ]
}
```

Runnable: [`examples/04-discount.jsonc`](examples/04-discount.jsonc)

### 7. Script Plugins (Multi-Language)

go-json can orchestrate scripts from a single JSON program via the `script:`
import prefix. Each script file is loaded by the matching runtime (detected by
extension) and exposed as a callable module.

This requires the [go-json-runtimes](https://github.com/bitcode-framework/go-json-runtimes)
package, which provides the runtime engines. go-json core stays zero-dependency.

```json
{
  "name": "script_pipeline",
  "go_json": "1",
  "import": {
    "ml": "script:./plugins/predict.js",
    "utils": "script:./helpers/format.js"
  },
  "steps": [
    {"let": "prediction", "call": "ml.call", "args": ["predict", [1, 2, 3, 4, 5]]},
    {"let": "formatted", "call": "utils.call", "with": ["'format'", "prediction"]},
    {"let": "summary", "expr": "{prediction: prediction, formatted: formatted}"},
    {"return": "summary"}
  ]
}
```

Each imported script exposes `call(functionName, ...args)` and `exec(code)` to
your JSON program. Runnable end to end with the goja runtime:
[`examples/07-script-plugins/`](examples/07-script-plugins/).

Of the bundled runtimes, only the embedded ones (goja, quickjs, yaegi) and WASM
execute scripts directly; the Python and Node runtimes validate but report
"direct execution not yet implemented".


## Embedding in Go

```go
package main

import (
    "fmt"
    gojson "github.com/bitcode-framework/go-json/runtime"
    "github.com/bitcode-framework/go-json/stdlib"
)

func main() {
    // Create runtime
    reg := stdlib.DefaultRegistry()
    rt := gojson.NewRuntime(
        gojson.WithStdlib(reg.All()),
        gojson.WithStdlibEnv(reg.EnvVars()),
    )
    defer rt.Close()

    // Compile program
    program, err := rt.CompileFile("program.json")
    if err != nil {
        panic(err)
    }

    // Execute with input
    result, err := rt.Execute(program, map[string]any{
        "name": "Alice",
        "age":  30,
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(result.Value)
    fmt.Printf("Completed in %d steps, %v\n", result.Steps, result.Duration)
}
```

Runnable: [`examples/go/embed-hello`](examples/go/embed-hello)

Use `CompileFile` whenever the program has `import` — `Compile` skips import
resolution, so `io:`/`ext:`/`script:` imports compile but fail at run time with
`variable 'fs' not defined`.

### With I/O Modules

```go
sec := goio.DefaultSecurityConfig()

rt := gojson.NewRuntime(
    gojson.WithStdlib(reg.All()),
    gojson.WithStdlibEnv(reg.EnvVars()),
    gojson.WithIOSecurity(sec),
    gojson.WithIO(goio.FS(sec)),
)
```

Module constructors take a `*SecurityConfig` (pass `nil` for the defaults).
Filesystem writes are disabled by default: set `sec.FS.AllowWrite = true` and
scope it with `sec.FS.AllowedPaths`.

Runnable: [`examples/go/embed-io`](examples/go/embed-io)

### With Custom Extensions

```go
rt := gojson.NewRuntime(
    gojson.WithStdlib(reg.All()),
    gojson.WithStdlibEnv(reg.EnvVars()),
    gojson.WithoutIO(),
    gojson.WithExtension("myapp", gojson.Extension{
        Functions: map[string]any{
            // Must be func(...any) (any, error) — the native signature.
            "getUser": func(args ...any) (any, error) {
                return map[string]any{"name": "Alice"}, nil
            },
        },
    }),
)
```

Runnable: [`examples/go/embed-extension`](examples/go/embed-extension)

## Documentation

| Document | Description |
|----------|-------------|
| [Examples](examples/README.md) | Runnable programs for every snippet in this README |
| [Features](features.md) | Complete feature overview with capability matrix |
| **Language** | |
| [Language Reference](docs/language-reference.md) | Step types, syntax, type system, scoping rules |
| [Built-in Functions](docs/built-in-functions.md) | All 110+ functions (expr-lang + stdlib) |
| **I/O & Server** | |
| [I/O Modules](docs/io-modules.md) | HTTP, FS, SQL, Exec, MongoDB, Redis |
| [Web Server](docs/web-server.md) | Routing, middleware, auth, templates, OpenAPI |
| **Integration** | |
| [Embedding Guide](docs/embedding-guide.md) | Go API, extensions, resource limits |
| [CLI Reference](docs/cli-reference.md) | All commands and flags |
| [Code Generation](docs/code-generation.md) | Go/JS/Python codegen, CRUD generator, patterns |

## Architecture

```
                    ┌─────────────────────────────────────────┐
                    │              Your Program               │
                    │           (.json / .jsonc)               │
                    └──────────────┬──────────────────────────┘
                                   │
                    ┌──────────────▼──────────────────────────┐
                    │         JSONC Pre-processor              │
                    │   Strip comments, trailing commas        │
                    └──────────────┬──────────────────────────┘
                                   │
                    ┌──────────────▼──────────────────────────┐
                    │            Parser                        │
                    │      JSON → AST nodes                    │
                    └──────────────┬──────────────────────────┘
                                   │
                    ┌──────────────▼──────────────────────────┐
                    │        Import Resolver                   │
                    │   Resolve ./relative, stdlib:, io:, ext:,│
                    │                    script:                │
                    └──────────────┬──────────────────────────┘
                                   │
                    ┌──────────────▼──────────────────────────┐
                    │           Compiler                       │
                    │  Struct registration, validation,        │
                    │  limit resolution → CompiledProgram      │
                    └──────────────┬──────────────────────────┘
                                   │
                    ┌──────────────▼──────────────────────────┐
                    │         Virtual Machine                  │
                    │  Tree-walk interpreter with debug hooks  │
                    │  Expressions via expr-lang/expr          │
                    └─────────────────────────────────────────┘
```

**Key design decisions:**

- **Compile-once, run-many** — Programs compile to an immutable `CompiledProgram`. Multiple goroutines can execute the same program concurrently.
- **Expression engine abstraction** — The VM never calls expr-lang directly. All expressions go through the `ExprEngine` interface.
- **Safe by default** — Resource limits (step count, call depth, loop iterations, timeout, memory proxies) enforced at every step.
- **JSON-native** — Programs are valid JSON (or JSONC). No custom syntax to learn beyond JSON.

## Related Repositories

| Repository | Description |
|------------|-------------|
| [go-json](https://github.com/bitcode-framework/go-json) | This repo. JSON/JSONC programming language engine in Go |
| [go-json-runtimes](https://github.com/bitcode-framework/go-json-runtimes) | Script runtime engines for go-json (Goja, QuickJS, Yaegi, Python subprocess) |
| [ui-stencil-web-components](https://github.com/bitcode-framework/ui-stencil-web-components) | 119 Stencil Web Components for enterprise applications |
| [ui-tauri](https://github.com/bitcode-framework/ui-tauri) | Tauri 2.0 native shell for desktop and mobile |

## License

MIT
