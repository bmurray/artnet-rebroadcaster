package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/signal"

	"github.com/jsimonetti/go-artnet"
	"github.com/jsimonetti/go-artnet/packet"
	"github.com/jsimonetti/go-artnet/packet/code"
	"github.com/sirupsen/logrus"
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	listenOn := flag.String("listen", "", "UDP port to listen on (eg: 192.168.1.2:6454)")
	listenNetwork := flag.String("network", "", "Network to listen on (eg: 192.168.88.0/24)")
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <transmit to> [<transmit to> ...]\n", os.Args[0])
		flag.Usage()
		return
	}

	broadcastTo := make([]*net.UDPAddr, len(args))

	for i, v := range args {
		a, err := netip.ParseAddrPort(v)
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "Error parsing %s: %s\n", v, err)
			flag.Usage()
			return
		}
		broadcastTo[i] = net.UDPAddrFromAddrPort(a)
	}

	logger := logrus.New()
	// logger.SetLevel(logrus.WarnLevel)

	ip, err := getIP(*listenOn, *listenNetwork)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "Error getting IP to listen on: %s\n", err)
		flag.Usage()
		return
	}

	go runNode(ctx, logger, ip, broadcastTo)

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

func runNode(ctx context.Context, logger logrus.FieldLogger, listenOn net.IP, broadcastTo []*net.UDPAddr) {

	logger.Infof("Starting node on %s", listenOn)
	node := artnet.NewNode("Rebroadcaster", code.StNode, listenOn,
		artnet.NewLogger(logger.WithField("rebroadcaster", "true")),
		artnet.NodeBroadcastAddress(net.UDPAddr{
			IP:   net.IPv4(255, 255, 255, 255),
			Port: packet.ArtNetPort,
		}),
		// artnet.NodeListenIP(listenOn),
	)

	// pt := code.PortType(0).WithInput(true).WithOutput(true).WithType("DMX512")
	// gi := code.

	node.RegisterCallback(code.OpDMX, func(p packet.ArtNetPacket) {
		// logger.Info("Received DMX packet")
		pkt, ok := p.(*packet.ArtDMXPacket)
		if !ok {
			return
		}
		m, err := pkt.MarshalBinary()
		if err != nil {
			logger.WithError(err).Warn("Error marshalling packet")
			return
		}
		conn := node.Connection()
		for _, v := range broadcastTo {
			_, err := conn.WriteTo(m, v)
			if err != nil {
				logger.WithError(err).Warn("Error writing packet")
			}
		}
	})
	if err := node.Start(); err != nil {
		logger.WithError(err).Fatal("Error starting node")
		return
	}
	defer node.Stop()
	logger.Info("Node started")
	<-ctx.Done()
}
