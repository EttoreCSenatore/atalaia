// Command releasenotes checks releases.yaml and prints a release's text. It is
// a maintainer tool used by CI and the release workflow; it is not shipped.
//
//	go run ./cmd/releasenotes check [-changelog CHANGELOG.md] [-tag vX.Y.Z]
//	go run ./cmd/releasenotes notes vX.Y.Z
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/EttoreCSenatore/atalaia/internal/releases"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "releasenotes:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: releasenotes check [-changelog FILE] [-tag vX.Y.Z] | notes vX.Y.Z")
	}
	rs, err := releases.All()
	if err != nil {
		return err
	}
	switch args[0] {
	case "check":
		fs := flag.NewFlagSet("check", flag.ContinueOnError)
		changelog := fs.String("changelog", "CHANGELOG.md", "CHANGELOG to compare against (empty to skip)")
		tag := fs.String("tag", "", "release tag that must be the newest entry")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *changelog != "" {
			b, err := os.ReadFile(*changelog)
			if err != nil {
				return err
			}
			if err := releases.CheckChangelog(rs, b); err != nil {
				return err
			}
		}
		if *tag != "" {
			if err := releases.CheckTag(rs, *tag); err != nil {
				return err
			}
		}
		fmt.Fprintf(out, "releases.yaml ok: %d releases, newest v%s\n", len(rs), rs[0].Version)
		return nil
	case "notes":
		if len(args) != 2 {
			return fmt.Errorf("usage: releasenotes notes vX.Y.Z")
		}
		r, ok := releases.Find(rs, args[1])
		if !ok {
			return fmt.Errorf("%s is not in releases.yaml", args[1])
		}
		_, err := io.WriteString(out, releases.Markdown(r))
		return err
	}
	return fmt.Errorf("unknown command %q", args[0])
}
