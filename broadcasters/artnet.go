package broadcasters

import (
	"context"
	"net"
	"net/netip"

	"github.com/bmurray/go-artnet/packet"
)

type DMXNetworkHost struct {
	hosts []*net.UDPAddr
}

func NewDMXNetworkHost() *DMXNetworkHost {
	return &DMXNetworkHost{}
}

func (n *DMXNetworkHost) String() string {
	return "Hosts"
}

func (n *DMXNetworkHost) Set(value string) error {
	a, err := netip.ParseAddrPort(value)
	if err != nil {
		// fmt.Fprintf(flag.CommandLine.Output(), "Error parsing %s: %s\n", v, err)
		// flag.Usage()
		return err
	}
	n.hosts = append(n.hosts, net.UDPAddrFromAddrPort(a))
	return nil
}

func (n *DMXNetworkHost) Broadcast(ctx context.Context, pkt *packet.ArtDMXPacket, conn *net.UDPConn) error {
	m, err := pkt.MarshalBinary()
	if err != nil {
		return err
	}
	for _, v := range n.hosts {
		_, err := conn.WriteTo(m, v)
		if err != nil {
			// logger.WithError(err).Warn("Error writing packet")
			return err
		}
	}
	return nil
}
