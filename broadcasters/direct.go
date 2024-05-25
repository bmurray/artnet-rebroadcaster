package broadcasters

import (
	"context"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/bmurray/artnet-rebroadcaster/dmx"
	"github.com/jsimonetti/go-artnet/packet"
)

type dmxHost struct {
	connection *dmx.DMX
	universe   uint8
}
type DMXDirect struct {
	dmx []*dmxHost
}

func NewDMXDirect() *DMXDirect {
	return &DMXDirect{}
}

func (n *DMXDirect) String() string {
	return "DMX Hosts"
}

func (n *DMXDirect) Set(value string) error {
	s := strings.Split(value, ":")
	var universe uint8 = 0
	if len(s) == 2 {
		u, err := strconv.Atoi(s[1])
		if err != nil {
			return err
		}
		universe = uint8(u)
	}
	host, err := dmx.NewDMXConnection(s[0])
	if err != nil {
		return err
	}

	n.dmx = append(n.dmx, &dmxHost{
		connection: host,
		universe:   universe,
	})
	return nil
}

func (n *DMXDirect) Broadcast(ctx context.Context, pkt *packet.ArtDMXPacket, conn *net.UDPConn) error {

	// log.Println("Broadcasting DMX packet", len(n.dmx))
	for _, v := range n.dmx {
		if v.universe != pkt.SubUni {
			log.Println("Skipping universe", v.universe, pkt.Net, pkt.SubUni)
			continue
		}
		for i := 301; i < 301+9; i++ {
			log.Println("Setting channel", i, pkt.Data[i-1])
		}
		// for i, d := range pkt.Data {
		if err := v.connection.Send(pkt.Data); err != nil {
			return err
		}
		// }

	}
	return nil

}
