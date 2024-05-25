package broadcasters

import (
	"context"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/jsimonetti/go-artnet/packet"
	"github.com/oliread/usbdmx"
	"github.com/oliread/usbdmx/ft232"
)

// vendor and product IDs for the USB DMX device
var vid = uint16(0x0403)
var pid = uint16(0x6001)

type usbdmxHost struct {
	universe   uint8
	controller *ft232.DMXController
}

type USBDMX struct {
	dmx []*usbdmxHost
}

func NewUSBDMX() *USBDMX {
	return &USBDMX{}
}

func (n *USBDMX) String() string {
	return "USB DMX Hosts"
}

func (n *USBDMX) Set(value string) error {
	s := strings.Split(value, ":")
	var universe uint8 = 0
	if len(s) == 2 {
		u, err := strconv.Atoi(s[1])
		if err != nil {
			return err
		}
		universe = uint8(u)
	}
	outputInterfaceId, err := strconv.Atoi(s[0])
	if err != nil {
		return err
	}

	config := usbdmx.NewConfig(vid, pid, outputInterfaceId, 0, 420)
	config.GetUSBContext()

	controller := ft232.NewDMXController(config)
	if err := controller.Connect(); err != nil {
		return err
	}

	n.dmx = append(n.dmx, &usbdmxHost{
		universe:   universe,
		controller: &controller,
	})

	return nil
}

func (n *USBDMX) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Millisecond):
			for _, v := range n.dmx {
				if err := v.controller.Render(); err != nil {
					log.Fatalf("Failed to render output: %s", err)
				}
			}
		}
	}

}

func (n *USBDMX) Broadcast(ctx context.Context, pkt *packet.ArtDMXPacket, conn *net.UDPConn) error {

	for _, v := range n.dmx {
		if v.universe != pkt.SubUni {
			continue
		}
		for i := 0; i < 512; i++ {
			v.controller.SetChannel(int16(i+1), pkt.Data[i])
		}
	}

	return nil

}
