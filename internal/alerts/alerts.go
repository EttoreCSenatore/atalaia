// Package alerts decides what is worth telling the user after a scan and
// sends it through the configured channels.
package alerts

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/EttoreCSenatore/atalaia/internal/dhcpprobe"
	"github.com/EttoreCSenatore/atalaia/internal/notify"
	"github.com/EttoreCSenatore/atalaia/internal/store"
)

// Setting keys.
const (
	ChannelsKey = "alert_channels" // JSON list of notify.Channel
	NewKey      = "alert_new"      // "1"/"0": alert when an unknown device joins (default on)
	OfflineKey  = "alert_offline"  // "1"/"0": alert when a known device goes offline (default off)
	ARPKey      = "alert_arp"      // "1"/"0": alert on signs of ARP spoofing (default on)
	DHCPKey     = "alert_dhcp"     // "1"/"0": alert when a new DHCP server answers (default on)
	PortsKey    = "alert_ports"    // "1"/"0": alert when a port scan finds newly opened ports (default on)
	MaxChannels = 8
)

// Rules choose what triggers an alert.
type Rules struct {
	NewDevice    bool `json:"new_device"`
	KnownOffline bool `json:"known_offline"`
	ARPAnomaly   bool `json:"arp_anomaly"`
	RogueDHCP    bool `json:"rogue_dhcp"`
	NewPorts     bool `json:"new_ports"`
}

// Deliverer sends one message through one channel.
type Deliverer interface {
	Send(ctx context.Context, c notify.Channel, m notify.Message) error
}

// Service loads the configuration and sends alerts.
type Service struct {
	Store  *store.Store
	Sender Deliverer
	Logf   func(format string, args ...any)
}

func (s *Service) logf(f string, a ...any) {
	if s.Logf != nil {
		s.Logf(f, a...)
		return
	}
	log.Printf(f, a...)
}

// Channels returns the saved channels, secrets included (server side only).
func (s *Service) Channels(ctx context.Context) ([]notify.Channel, error) {
	raw, err := s.Store.Setting(ctx, ChannelsKey, "[]")
	if err != nil {
		return nil, err
	}
	var out []notify.Channel
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("stored alert channels are unreadable: %w", err)
	}
	return out, nil
}

// SaveChannels validates and stores the channels. An empty ID gets a new one.
func (s *Service) SaveChannels(ctx context.Context, chans []notify.Channel) error {
	if len(chans) > MaxChannels {
		return fmt.Errorf("at most %d channels", MaxChannels)
	}
	ids := map[string]bool{}
	for i := range chans {
		if chans[i].ID == "" {
			chans[i].ID = newID()
		}
		if ids[chans[i].ID] {
			return errors.New("duplicate channel id")
		}
		ids[chans[i].ID] = true
		if err := notify.Validate(chans[i]); err != nil {
			return fmt.Errorf("%s: %w", label(chans[i]), err)
		}
	}
	b, err := json.Marshal(chans)
	if err != nil {
		return err
	}
	return s.Store.SetSetting(ctx, ChannelsKey, string(b))
}

func label(c notify.Channel) string {
	if c.Name != "" {
		return c.Name
	}
	return c.Type
}

func newID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Rules returns the saved rules (defaults: new devices on, offline off).
func (s *Service) Rules(ctx context.Context) (Rules, error) {
	n, err := s.Store.SettingInt(ctx, NewKey, 1)
	if err != nil {
		return Rules{}, err
	}
	o, err := s.Store.SettingInt(ctx, OfflineKey, 0)
	if err != nil {
		return Rules{}, err
	}
	a, err := s.Store.SettingInt(ctx, ARPKey, 1)
	if err != nil {
		return Rules{}, err
	}
	d, err := s.Store.SettingInt(ctx, DHCPKey, 1)
	if err != nil {
		return Rules{}, err
	}
	p, err := s.Store.SettingInt(ctx, PortsKey, 1)
	if err != nil {
		return Rules{}, err
	}
	return Rules{NewDevice: n != 0, KnownOffline: o != 0, ARPAnomaly: a != 0, RogueDHCP: d != 0, NewPorts: p != 0}, nil
}

// SaveRules stores the rules.
func (s *Service) SaveRules(ctx context.Context, r Rules) error {
	if err := s.Store.SetSetting(ctx, NewKey, b2s(r.NewDevice)); err != nil {
		return err
	}
	if err := s.Store.SetSetting(ctx, OfflineKey, b2s(r.KnownOffline)); err != nil {
		return err
	}
	if err := s.Store.SetSetting(ctx, ARPKey, b2s(r.ARPAnomaly)); err != nil {
		return err
	}
	if err := s.Store.SetSetting(ctx, DHCPKey, b2s(r.RogueDHCP)); err != nil {
		return err
	}
	return s.Store.SetSetting(ctx, PortsKey, b2s(r.NewPorts))
}

