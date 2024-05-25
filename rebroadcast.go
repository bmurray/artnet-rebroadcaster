package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/bmurray/artnet-rebroadcaster/broadcasters"
	"github.com/jsimonetti/go-artnet"
	"github.com/jsimonetti/go-artnet/packet"
	"github.com/jsimonetti/go-artnet/packet/code"
	"github.com/sirupsen/logrus"
)

type Brodcaster interface {
	Broadcast(ctx context.Context, p *packet.ArtDMXPacket, conn net.Conn) error
}

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	bcast := broadcasters.NewDMXNetworkHost()
	dmxl := broadcasters.NewDMXDirect()
	udmx := broadcasters.NewUSBDMX()

	listenOn := flag.String("listen", "", "UDP port to listen on (eg: 192.168.1.2:6454)")
	listenNetwork := flag.String("network", "", "Network to listen on (eg: 192.168.88.0/24)")
	flag.Var(bcast, "broadcast", "UDP port to broadcast to (eg: 192.168.88.123:6454) (can be specified multiple times)")
	flag.Var(dmxl, "dmx", "DMX port to broadcast to and the universe (optional, defaults to 0) (eg: /dev/ttyUSB0:0) (can be specified multiple times)")
	flag.Var(udmx, "usbdmx", "USB DMX port to broadcast to and the universe (optional, defaults to 0) (eg: 0:0) (can be specified multiple times)")

	flag.Parse()

	logger := logrus.New()
	// logger.SetLevel(logrus.WarnLevel)

	ip, err := getIP(*listenOn, *listenNetwork)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "Error getting IP to listen on: %s\n", err)
		flag.Usage()
		return
	}

	go runNode(ctx, logger, ip, func(ctx context.Context, p *packet.ArtDMXPacket, conn *net.UDPConn) {
		if err := bcast.Broadcast(ctx, p, conn); err != nil {
			logger.WithError(err).Warn("Error broadcasting packet")
		}
		if err := dmxl.Broadcast(ctx, p, conn); err != nil {
			logger.WithError(err).Warn("Error broadcasting packet")
		}
		if err := udmx.Broadcast(ctx, p, conn); err != nil {
			logger.WithError(err).Warn("Error broadcasting packet")
		}

	})

	<-ctx.Done()

}

func getIP(listenOn string, listenNetwork string) (net.IP, error) {
	if listenOn != "" {
		a := net.ParseIP(listenOn)
		if a == nil {
			return nil, fmt.Errorf("invalid IP address")
		}
		return a, nil
	}

	if listenNetwork == "" {
		return nil, fmt.Errorf("must specify either -listen or -network")
	}
	_, listenOnNetwork, err := net.ParseCIDR(listenNetwork)
	if err != nil {
		return nil, err
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("gett interface addresses: %w", err)
	}
	for _, v := range addrs {
		if ipnet, ok := v.(*net.IPNet); ok {
			if ipnet.Contains(listenOnNetwork.IP) {
				return ipnet.IP, nil
			}
		}
	}
	return nil, fmt.Errorf("no matching interface")
}

func runNode(ctx context.Context, logger logrus.FieldLogger, listenOn net.IP, fn func(ctx context.Context, p *packet.ArtDMXPacket, conn *net.UDPConn)) {

	logger.Infof("Starting node on %s", listenOn)
	node := artnet.NewNode("Rebroadcaster", code.StNode, listenOn,
		artnet.NewLogger(logger.WithField("rebroadcaster", "true")),
		artnet.NodeBroadcastAddress(net.UDPAddr{
			IP:   net.IPv4(255, 255, 255, 255),
			Port: packet.ArtNetPort,
		}),
		// artnet.NodeListenIP(listenOn),
	)

	node.RegisterCallback(code.OpDMX, func(p packet.ArtNetPacket) {
		pkt, ok := p.(*packet.ArtDMXPacket)
		if !ok {
			return
		}
		conn := node.Connection()
		fn(ctx, pkt, conn)
	})
	if err := node.Start(); err != nil {
		logger.WithError(err).Fatal("Error starting node")
		return
	}
	defer node.Stop()
	logger.Info("Node started")
	<-ctx.Done()
}
