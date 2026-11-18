//go:build !linux

package tools

import (
	"context"
	"errors"
	"net/netip"
	"time"
)

// Hop is one step on the way to a host.
type Hop struct {
	TTL     int     `json:"ttl"`
	Addr    string  `json:"addr"`
	RTTms   float64 `json:"rtt_ms"`
	Reached bool    `json:"reached"`
}

// Traceroute is only implemented on Linux.
func Traceroute(context.Context, netip.Addr, int, time.Duration) ([]Hop, error) {
	return nil, errors.New("traceroute is supported on Linux only")
}
