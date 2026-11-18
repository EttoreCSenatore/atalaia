package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"version"}, &out, &errOut); code != 0 || !strings.HasPrefix(out.String(), "atalaia ") {
		t.Fatalf("version: code=%d out=%q", code, out.String())
	}
	out.Reset()
	if code := run(nil, &out, &errOut); code != 2 {
		t.Fatalf("no args: code=%d, want 2", code)
	}
	if code := run([]string{"bogus"}, &out, &errOut); code != 2 {
		t.Fatalf("unknown: code=%d, want 2", code)
	}
}
