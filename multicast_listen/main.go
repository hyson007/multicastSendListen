package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"golang.org/x/net/ipv4"
)

func main() {
	var address string
	flag.StringVar(&address, "address", "224.100.100.100", "Multicast address")
	var port int
	flag.IntVar(&port, "port", 8080, "Multicast port")
	var intf string
	flag.StringVar(&intf, "interface", "en0", "Interface name")
	flag.Parse()

	en0, err := net.InterfaceByName(intf)
	if err != nil {
		log.Fatal("unable to get interface, err", err)
	}

	group := net.ParseIP(address)
	c, err := net.ListenPacket("udp4", fmt.Sprintf("%s:%d", group.String(), port))
	if err != nil {
		log.Fatal("unable to listen packet, err: ", err)
	}
	defer c.Close()

	// c2, err := net.ListenPacket("udp4", "224.0.0.0:20000")
	// if err != nil {
	// 	log.Fatal("unable to listen packet, err: ", err)
	// }
	// defer c2.Close()

	// c1, err := net.ListenPacket("udp4", "224.0.0.0:1024")
	// if err != nil {
	// 	// error handling
	// }
	// defer c1.Close()
	// c2, err := net.ListenPacket("udp4", "224.0.0.0:1024")
	// if err != nil {
	// 	// error handling
	// }
	// defer c2.Close()

	p := ipv4.NewPacketConn(c)
	if err := p.JoinGroup(en0, &net.UDPAddr{IP: group}); err != nil {
		log.Fatal("unable to join group, err: ", err)
	}

	// test
	if loop, err := p.MulticastLoopback(); err == nil {
		fmt.Printf("MulticastLoopback status:%v\n", loop)
		if !loop {
			if err := p.SetMulticastLoopback(true); err != nil {
				fmt.Printf("SetMulticastLoopback error:%v\n", err)
			}
		}
	}

	if err := p.SetControlMessage(ipv4.FlagDst, true); err != nil {
		log.Fatal("unable to set dst-mode on accept:", err)
	}

	// if _, err := c.WriteTo([]byte("hello"), &net.UDPAddr{IP: group, Port: 1024}); err != nil {
	// 	fmt.Printf("Write failed, %v\n", err)
	// }

	b := make([]byte, 1500)
	fmt.Println("listening on interface:", en0.Name, "group:", group, "port:", port)
	for {
		n, cm, src, err := p.ReadFrom(b)
		if err != nil {
			log.Fatal("unable to read from packet, err: ", err)
		}
		if cm.Dst.IsMulticast() {
			if cm.Dst.Equal(group) {
				// joined group, do something
				log.Println("received multicast packet from: ", src, "data: ", string(b[:n]))
			} else {
				// unknown group, discard
				continue
			}
		}
	}
}
