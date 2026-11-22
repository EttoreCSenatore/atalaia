package main

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// The configuration reference (pt-BR and English) names every command, every
// option of every command and every ATALAIA_ environment variable in the code.
func TestConfigReferenceIsComplete(t *testing.T) {
	docs := map[string]string{}
	for _, p := range []string{
		"../../website/docs/operacao/configuracao.md",
		"../../website/i18n/en/docusaurus-plugin-content-docs/current/operacao/configuracao.md",
	} {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		docs[p] = string(b)
	}
	commands := []string{"serve", "scan", "devices", "history", "export", "backup", "verify-backup", "reset-password", "health", "scope", "version"}
	var usage syncBuf
	(&app{out: &usage, err: &usage, now: time.Now}).run(context.Background(), []string{"help"})
	for _, m := range regexp.MustCompile(`(?m)^  atalaia ([a-z-]+)`).FindAllStringSubmatch(usage.String(), -1) {
		if !contains(commands, m[1]) {
			t.Errorf("command %q is in the usage text but not in this test's list", m[1])
		}
	}
	flagRE := regexp.MustCompile(`(?m)^  -([a-z-]+)`)
	for _, c := range commands {
		if c == "scope" || c == "version" {
			continue
		}
		var out syncBuf
		(&app{out: &out, err: &out, now: time.Now}).run(context.Background(), []string{c, "-h"})
		for _, m := range flagRE.FindAllStringSubmatch(out.String(), -1) {
			for p, d := range docs {
				if !strings.Contains(d, "`--"+m[1]+"`") && !strings.Contains(d, "`--"+m[1]+",") {
					t.Errorf("%s: option --%s of %s is missing", p, m[1], c)
				}
			}
		}
	}
	for _, c := range commands {
		for p, d := range docs {
			if !strings.Contains(d, "`atalaia "+c+"`") {
				t.Errorf("%s: command %s is missing", p, c)
			}
		}
	}
	envRE := regexp.MustCompile(`"(ATALAIA_[A-Z_]+)"`)
	vars := map[string]bool{}
	for _, dir := range []string{".", "../../internal"} {
		_ = filepath.WalkDir(dir, func(path string, e os.DirEntry, err error) error {
			if err != nil || e.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			b, _ := os.ReadFile(path)
			for _, m := range envRE.FindAllStringSubmatch(string(b), -1) {
				vars[m[1]] = true
			}
			return nil
		})
	}
	if len(vars) < 10 {
		t.Fatalf("found only %d variables; the scan is broken", len(vars))
	}
	for v := range vars {
		for p, d := range docs {
			if !strings.Contains(d, "`"+v) {
				t.Errorf("%s: variable %s is missing", p, v)
			}
		}
	}
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
