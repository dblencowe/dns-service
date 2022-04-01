package main

import (
	"log"
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

const (
	udpPort   int = 53
	packetLen int = 512
)

func main() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: udpPort})
	chk(err)
	log.Printf("[INFO] listening on port %d\n", udpPort)
	defer conn.Close()

	var lastAddress *net.UDPAddr
	for {
		buf := make([]byte, packetLen)
		_, addr, err := conn.ReadFromUDP(buf)
		chk(err)
		var m dnsmessage.Message
		err = m.Unpack(buf)
		chk(err)
		if m.Header.Response && lastAddress != nil {
			packed, err := m.Pack()
			chk(err)
			_, err = conn.WriteToUDP(packed, lastAddress)
			chk(err)
		} else {
			lastAddress = addr
			go doQuery(conn, m)
		}
	}
}

func doQuery(conn *net.UDPConn, m dnsmessage.Message) {
	packed, err := m.Pack()
	chk(err)
	resolver := net.UDPAddr{IP: net.IP{1, 1, 1, 1}, Port: 53}
	_, err = conn.WriteToUDP(packed, &resolver)
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