func b2s(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// Compose builds the messages a scan result deserves. baseline is true for the
// very first scan, which finds every device on the network: that is the
// starting inventory, not news.
func Compose(res store.Result, rules Rules, baseline bool) []notify.Message {
	var out []notify.Message
	if rules.NewDevice && !baseline && len(res.New) > 0 {
		title := "Atalaia: new device on the network"
		if len(res.New) > 1 {
			title = fmt.Sprintf("Atalaia: %d new devices on the network", len(res.New))
		}
		out = append(out, notify.Message{Title: title, Body: describe(res.New)})
	}
	if rules.KnownOffline {
		var gone []store.Device
		for _, d := range res.WentOffline {
			if d.Known {
				gone = append(gone, d)
			}
		}
		if len(gone) > 0 {
			out = append(out, notify.Message{Title: "Atalaia: known device went offline", Body: describe(gone)})
		}
	}
	if rules.ARPAnomaly {
		for _, f := range res.Findings {
			switch f.Kind {
			case store.EventGatewayChanged:
				out = append(out, notify.Message{
					Title: "Atalaia: the gateway's MAC address changed",
					Body: "The router's address now answers from a different device. This can be a router replacement, or someone intercepting traffic (ARP spoofing).\n" +
						f.Detail + "\n" + describe([]store.Device{f.Device}),
				})
			case store.EventMultiIP:
				out = append(out, notify.Message{
					Title: "Atalaia: one device answers for several addresses",
					Body: "A single device claims many addresses at once, a common sign of ARP spoofing (it can also be a router or a host with virtual addresses; mark it as known to silence this).\n" +
						f.Detail + "\n" + describe([]store.Device{f.Device}),
				})
			}
		}
	}
	return out
}

func describe(devs []store.Device) string {
	var b strings.Builder
	for i, d := range devs {
		if i == 10 {
			fmt.Fprintf(&b, "… and %d more\n", len(devs)-10)
			break
		}
		name := d.Label
		if name == "" {
			name = d.Name
		}
		if name == "" {
			name = "(no name)"
		}
		vendor := d.Vendor
		if vendor == "" {
			vendor = "unknown maker"
		}
		fmt.Fprintf(&b, "%s — %s — %s — %s", name, vendor, d.IP, d.MAC)
		if d.Randomized {
			b.WriteString(" (random MAC)")
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// Handle is called by the monitor after every recorded scan.
func (s *Service) Handle(ctx context.Context, res store.Result, baseline bool) {
	rules, err := s.Rules(ctx)
	if err != nil {
		s.logf("alerts: %v", err)
		return
	}
	s.send(ctx, Compose(res, rules, baseline))
}

// HandleDHCP alerts about DHCP servers that were not there before.
func (s *Service) HandleDHCP(ctx context.Context, servers []store.DHCPServer) {
	rules, err := s.Rules(ctx)
	if err != nil {
		s.logf("alerts: %v", err)
		return
	}
	s.send(ctx, ComposeDHCP(servers, rules))
}

// HandlePorts alerts about ports that a port scan found open and were not before.
func (s *Service) HandlePorts(ctx context.Context, d store.Device, opened []int) {
	rules, err := s.Rules(ctx)
	if err != nil {
		s.logf("alerts: %v", err)
		return
	}
	s.send(ctx, ComposePorts(d, opened, rules))
}

func (s *Service) send(ctx context.Context, msgs []notify.Message) {
	if len(msgs) == 0 {
		return
	}
	chans, err := s.Channels(ctx)
	if err != nil {
		s.logf("alerts: %v", err)
		return
	}
	for _, c := range chans {
		if !c.Enabled {
			continue
		}
		for _, m := range msgs {
			if err := s.Sender.Send(ctx, c, m); err != nil {
				s.logf("alerts: sending through %q failed: %v", label(c), err)
			}
		}
	}
}

// ComposeDHCP builds the alert for newly seen DHCP servers.
func ComposeDHCP(servers []store.DHCPServer, rules Rules) []notify.Message {
	if !rules.RogueDHCP || len(servers) == 0 {
		return nil
	}
	var b strings.Builder
	for i, sv := range servers {
		if i == 10 {
			fmt.Fprintf(&b, "… and %d more\n", len(servers)-10)
			break
		}
		var o dhcpprobe.Offer
		_ = json.Unmarshal([]byte(sv.Detail), &o)
		fmt.Fprintf(&b, "DHCP server %s", sv.IP)
		if o.Router.IsValid() {
			fmt.Fprintf(&b, " — would hand out gateway %s", o.Router)
		}
		if len(o.DNS) > 0 {
			dns := make([]string, 0, len(o.DNS))
			for _, d := range o.DNS {
				dns = append(dns, d.String())
			}
			fmt.Fprintf(&b, " and DNS %s", strings.Join(dns, ", "))
		}
		b.WriteByte('\n')
	}
	return []notify.Message{{
		Title: "Atalaia: a new DHCP server answered on the network",
		Body:  "A DHCP server that was not there before is offering addresses. If you did not add it (a second router, an access point in router mode), it may be a rogue server redirecting your traffic.\n" + b.String(),
	}}
}

// ComposePorts builds the alert for newly opened ports on a device.
func ComposePorts(d store.Device, opened []int, rules Rules) []notify.Message {
	if !rules.NewPorts || len(opened) == 0 {
		return nil
	}
	ports := make([]string, 0, len(opened))
	for _, p := range opened {
		ports = append(ports, fmt.Sprint(p))
	}
	return []notify.Message{{
		Title: "Atalaia: new open port on a device",
		Body:  "A port scan found ports that were closed before: " + strings.Join(ports, ", ") + "\n" + describe([]store.Device{d}),
	}}
}
