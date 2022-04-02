package service

import (
	"fmt"
	"log"
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

const (
	udpPort   int = 53
	packetLen int = 512
)

type DNSServer interface {
	Listen()
}

type DNSService struct {
	conn  *net.UDPConn
	cache store
}

type Packet struct {
	addr    net.UDPAddr
	message dnsmessage.Message
}

func (svc *DNSService) Listen() {
	var err error
	svc.conn, err = net.ListenUDP("udp", &net.UDPAddr{Port: udpPort})
	chk(err)
	svc.cache.data = make(map[string]entry)
	log.Printf("[INFO] listening on port %d\n", udpPort)
	defer svc.conn.Close()

	for {
		buf := make([]byte, packetLen)
		_, addr, err := svc.conn.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		var m dnsmessage.Message
		err = m.Unpack(buf)
		if err != nil {
			log.Println(err)
			continue
		}

		if len(m.Questions) == 0 {
			continue
		}

		log.Printf("received new question: %+v\n", m)
		question := m.Questions[0].Name.String()
		resp, ok := svc.cache.get(question)
		if !ok {
			log.Printf("no cached record for %s, fetching...\n", question)
			resp, err = DoForwarderRequest(question)
			chk(err)
			svc.cache.set(question, *resp)
		}

		answer, err := toResource(resp)
		chk(err)

		go doQuery(svc.conn, Packet{
			addr: *addr,
			message: dnsmessage.Message{
				Header:      m.Header,
				Questions:   m.Questions,
				Answers:     []dnsmessage.Resource{answer},
				Authorities: m.Authorities,
				Additionals: m.Additionals,
			},
		})
	}
}

func doQuery(conn *net.UDPConn, p Packet) {
	packed, err := p.message.Pack()
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = conn.WriteToUDP(packed, &p.addr)
	if err != nil {
		fmt.Println(err)
	}
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
