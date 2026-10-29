package cli

import (
	"flag"
	"strings"
)

// boolFlag is implemented by flag.Value types that take no argument.
type boolFlag interface {
	IsBoolFlag() bool
}

// isBoolFlag reports whether f takes no value.
func isBoolFlag(f *flag.Flag) bool {
	if f == nil {
		return false
	}
	if bf, ok := f.Value.(boolFlag); ok {
		return bf.IsBoolFlag()
	}
	return false
}

// reorderArgs moves positional arguments after all flags.
//
// Go's flag package stops parsing at the first non-flag argument, so the
// documented invocation form
//
//	go-json run program.json --input '{"name":"Alice"}'
//
// silently dropped every flag written after the program path: --input was
// never applied and the program ran with empty input. Reordering first makes
// flags work in either position.
//
// A flag's value is consumed together with the flag unless the flag is boolean
// or the value is attached with '='. Unknown flags are left for flag.Parse to
// report. "--" terminates flag parsing, as usual.
func reorderArgs(fs *flag.FlagSet, args []string) []string {
	var flags, positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}

		// A bare "-" is a positional argument (commonly stdin), not a flag.
		if len(arg) < 2 || arg[0] != '-' {
			positional = append(positional, arg)
			continue
		}

		flags = append(flags, arg)

		name := strings.TrimLeft(arg, "-")
		if strings.ContainsRune(name, '=') {
			continue // value attached: --input={...}
		}
		if isBoolFlag(fs.Lookup(name)) {
			continue // boolean flags take no value
		}
		if i+1 < len(args) {
			i++
			flags = append(flags, args[i])
		}
	}

	return append(flags, positional...)
}

// parseFlags parses args after reordering them so flags may appear before or
// after positional arguments.
func parseFlags(fs *flag.FlagSet, args []string) {
	fs.Parse(reorderArgs(fs, args))
}
