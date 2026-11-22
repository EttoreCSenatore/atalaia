// Package releases reads releases.yaml, the single source for what changed in
// each version: the web app's "What's new" page, the GitHub release text and
// the CI checks that keep versions, tags and CHANGELOG.md in step.
package releases

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.yaml.in/yaml/v3"

	atalaia "github.com/EttoreCSenatore/atalaia"
)

// Notes is one language's text for a release.
type Notes struct {
	Title string   `yaml:"title" json:"title"`
	Items []string `yaml:"items" json:"items"`
}

// Release is one published version, newest first in the file.
type Release struct {
	Version string `yaml:"version" json:"version"`
	Date    string `yaml:"date" json:"date"`
	EN      Notes  `yaml:"en" json:"en"`
	PT      Notes  `yaml:"pt-BR" json:"pt-BR"`
}

type file struct {
	Releases []Release `yaml:"releases"`
}

var semverRE = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-([0-9A-Za-z.-]+))?$`)

// Parse decodes and validates releases.yaml: unknown keys are refused, versions
// are semver (no leading "v"), strictly newest first and unique, dates are
// YYYY-MM-DD and never newer than the release above, and both languages have a
// title and the same number of items.
func Parse(b []byte) ([]Release, error) {
	dec := yaml.NewDecoder(bytes.NewReader(b))
	dec.KnownFields(true)
	var f file
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("releases.yaml: %w", err)
	}
	if len(f.Releases) == 0 {
		return nil, errors.New("releases.yaml: no releases")
	}
	var prevDate time.Time
	for i, r := range f.Releases {
		where := fmt.Sprintf("releases.yaml: release %d (%q)", i+1, r.Version)
		if !semverRE.MatchString(r.Version) {
			return nil, fmt.Errorf("%s: version must be X.Y.Z without a leading v", where)
		}
		d, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			return nil, fmt.Errorf("%s: date must be YYYY-MM-DD", where)
		}
		if i > 0 {
			if Compare(r.Version, f.Releases[i-1].Version) >= 0 {
				return nil, fmt.Errorf("%s: must be older than %s (newest first, no duplicates)", where, f.Releases[i-1].Version)
			}
			if d.After(prevDate) {
				return nil, fmt.Errorf("%s: date is after the newer release's date", where)
			}
		}
		prevDate = d
		for lang, n := range map[string]Notes{"en": r.EN, "pt-BR": r.PT} {
			if strings.TrimSpace(n.Title) == "" || len(n.Items) == 0 {
				return nil, fmt.Errorf("%s: %s needs a title and at least one item", where, lang)
			}
			for _, it := range n.Items {
				if strings.TrimSpace(it) == "" {
					return nil, fmt.Errorf("%s: %s has an empty item", where, lang)
				}
			}
		}
		if len(r.EN.Items) != len(r.PT.Items) {
			return nil, fmt.Errorf("%s: en has %d items and pt-BR has %d", where, len(r.EN.Items), len(r.PT.Items))
		}
	}
	return f.Releases, nil
}

var load = sync.OnceValues(func() ([]Release, error) { return Parse(atalaia.ReleasesYAML) })

// All returns the embedded releases, newest first. The file is checked in CI
// and in tests, so an error here means a broken build.
func All() ([]Release, error) { return load() }

// Find returns the release for v, with or without a leading "v".
func Find(rs []Release, v string) (Release, bool) {
	v = strings.TrimPrefix(v, "v")
	for _, r := range rs {
		if r.Version == v {
			return r, true
		}
	}
	return Release{}, false
}

// Compare orders two semver versions (no leading "v"): -1, 0 or 1. A
// pre-release sorts before its release; pre-release tags compare as text.
func Compare(a, b string) int {
	ma, mb := semverRE.FindStringSubmatch(a), semverRE.FindStringSubmatch(b)
	if ma == nil || mb == nil {
		return strings.Compare(a, b)
	}
	for i := 1; i <= 3; i++ {
		x, _ := strconv.Atoi(ma[i])
		y, _ := strconv.Atoi(mb[i])
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}
	switch {
	case ma[4] == mb[4]:
		return 0
	case ma[4] == "":
		return 1
	case mb[4] == "":
		return -1
	}
	return strings.Compare(ma[4], mb[4])
}

// Markdown is the GitHub release text for r: English first, then Portuguese.
func Markdown(r Release) string {
	var b strings.Builder
	section := func(n Notes) {
		fmt.Fprintf(&b, "### %s\n\n", n.Title)
		for _, it := range n.Items {
			fmt.Fprintf(&b, "- %s\n", it)
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "**English**\n\n")
	section(r.EN)
	fmt.Fprintf(&b, "**Português**\n\n")
	section(r.PT)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

var changelogRE = regexp.MustCompile(`(?m)^## v(\S+) — (\d{4}-\d{2}-\d{2})\s*$`)

// CheckChangelog makes sure every release has a "## vX.Y.Z — date" heading in
// CHANGELOG.md with the same date, and every such heading has a release.
func CheckChangelog(rs []Release, changelog []byte) error {
	heads := map[string]string{}
	for _, m := range changelogRE.FindAllSubmatch(changelog, -1) {
		heads[string(m[1])] = string(m[2])
	}
	var errs []error
	for _, r := range rs {
		d, ok := heads[r.Version]
		switch {
		case !ok:
			errs = append(errs, fmt.Errorf("CHANGELOG.md has no \"## v%s — %s\" heading", r.Version, r.Date))
		case d != r.Date:
			errs = append(errs, fmt.Errorf("v%s: CHANGELOG.md says %s, releases.yaml says %s", r.Version, d, r.Date))
		}
		delete(heads, r.Version)
	}
	for v := range heads {
		errs = append(errs, fmt.Errorf("CHANGELOG.md has v%s, which is missing from releases.yaml", v))
	}
	return errors.Join(errs...)
}

// CheckTag makes sure a release tag (vX.Y.Z) is the newest entry, so the
// published notes and the binary's version always match the file.
func CheckTag(rs []Release, tag string) error {
	if !strings.HasPrefix(tag, "v") {
		return fmt.Errorf("tag %q must start with v", tag)
	}
	if len(rs) == 0 || "v"+rs[0].Version != tag {
		newest := ""
		if len(rs) > 0 {
			newest = "v" + rs[0].Version
		}
		return fmt.Errorf("tag %s is not the newest release in releases.yaml (%s): add it at the top first", tag, newest)
	}
	return nil
}
