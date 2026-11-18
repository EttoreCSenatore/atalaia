// Package identity tells what a device is from what it announces on the
// network: vendor from the MAC address (IEEE OUI registry) and whether the MAC
// is a randomized private address.
//
// Everything a device announces (names, banners) is untrusted input, so labels
// are sanitized before they are stored or printed.
package identity

import (
	"bufio"
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/hex"
	"io"
	"net"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

//go:embed oui.tsv.gz
var ouiGz []byte

// DB maps MAC prefixes (6, 7 or 9 uppercase hex digits) to vendor names.
type DB struct {
	prefixes map[string]string
}

var (
	defaultOnce sync.Once
	defaultDB   *DB
)

// Default returns the embedded IEEE registry (loaded on first use).
func Default() *DB {
	defaultOnce.Do(func() {
		zr, err := gzip.NewReader(bytes.NewReader(ouiGz))
		if err == nil {
			defaultDB, err = Load(zr)
		}
		if err != nil {
			defaultDB = &DB{prefixes: map[string]string{}}
		}
	})
	return defaultDB
}

// Load reads `PREFIX<TAB>VENDOR` lines.
func Load(r io.Reader) (*DB, error) {
	db := &DB{prefixes: make(map[string]string, 60000)}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		prefix, name, ok := strings.Cut(sc.Text(), "\t")
		if ok && prefix != "" && name != "" {
			db.prefixes[prefix] = name
		}
	}
	return db, sc.Err()
}

// Len returns the number of prefixes.
func (d *DB) Len() int { return len(d.prefixes) }

// Randomized reports whether the MAC is locally administered (the 0x02 bit of
// the first octet). Phones and laptops use such addresses on purpose for
// privacy, so they have no vendor and change between networks.
func Randomized(mac net.HardwareAddr) bool {
	return len(mac) > 0 && mac[0]&0x02 != 0
}

// Vendor returns the registered vendor of the MAC, or "" when unknown or when
// the MAC is randomized. The most specific registration (36, 28, then 24 bits)
// wins.
func (d *DB) Vendor(mac net.HardwareAddr) string {
	if len(mac) < 5 || Randomized(mac) {
		return ""
	}
	h := strings.ToUpper(hex.EncodeToString(mac[:5])) // 10 digits; the longest key is 9
	for _, n := range []int{9, 7, 6} {
		if v, ok := d.prefixes[h[:n]]; ok {
			return v
		}
	}
	return ""
}

// SanitizeLabel makes text announced by a device safe to store and print:
// control and bidirectional-override characters are dropped, whitespace is
// collapsed and the result is cut to max runes.
func SanitizeLabel(s string, max int) string {
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "")
	}
	var b strings.Builder
	space := false
	for _, r := range s {
		switch {
		case unicode.IsSpace(r):
			space = b.Len() > 0
			continue
		case unicode.IsControl(r), unicode.Is(unicode.Cf, r), r == ' ', r == ' ':
			continue
		}
		if space {
			b.WriteByte(' ')
			space = false
		}
		b.WriteRune(r)
	}
	out := b.String()
	if r := []rune(out); len(r) > max {
		out = string(r[:max])
	}
	return out
}
