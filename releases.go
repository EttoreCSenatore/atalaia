// Package atalaia holds files that live at the repository root and are embedded
// in the binary. The program itself is in cmd/atalaia and internal/.
package atalaia

import _ "embed"

// ReleasesYAML is releases.yaml, the single source for "What's new", the
// version shown in the app and the GitHub release text. Parse it with
// internal/releases.
//
//go:embed releases.yaml
var ReleasesYAML []byte
