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
	Query(host string, requestType dnsmessage.Type) ([]dnsmessage.Resource, dnsmessage.RCode)
}

type DNSService struct {
	conn  *net.UDPConn
	cache store
}

type Packet struct {
	addr    net.UDPAddr
	message dnsmessage.Message
}

func (svc *DNSService) Query(host string, requestType dnsmessage.Type) (answerResources []dnsmessage.Resource, dnsStatusCode dnsmessage.RCode) {
	dnsStatusCode = dnsmessage.RCodeSuccess
	cacheKey := host + requestType.String()
	resp, ok := svc.cache.get(cacheKey)
	if !ok {
		log.Printf("no cached record for %s, fetching...\n", host)
		resp, dnsStatusCode, err := DoForwarderRequest(host, requestType)
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
		}
	}

	return
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

		question := m.Questions[0]
		questionName := question.Name.String()
		requestType := question.Type

		answerResources, responseStatusCode := svc.Query(questionName, requestType)
		m.Header.RCode = responseStatusCode

		go svc.sendPacket(Packet{
			addr: *addr,
			message: dnsmessage.Message{
				Header:      m.Header,
				Questions:   m.Questions,
				Answers:     answerResources,
				Authorities: m.Authorities,
				Additionals: m.Additionals,
			},
		})
	}
}

func (svc *DNSService) sendPacket(p Packet) {
	packed, err := p.message.Pack()
	if err != nil {
		log.Println("error packing", err)
		return
	}
	_, err = svc.conn.WriteToUDP(packed, &p.addr)
	if err != nil {
		log.Println("error sending to socket", err)
	}
}
