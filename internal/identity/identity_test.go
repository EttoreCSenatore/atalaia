package identity

import (
	"net"
	"strings"
	"testing"
)

func mac(s string) net.HardwareAddr {
	m, err := net.ParseMAC(s)
	if err != nil {
		panic(err)
	}
	return m
}

func TestEmbeddedRegistry(t *testing.T) {
	db := Default()
	if db.Len() < 30000 {
		t.Fatalf("embedded registry too small: %d", db.Len())
	}
	if v := db.Vendor(mac("b8:27:eb:12:34:56")); !strings.Contains(v, "Raspberry Pi") {
		t.Fatalf("vendor = %q", v)
	}
	if v := db.Vendor(mac("00:00:00:00:00:00")); strings.Contains(v, "\x00") {
		t.Fatalf("unsafe vendor %q", v)
	}
}

func TestVendorMostSpecificWins(t *testing.T) {
	db, err := Load(strings.NewReader("A8BBCC\tBig Corp\nA8BBCCD\tMedium Unit\nA8BBCCDDE\tSmall Unit\n"))
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"a8:bb:cc:dd:e1:00": "Small Unit",
		"a8:bb:cc:d1:00:00": "Medium Unit",
		"a8:bb:cc:11:22:33": "Big Corp",
		"a8:bb:cd:11:22:33": "",
	}
	for m, want := range cases {
		if got := db.Vendor(mac(m)); got != want {
			t.Errorf("Vendor(%s) = %q, want %q", m, got, want)
		}
	}
}

func TestRandomizedMACHasNoVendor(t *testing.T) {
	db, _ := Load(strings.NewReader("BA27EB\tShould Not Match\n"))
	m := mac("ba:27:eb:00:00:01") // locally administered bit set
	if !Randomized(m) {
		t.Fatal("expected randomized")
	}
	if db.Vendor(m) != "" {
		t.Fatal("randomized MAC must not get a vendor")
	}
	if Randomized(mac("b8:27:eb:00:00:01")) {
		t.Fatal("globally unique MAC reported as randomized")
	}
}

func TestSanitizeLabel(t *testing.T) {
	cases := []struct{ in, want string }{
		{"  my\tphone \n 2  ", "my phone 2"},
		{"evil\x1b[31mred\x00", "evil[31mre"},
		{"abc\u202edef", "abcdef"}, // right-to-left override
		{"caf\xc3\xa9\xff", "café"},
		{strings.Repeat("a", 100), strings.Repeat("a", 10)},
	}
	for _, c := range cases {
		if got := SanitizeLabel(c.in, 10); got != c.want {
			t.Errorf("SanitizeLabel(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
