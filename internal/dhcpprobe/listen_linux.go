//go:build linux

package dhcpprobe

import (
	"context"
	"net"
	"strconv"
	"syscall"
)

// Listen opens the client socket on UDP port (68 in production). That port is
// privileged on Linux: without root or CAP_NET_BIND_SERVICE this fails with a
// permission error (see IsPermission). SO_REUSEADDR lets it coexist with the
// machine's own DHCP client, and broadcast offers reach every such socket.
func Listen(port int) (*net.UDPConn, error) {
	lc := net.ListenConfig{Control: func(_, _ string, c syscall.RawConn) error {
		var serr error
		if err := c.Control(func(fd uintptr) {
			serr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
		}); err != nil {
			return err
		}
		return serr
	}}
	pc, err := lc.ListenPacket(context.Background(), "udp4", net.JoinHostPort("0.0.0.0", strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}
	return pc.(*net.UDPConn), nil
}
