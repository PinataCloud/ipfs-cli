package main

import (
	"strings"

	"github.com/urfave/cli/v2"
)

// reorderArgs rewrites the argument list so flags may appear after positional
// arguments. urfave/cli v2 parses flags with the stdlib flag package, which
// stops at the first positional token — so a flag typed after a positional
// (e.g. `templates update <id> --ref x`) is silently dropped. We resolve the
// target command from the app definition, then move that command's flags ahead
// of its positionals so either ordering works.
func reorderArgs(app *cli.App, args []string) []string {
	if len(args) <= 1 {
		return args
	}

	globalFlags := map[string]bool{}
	collectFlagValueInfo(app.Flags, globalFlags)

	// Navigate the command tree to find where the resolved command's argument
	// region begins and which flags belong to it.
	cmds := app.Commands
	var resolvedFlags []cli.Flag
	i := 1
	for i < len(args) {
		tok := args[i]
		if tok == "--" {
			break
		}
		if isFlag(tok) {
			i++
			if !strings.Contains(tok, "=") {
				name := strings.TrimLeft(tok, "-")
				if takesValue, ok := globalFlags[name]; ok && takesValue && i < len(args) {
					i++ // skip this global flag's value
				}
			}
			continue
		}
		next := findCommand(cmds, tok)
		if next == nil {
			break
		}
		resolvedFlags = next.Flags
		cmds = next.Subcommands
		i++
	}

	valueFlags := map[string]bool{}
	collectFlagValueInfo(app.Flags, valueFlags)
	collectFlagValueInfo(resolvedFlags, valueFlags)

	// Split the tail into flags (with their values) and positionals, then
	// re-emit flags first. A "--" terminator and everything after it is left
	// untouched at the very end.
	var flagsOut, posOut, rest []string
	tail := args[i:]
	for j := 0; j < len(tail); j++ {
		tok := tail[j]
		if tok == "--" {
			rest = append(rest, tail[j:]...)
			break
		}
		if isFlag(tok) {
			flagsOut = append(flagsOut, tok)
			if strings.Contains(tok, "=") {
				continue
			}
			name := strings.TrimLeft(tok, "-")
			if takesValue, ok := valueFlags[name]; ok && takesValue && j+1 < len(tail) {
				flagsOut = append(flagsOut, tail[j+1])
				j++
			}
			continue
		}
		posOut = append(posOut, tok)
	}

	out := make([]string, 0, len(args))
	out = append(out, args[:i]...)
	out = append(out, flagsOut...)
	out = append(out, posOut...)
	out = append(out, rest...)
	return out
}

// isFlag reports whether a token is a flag (starts with "-" but is not a lone
// dash, which conventionally means stdin/positional).
func isFlag(tok string) bool {
	return len(tok) > 1 && strings.HasPrefix(tok, "-")
}

// collectFlagValueInfo records, for every flag name and alias, whether it
// consumes a following value. Only bool flags do not.
func collectFlagValueInfo(flags []cli.Flag, m map[string]bool) {
	for _, f := range flags {
		_, isBool := f.(*cli.BoolFlag)
		takesValue := !isBool
		for _, name := range f.Names() {
			m[name] = takesValue
		}
	}
}

func findCommand(cmds []*cli.Command, name string) *cli.Command {
	for _, c := range cmds {
		if c.Name == name {
			return c
		}
		for _, alias := range c.Aliases {
			if alias == name {
				return c
			}
		}
	}
	return nil
}
