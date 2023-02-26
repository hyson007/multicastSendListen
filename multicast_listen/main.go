package main

import (
	"fmt"
	"log"
	"net"

	//	"syscall"

	//	"golang.org/x/sys/unix"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
	//	"golang.org/x/net/ipv4"
)

var (
	rootCmd = &cobra.Command{
		Use:   "multicast-listener",
		Short: "A UDP Multicast listener",
		Long:  `A simple tool to listen multicast packets with customizable parameters`,
		Run:   receivePackets,
	}
	address        string
	port           int
	intf           string
	printPayload   bool
	enableTimeDiff bool
)

func init() {
	rootCmd.Flags().StringVarP(&address, "address", "a", "", "Multicast address")
	rootCmd.Flags().IntVarP(&port, "port", "p", 0, "Multicast port")
	rootCmd.Flags().StringVarP(&intf, "interface", "i", "", "Interface name")
	rootCmd.Flags().BoolVar(&printPayload, "payload", false, "Print payload")
	rootCmd.Flags().BoolVar(&enableTimeDiff, "time-diff", false, "Enable time difference")
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

func receivePackets(cmd *cobra.Command, args []string) {
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
	log.Println(interFace)
	addr := &net.UDPAddr{
		IP:   multicastIP,
		Port: port,
	}
	conn, err := net.ListenMulticastUDP("udp", interFace, addr)
	if err != nil {
		log.Fatalf("ListenMulticastUDP error: %v\n", err)
	}
	defer conn.Close()

	//	file, err := conn.File()
	//	if err!= nil {
	//		log.Fatal("unable to get fd", err)
	//	}
	//	fd := file.Fd()
	//	log.Println(fd)
	// Set the IP_FREEBIND socket option to allow binding to an address that is already in use
	//	err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.IP_FREEBIND, 1)
	//	if err != nil {
	//		fmt.Println("error enabling IP_FREEBIND")
	//		return
	//	}

	//	err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	//	if err != nil {
	//		fmt.Printf("Error enabling SO_REUSEADDR option: %v\n", err)
	//		return
	//	}

	//	p := ipv4.NewPacketConn(conn)
	//	fmt.Println(addr)
	//	if err := p.JoinGroup(interFace, addr); err != nil {
	//		log.Fatal("unable to join group, err: ", err)
	//		return
	//	}

	// // test
	// if loop, err := p.MulticastLoopback(); err == nil {
	// 	log.Printf("MulticastLoopback status:%v\n", loop)
	// 	if !loop {
	// 		if err := p.SetMulticastLoopback(true); err != nil {
	// 			log.Printf("SetMulticastLoopback error:%v\n", err)
	// 		}
	// 	}
	// }

	//	if err := p.SetControlMessage(ipv4.FlagDst, true); err != nil {
	//		log.Fatal("unable to set dst-mode on accept:", err)
	//	}

	//	fmt.Println("hit")
	buf := make([]byte, 1500)
	fmt.Println("listening on interface:", "group:", address, "port:", port, "on interface:", interFace.Name)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			continue
		}
		fmt.Println("Received data:", string(buf[:n]))
		//n, cm, src, err := p.ReadFrom(b)
		//		if err != nil {
		//			log.Fatal("unable to read from packet, err: ", err)
		//		}
		//		log.Println("hit")
		//		if cm.Dst.IsMulticast() {
		//			if cm.Dst.Equal(net.IP(address)) {
		//				// joined group, do something
		//				if enableTimeDiff {
		//					senderNano, err := strconv.Atoi(string(b[:n]))
		//					if err != nil {
		//						log.Println("error converting sender time to Nano, please enable -time-diff on sender")
		//						continue
		//					}
		//					diffInNano := time.Now().UnixNano() - int64(senderNano)
		//					log.Println("diffInMicroSeconds: ", diffInNano/1000)
		//					continue
		//				}
		//
		//				if printPayload {
		//					log.Println("received multicast packet from: ", src, "data: ", string(b[:n]))
		//				} else {
		//					log.Print(".")
		//				}
		//
		///			} else {
		//				// unknown group, discard
		//				continue
		//			}
		//		}
		//	}

	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
