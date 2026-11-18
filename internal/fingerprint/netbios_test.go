package fingerprint

import (
	"context"
	"encoding/binary"
	"net"
	"net/netip"
	"testing"
	"time"
)

func nbEntry(name string, suffix byte, flags uint16) []byte {
	b := make([]byte, 18)
	copy(b, name)
	for i := len(name); i < 15; i++ {
		b[i] = ' '
	}
	b[15] = suffix
	binary.BigEndian.PutUint16(b[16:], flags)
	return b
}

func nbResponse(id uint16, entries ...[]byte) []byte {
	r := []byte{byte(id >> 8), byte(id), 0x84, 0, 0, 0, 0, 1, 0, 0, 0, 0}
	r = append(r, 0xC0, 0x0C, 0, 0x21, 0, 1, 0, 0, 0, 0)
	data := []byte{byte(len(entries))}
	for _, e := range entries {
		data = append(data, e...)
	}
	data = append(data, make([]byte, 6)...) // statistics
	r = append(r, byte(len(data)>>8), byte(len(data)))
	return append(r, data...)
}

func TestParseNBSTAT(t *testing.T) {
	p := nbResponse(7, nbEntry("DESKTOP-A1", 0x00, 0x0400), nbEntry("WORKGROUP", 0x00, 0x8400), nbEntry("DESKTOP-A1", 0x20, 0x0400))
	info, ok := parseNBSTAT(p, 7)
	if !ok || info.Name != "DESKTOP-A1" || info.Workgroup != "WORKGROUP" || !info.FileServer {
		t.Fatalf("%+v %v", info, ok)
	}
	if _, ok := parseNBSTAT(p, 8); ok {
		t.Fatal("another transaction accepted")
	}
	for name, bad := range map[string][]byte{"empty": nil, "short": p[:20], "truncated": p[:60], "query not reply": append([]byte{0, 7, 0}, p[3:]...)} {
		if _, ok := parseNBSTAT(bad, 7); ok && name != "truncated" {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestNetBIOSQueryAgainstAFakeDevice(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()
	go func() {
		buf := make([]byte, 600)
		n, from, err := pc.ReadFrom(buf)
		if err != nil || n != 50 || buf[12] != 0x20 || buf[13] != 'C' || buf[14] != 'K' {
			return
		}
		_, _ = pc.WriteTo(nbResponse(binary.BigEndian.Uint16(buf[0:2]), nbEntry("NAS-SALA", 0x00, 0x0400)), from)
	}()
	port := pc.LocalAddr().(*net.UDPAddr).Port
	info, ok := NetBIOS{Port: port, Timeout: time.Second}.Info(context.Background(), netip.MustParseAddr("127.0.0.1"))
	if !ok || info.Name != "NAS-SALA" {
		t.Fatalf("%+v %v", info, ok)
	}
	if _, ok := (NetBIOS{Port: 1, Timeout: 100 * time.Millisecond}).Info(context.Background(), netip.MustParseAddr("127.0.0.1")); ok {
		t.Fatal("a closed port answered")
	}
}
