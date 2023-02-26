package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
	"golang.org/x/net/ipv4"
)

var data []byte

var (
	rootCmd = &cobra.Command{
		Use:   "multicast-sender",
		Short: "A tool to send multicast packets",
		Long:  `A simple tool to send multicast packets with customizable parameters`,
		Run:   sendPackets,
	}
	address        string
	port           int
	intf           string
	rate           int
	ttl            int
	enableTimeDiff bool
	packetLen      int
)

func init() {
	rootCmd.Flags().StringVarP(&address, "address", "a", "", "Multicast address")
	rootCmd.Flags().IntVarP(&port, "port", "p", 0, "Multicast port")
	rootCmd.Flags().StringVarP(&intf, "interface", "i", "", "Egress Interface name, if not specified, will use route lookup to decide the interface name")
	rootCmd.Flags().IntVarP(&rate, "rate", "r", 1, "Packet rate per second")
	rootCmd.Flags().IntVarP(&ttl, "ttl", "t", 16, "TTL value")
	rootCmd.Flags().BoolVar(&enableTimeDiff, "time-diff", false, "Enable time difference")
	rootCmd.Flags().IntVarP(&packetLen, "len", "l", 8, "Packet payload length")
	rootCmd.MarkFlagRequired("address")
	rootCmd.MarkFlagRequired("port")
}

func isMulticastAddress(ip net.IP) bool {
	// Create the netmask for multicast addresses (11110000 00000000 00000000 00000000)
	multicastNetmask := net.IPv4Mask(240, 0, 0, 0)

	// Perform a bitwise AND between the IP address and the multicast netmask
	masked := ip.Mask(multicastNetmask)

	// Check if the result of the bitwise AND is equal to 224.0.0.0
	return masked.Equal(net.IPv4(224, 0, 0, 0))
}

func sendPackets(cmd *cobra.Command, args []string) {
	multicastIP := net.ParseIP(address)
	if multicastIP == nil || !isMulticastAddress(multicastIP) {
		log.Fatalf("Invalid multicast address: %s\n", address)
	}

	var interFace *net.Interface
	var err error

	if intf == "" {
		// perform route lookup
		routes, err := netlink.RouteGet(multicastIP)
		if err != nil {
			fmt.Println("Failed to get routes:", err)
			return
		}

		interFace, err = net.InterfaceByIndex(routes[0].LinkIndex)
		if err != nil {
			fmt.Println("Error getting interface:", err)
			return
		}

	} else {
		interFace, err = net.InterfaceByName(intf)
		if err != nil {
			log.Fatal("unable to get interface")
		}
	}

	addr := &net.UDPAddr{
		IP:   multicastIP,
		Port: port,
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("ListenMulticastUDP error: %v\n", err)
	}
	defer conn.Close()

	p := ipv4.NewPacketConn(conn)
	p.SetTTL(ttl)
	fmt.Println("Sending packets to", addr.String(), "port", port, "on", interFace.Name)
	dst := &net.UDPAddr{IP: addr.IP, Port: port}
	for {
		for _, ifi := range []*net.Interface{interFace} {
			if err := p.SetMulticastInterface(ifi); err != nil {
				log.Println("error setting multicast interface")
			}
			p.SetMulticastTTL(ttl)
			if enableTimeDiff {
				Payload := time.Now().UnixNano()
				PayloadString := strconv.FormatInt(Payload, 10)
				data = []byte(PayloadString)
			} else {
				data = make([]byte, packetLen)
				rand.Read(data)
			}
			if _, err := p.WriteTo(data, nil, dst); err != nil {
				log.Println("error write multicast", err)
			}
		}
		time.Sleep(time.Millisecond * time.Duration(1000/rate))
		fmt.Print(".")
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
