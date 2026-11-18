package fingerprint

import "testing"

func TestGuess(t *testing.T) {
	cases := []struct {
		name string
		in   GuessInput
		want string
	}{
		{"gateway wins over everything", GuessInput{Gateway: true, Services: []string{"_ipp._tcp"}}, KindRouter},
		{"UPnP gateway", GuessInput{DeviceType: "InternetGatewayDevice"}, KindRouter},
		{"printer by service", GuessInput{Services: []string{"_ipp._tcp", "_http._tcp"}}, KindPrinter},
		{"printer by vendor", GuessInput{Vendor: "Brother Industries, Ltd."}, KindPrinter},
		{"chromecast", GuessInput{Services: []string{"_googlecast._tcp"}, Model: "Chromecast Ultra"}, KindTV},
		{"google home", GuessInput{Services: []string{"_googlecast._tcp"}, Model: "Google Home Mini"}, KindSpeaker},
		{"apple tv", GuessInput{Services: []string{"_airplay._tcp", "_companion-link._tcp"}, Model: "AppleTV6,2"}, KindTV},
		{"homepod", GuessInput{Services: []string{"_airplay._tcp"}, Model: "AudioAccessory5,1"}, KindSpeaker},
		{"iphone by name", GuessInput{Name: "Marias-iPhone.local", Randomized: true, Services: []string{"_companion-link._tcp"}}, KindPhone},
		{"unnamed apple device", GuessInput{Services: []string{"_companion-link._tcp"}, Randomized: true}, KindMobile},
		{"samsung tv", GuessInput{DeviceType: "MediaRenderer", Vendor: "Samsung Electronics"}, KindTV},
		{"sonos", GuessInput{DeviceType: "ZonePlayer"}, KindSpeaker},
		{"raspberry pi", GuessInput{Vendor: "Raspberry Pi Trading Ltd"}, KindServer},
		{"esp32", GuessInput{Vendor: "Espressif Inc."}, KindIoT},
		{"homekit", GuessInput{Services: []string{"_hap._tcp"}}, KindIoT},
		{"access point", GuessInput{Vendor: "Ubiquiti Inc"}, KindNetwork},
		{"windows pc", GuessInput{NetBIOS: true}, KindComputer},
		{"nas", GuessInput{Name: "DiskStation", NetBIOS: true}, KindNAS},
		{"random MAC and nothing else", GuessInput{Randomized: true}, KindMobile},
		{"a name that merely contains nas", GuessInput{Name: "Anastasia", Randomized: true}, KindMobile},
		{"a name that contains tv in a word", GuessInput{Name: "Settlement", Randomized: true}, KindMobile},
		{"nothing to go on", GuessInput{Vendor: "Some Unknown Co"}, ""},
	}
	for _, c := range cases {
		if got := Guess(c.in); got != c.want {
			t.Errorf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}

func TestEveryKindIsListed(t *testing.T) {
	listed := map[string]bool{}
	for _, k := range Kinds {
		listed[k] = true
	}
	for _, k := range []string{KindRouter, KindNetwork, KindComputer, KindMobile, KindPhone, KindTablet, KindTV, KindSpeaker, KindPrinter, KindCamera, KindNAS, KindConsole, KindIoT, KindServer} {
		if !listed[k] {
			t.Errorf("%s missing from Kinds", k)
		}
	}
}
