package service

import (
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
	if err != nil {
		panic(err)
	}
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
		go svc.handleQuestion(Packet{
			addr:    *addr,
			message: m,
		})
		log.Printf("handling question for %s: %+v\n", addr, m)
	}
}

func (svc *DNSService) handleQuestion(p Packet) {
	var dnsStatusCode dnsmessage.RCode = dnsmessage.RCodeSuccess
	var answerResources []dnsmessage.Resource
	question := p.message.Questions[0]
	questionName := question.Name.String()
	requestType := question.Type
	cacheKey := questionName + requestType.String()
	resp, ok := svc.cache.get(cacheKey)
	if !ok {
		log.Printf("no cached record for %s, fetching...\n", question)
		resp, dnsStatusCode, err := DoForwarderRequest(questionName, requestType)
		log.Printf("fetched result from forwarder: %d(%+v)", dnsStatusCode, resp)
		if err == nil {
			svc.cache.set(cacheKey, *resp)
		}
	}
	if dnsStatusCode == dnsmessage.RCodeSuccess {
		resource, err := resp.ToResource()
		if err == nil {
			answerResources = append(answerResources, resource)
		} else {
			log.Printf("error: %+v\n", err)
			return
		}
	}

	message := dnsmessage.Message{
		Header:      p.message.Header,
		Questions:   p.message.Questions,
		Answers:     answerResources,
		Authorities: p.message.Authorities,
		Additionals: p.message.Additionals,
	}

	log.Printf("sending response: %+v\n", message)

	go doQuery(svc.conn, Packet{
		addr:    p.addr,
		message: message,
	})
}

func doQuery(conn *net.UDPConn, p Packet) {
	packed, err := p.message.Pack()
	if err != nil {
		log.Println("error packing", err)
		return
	}
	_, err = conn.WriteToUDP(packed, &p.addr)
	if err != nil {
		log.Println("error sending to socket", err)
	}
}
