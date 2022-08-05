package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"golang.org/x/net/ipv4"
)

func main() {
	var address string
	flag.StringVar(&address, "address", "224.100.100.100", "Multicast address")
	var port int
	flag.IntVar(&port, "port", 8080, "Multicast port")
	var intf string
	flag.StringVar(&intf, "interface", "en0", "Interface name")
	var rate int
	flag.IntVar(&rate, "rate", 1, "packet rate per second")
	var ttl int
	flag.IntVar(&ttl, "ttl", 16, "TTL value")
	flag.Parse()

	interFace, err := net.InterfaceByName(intf)
	if err != nil {
		log.Fatal("unable to get interface, err", err)
	}
	group := net.ParseIP(address)
	c, err := net.ListenPacket("udp4", fmt.Sprint(group.String(), ":", port))
	if err != nil {
		log.Fatal("unable to listen packet, err: ", err)
	}
	defer c.Close()

	p := ipv4.NewPacketConn(c)
	p.SetTTL(64)
	// if _, err := p.WriteTo(data, nil, src); err != nil {
	// 	// error handling
	// }

	dst := &net.UDPAddr{IP: group, Port: port}
	for {
		for _, ifi := range []*net.Interface{interFace} {
			if err := p.SetMulticastInterface(ifi); err != nil {
				log.Println("error setting multicast interface")
			}
			p.SetMulticastTTL(ttl)
			Payload := time.Now().UnixNano()
			PayloadString := strconv.FormatInt(Payload, 10)
			data := []byte(PayloadString)
			if _, err := p.WriteTo(data, nil, dst); err != nil {
				log.Println("error write multicast")
			}
		}
		time.Sleep(time.Millisecond * time.Duration(1000/rate))
		fmt.Print(".")
	}

}
