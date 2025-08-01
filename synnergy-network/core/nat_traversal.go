package core

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/huin/goupnp/dcps/internetgateway1"
	"github.com/jackpal/gateway"
	natpmp "github.com/jackpal/go-nat-pmp"
)

// NATManager manages NAT traversal using NAT-PMP or UPnP.
type NATManager struct {
	ip         net.IP
	pmp        *natpmp.Client
	upnp       *internetgateway1.WANIPConnection1
	mappedPort int
}

// NewNATManager discovers the gateway and external IP.
func NewNATManager() (*NATManager, error) {
	m := &NATManager{}
	if gw, err := gateway.DiscoverGateway(); err == nil {
		m.pmp = natpmp.NewClient(gw)
		if res, err := m.pmp.GetExternalAddress(); err == nil {
			m.ip = net.IPv4(res.ExternalIPAddress[0], res.ExternalIPAddress[1], res.ExternalIPAddress[2], res.ExternalIPAddress[3])
		}
	}
	if m.ip == nil {
		if clients, _, err := internetgateway1.NewWANIPConnection1Clients(); err == nil && len(clients) > 0 {
			m.upnp = clients[0]
			if ipStr, err := m.upnp.GetExternalIPAddress(); err == nil {
				m.ip = net.ParseIP(ipStr)
			}
		}
	}
	if m.ip == nil {
		return nil, fmt.Errorf("nat_traversal: gateway not found")
	}
	return m, nil
}

// ExternalIP returns the detected public IP address.
func (m *NATManager) ExternalIP() net.IP { return m.ip }

// Map opens the given TCP port on the gateway.
func (m *NATManager) Map(port int) error {
	if m.pmp != nil {
		if _, err := m.pmp.AddPortMapping("tcp", port, port, 3600); err == nil {
			m.mappedPort = port
			return nil
		}
	}
	if m.upnp != nil {
		if err := m.upnp.AddPortMapping("", uint16(port), "TCP", uint16(port), m.ip.String(), true, "synnergy", 3600); err == nil {
			m.mappedPort = port
			return nil
		}
	}
	return fmt.Errorf("nat_traversal: mapping failed")
}

// Unmap removes the previously mapped port.
func (m *NATManager) Unmap() error {
	if m.mappedPort == 0 {
		return nil
	}
	if m.pmp != nil {
		if _, err := m.pmp.AddPortMapping("tcp", m.mappedPort, m.mappedPort, 0); err != nil {
			return err
		}
		m.mappedPort = 0
		return nil
	}
	if m.upnp != nil {
		if err := m.upnp.DeletePortMapping("", uint16(m.mappedPort), "TCP"); err != nil {
			return err
		}
		m.mappedPort = 0
	}
	return nil
}

// parsePort extracts the TCP port from a libp2p multiaddress string.
func parsePort(addr string) (int, error) {
	parts := strings.Split(addr, "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "tcp" {
			return strconv.Atoi(parts[i+1])
		}
	}
	return 0, fmt.Errorf("no tcp port in %s", addr)
}
