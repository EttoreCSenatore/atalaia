package fingerprint

import (
	"context"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/scope"
)

// Device kinds Guess can return ("" when it has no idea).
const (
	KindRouter   = "router"
	KindNetwork  = "network" // access point, switch, mesh node
	KindComputer = "computer"
	KindMobile   = "mobile" // phone, tablet or laptop using a private (random) MAC
	KindPhone    = "phone"
	KindTablet   = "tablet"
	KindTV       = "tv"
	KindSpeaker  = "speaker"
	KindPrinter  = "printer"
	KindCamera   = "camera"
	KindNAS      = "nas"
	KindConsole  = "console"
	KindIoT      = "iot"
	KindServer   = "server"
)

// Kinds lists every kind, for the web app and for tests.
var Kinds = []string{KindRouter, KindNetwork, KindComputer, KindMobile, KindPhone, KindTablet, KindTV, KindSpeaker, KindPrinter, KindCamera, KindNAS, KindConsole, KindIoT, KindServer}

// Result gathers everything learned about one address.
type Result struct {
	Services     []string
	Model        string
	Manufacturer string
	Friendly     string
	DeviceType   string
	NetBIOS      NetBIOSInfo
	HasNetBIOS   bool
}

// Gather asks every address in ips (mDNS, NetBIOS) and the multicast group
// (SSDP) at the same time, within about `budget`. Addresses outside sc are ignored.
func Gather(ctx context.Context, ips []netip.Addr, sc *scope.Scope, budget time.Duration) map[netip.Addr]Result {
	ctx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()
	out := map[netip.Addr]Result{}
	var mu sync.Mutex
	set := func(ip netip.Addr, f func(*Result)) {
		mu.Lock()
		r := out[ip]
		f(&r)
		out[ip] = r
		mu.Unlock()
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		found, _ := SSDP{Scope: sc}.Discover(ctx)
		for ip, s := range found {
			set(ip, func(r *Result) {
				r.Manufacturer, r.DeviceType = s.Manufacturer, s.DeviceType
				r.Friendly = s.Friendly
				if r.Model == "" {
					r.Model = s.Model
				}
			})
		}
	}()
	sem := make(chan struct{}, 16)
	for _, ip := range ips {
		if sc != nil && !sc.Allows(ip) {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			m := MDNS{}.Info(ctx, ip)
			nb, ok := NetBIOS{}.Info(ctx, ip)
			set(ip, func(r *Result) {
				r.Services = m.Services
				if m.Model != "" {
					r.Model = m.Model
				}
				r.NetBIOS, r.HasNetBIOS = nb, ok
			})
		}()
	}
	wg.Wait()
	return out
}

// GuessInput is what Guess looks at.
type GuessInput struct {
	Name, Vendor, Model string
	Services            []string
	DeviceType          string
	NetBIOS             bool
	Randomized, Gateway bool
}

func has(s string, words ...string) bool {
	s = strings.ToLower(s)
	for _, w := range words {
		if strings.Contains(s, w) {
			return true
		}
	}
	return false
}

// word reports whether any of words is a whole token of s (split on anything
// that is not a letter or digit), so "nas" matches "my-nas" but not "Anastasia".
func word(s string, words ...string) bool {
	toks := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9') })
	for _, t := range toks {
		for _, w := range words {
			if t == w {
				return true
			}
		}
	}
	return false
}

// Guess names the most likely kind of device. It is a heuristic: strong,
// specific signals (the gateway, a UPnP device type, a printer service) come
// before weak ones (a name, a vendor), and with nothing to go on it says "".
func Guess(in GuessInput) string {
	svc := map[string]bool{}
	for _, s := range in.Services {
		svc[s] = true
	}
	text := in.Name + " " + in.Model + " " + in.Vendor
	switch {
	case in.Gateway || in.DeviceType == "InternetGatewayDevice":
		return KindRouter
	case svc["_ipp._tcp"] || svc["_printer._tcp"] || svc["_pdl-datastream._tcp"] || svc["_scanner._tcp"] || in.DeviceType == "Printer" || in.DeviceType == "PrintBasic":
		return KindPrinter
	case in.DeviceType == "ZonePlayer" || has(text, "sonos", "homepod", "alexa", "soundbar", "speaker") || word(text, "echo"):
		return KindSpeaker
	case svc["_googlecast._tcp"]:
		if has(in.Model, "home", "nest", "mini", "audio") {
			return KindSpeaker
		}
		return KindTV
	case svc["_airplay._tcp"]:
		if has(in.Model, "audioaccessory", "homepod") {
			return KindSpeaker
		}
		if has(in.Model, "appletv") || word(in.Model, "tv") || !svc["_companion-link._tcp"] {
			return KindTV
		}
	case in.DeviceType == "MediaRenderer" || in.DeviceType == "MediaServer":
		if word(text, "tv", "lg") || has(text, "bravia", "samsung", "webos", "roku", "vizio") {
			return KindTV
		}
		return KindSpeaker
	case svc["_spotify-connect._tcp"] || svc["_raop._tcp"] && !svc["_companion-link._tcp"] && !svc["_airplay._tcp"]:
		return KindSpeaker
	case has(text, "camera", "ipcam", "hikvision", "dahua", "reolink", "wyze", "arlo") || word(text, "cam"):
		return KindCamera
	case has(text, "playstation", "xbox", "nintendo"):
		return KindConsole
	case has(text, "iphone"):
		return KindPhone
	case has(text, "ipad"):
		return KindTablet
	case has(text, "android", "galaxy", "pixel", "redmi", "xiaomi", "oneplus", "motorola"):
		return KindPhone
	case has(text, "macbook", "imac", "laptop", "desktop", "notebook", "thinkpad") || word(text, "pc", "mac"):
		return KindComputer
	case has(text, "synology", "qnap", "diskstation", "readynas", "wd my cloud") || word(text, "nas") || in.NetBIOS && svc["_smb._tcp"] && !has(text, "desktop", "laptop"):
		return KindNAS
	case has(in.Vendor, "raspberry"):
		return KindServer
	case has(in.Vendor, "espressif", "tuya", "shelly", "sonoff", "philips lighting", "signify", "ikea", "xiaomi communications", "amazon technologies", "lifx", "tp-link smart"):
		return KindIoT
	case svc["_hap._tcp"] || svc["_homekit._tcp"] || svc["_matter._tcp"] || svc["_hue._tcp"]:
		return KindIoT
	case has(in.Vendor, "ubiquiti", "tp-link", "netgear", "mikrotik", "aruba", "cisco", "zyxel", "d-link", "asustek", "linksys", "eero", "huawei technologies") && !svc["_airplay._tcp"]:
		return KindNetwork
	case has(in.Vendor, "brother", "canon", "epson", "lexmark", "xerox", "kyocera"):
		return KindPrinter
	case has(in.Vendor, "roku", "vizio", "tcl", "hisense", "sony interactive"):
		return KindTV
	case svc["_companion-link._tcp"] || svc["_rdlink._tcp"] || svc["_apple-mobdev2._tcp"]:
		return KindMobile
	case in.NetBIOS || svc["_workstation._tcp"] || svc["_ssh._tcp"] || svc["_smb._tcp"] || svc["_afpovertcp._tcp"]:
		return KindComputer
	case in.Randomized:
		return KindMobile
	}
	return ""
}
