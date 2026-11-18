// Command atalaia is the Atalaia network watchtower.
//
// M0 only ships the skeleton: `version` and `scope`. Discovery arrives in M1.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/EttoreCSenatore/atalaia/internal/scope"
	"github.com/EttoreCSenatore/atalaia/internal/version"
)

const usage = `Atalaia: open-source network watchtower.

Usage:
  atalaia version   print the version
  atalaia scope     print the networks Atalaia is allowed to probe
`

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, out, errOut io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(errOut, usage)
		return 2
	}
	switch args[0] {
	case "version", "--version":
		fmt.Fprintln(out, "atalaia", version.Version)
		return 0
	case "scope":
		s, err := scope.Local()
		if err != nil {
			fmt.Fprintln(errOut, "error:", err)
			return 1
		}
		ps := s.Prefixes()
		if len(ps) == 0 {
			fmt.Fprintln(errOut, "no private IPv4 network found; nothing in scope")
			return 1
		}
		for _, p := range ps {
			fmt.Fprintln(out, p)
		}
		return 0
	case "help", "-h", "--help":
		fmt.Fprint(out, usage)
		return 0
	default:
		fmt.Fprintf(errOut, "unknown command %q\n\n%s", args[0], usage)
		return 2
	}
}
