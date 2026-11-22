package releases

import (
	"os"
	"strings"
	"testing"
)

func TestEmbeddedFileIsValidAndMatchesChangelog(t *testing.T) {
	rs, err := All()
	if err != nil {
		t.Fatal(err)
	}
	cl, err := os.ReadFile("../../CHANGELOG.md")
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckChangelog(rs, cl); err != nil {
		t.Fatal(err)
	}
}

const good = `releases:
  - version: 1.1.0
    date: 2026-02-01
    en: {title: Two, items: [b]}
    pt-BR: {title: Dois, items: [b]}
  - version: 1.0.0
    date: 2026-01-01
    en: {title: One, items: [a, c]}
    pt-BR: {title: Um, items: [a, c]}
`

func TestParseAndFind(t *testing.T) {
	rs, err := Parse([]byte(good))
	if err != nil {
		t.Fatal(err)
	}
	if r, ok := Find(rs, "v1.0.0"); !ok || r.PT.Title != "Um" {
		t.Fatalf("Find v1.0.0 = %+v %v", r, ok)
	}
	if _, ok := Find(rs, "2.0.0"); ok {
		t.Fatal("found a release that does not exist")
	}
	md := Markdown(rs[0])
	for _, want := range []string{"### Two", "- b", "**Português**", "### Dois"} {
		if !strings.Contains(md, want) {
			t.Fatalf("markdown lacks %q:\n%s", want, md)
		}
	}
}

func TestParseRejects(t *testing.T) {
	cases := map[string]string{
		"leading v":     strings.Replace(good, "version: 1.1.0", "version: v1.1.0", 1),
		"wrong order":   strings.Replace(good, "version: 1.1.0", "version: 0.9.0", 1),
		"duplicate":     strings.Replace(good, "version: 1.1.0", "version: 1.0.0", 1),
		"bad date":      strings.Replace(good, "2026-02-01", "01/02/2026", 1),
		"date order":    strings.Replace(good, "2026-02-01", "2025-12-01", 1),
		"no pt title":   strings.Replace(good, "title: Dois", "title: ''", 1),
		"item mismatch": strings.Replace(good, "items: [a, c]}\n", "items: [a]}\n", 1),
		"unknown key":   strings.Replace(good, "date: 2026-01-01", "date: 2026-01-01\n    datee: x", 1),
		"empty":         "releases: []\n",
	}
	for name, in := range cases {
		if _, err := Parse([]byte(in)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestCompare(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0}, {"1.2.0", "1.10.0", -1}, {"2.0.0", "1.9.9", 1},
		{"1.0.0-rc.1", "1.0.0", -1}, {"1.0.0", "1.0.0-rc.1", 1}, {"1.0.0-rc.1", "1.0.0-rc.2", -1},
	} {
		if got := Compare(c.a, c.b); got != c.want {
			t.Errorf("Compare(%s, %s) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestCheckChangelogAndTag(t *testing.T) {
	rs, _ := Parse([]byte(good))
	ok := "## Unreleased\n\n## v1.1.0 — 2026-02-01\n\n## v1.0.0 — 2026-01-01\n"
	if err := CheckChangelog(rs, []byte(ok)); err != nil {
		t.Fatal(err)
	}
	for name, cl := range map[string]string{
		"missing":    "## v1.1.0 — 2026-02-01\n",
		"wrong date": "## v1.1.0 — 2026-02-02\n\n## v1.0.0 — 2026-01-01\n",
		"extra":      ok + "\n## v0.9.0 — 2025-12-01\n",
	} {
		if CheckChangelog(rs, []byte(cl)) == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if err := CheckTag(rs, "v1.1.0"); err != nil {
		t.Fatal(err)
	}
	for _, tag := range []string{"v1.0.0", "1.1.0", "v1.2.0"} {
		if CheckTag(rs, tag) == nil {
			t.Errorf("tag %s accepted", tag)
		}
	}
}
