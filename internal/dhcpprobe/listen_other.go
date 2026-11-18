//go:build !linux

package dhcpprobe

import (
	"errors"
	"net"
)

// Listen is only implemented on Linux.
func Listen(int) (*net.UDPConn, error) {
	return nil, errors.New("the DHCP check is supported on Linux only")
}
