package service

import (
	"log"
	"net"
	"net/http"

	"github.com/dblencowe/dns-service/providers"
	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/net/http2"
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
	conn       *net.UDPConn
	cache      store
	httpClient *http.Client
	provider   *providers.CloudflareHttpsDNSProvider
}

type Packet struct {
	addr    net.UDPAddr
	message dnsmessage.Message
}

func InitDNSService(provider providers.CloudflareHttpsDNSProvider) (svc DNSService) {
	var err error
	svc.conn, err = net.ListenUDP("udp", &net.UDPAddr{Port: udpPort})
	log.Printf("[INFO] listening on port %d\n", udpPort)
	svc.cache.data = make(map[string]entry)
	svc.provider = &provider
	if err != nil {
		panic(err)
	}
	transport := &http2.Transport{}
	svc.httpClient = &http.Client{
		Transport: transport,
	}
	return
}

func (svc *DNSService) Query(host string, requestType dnsmessage.Type) (answerResources []dnsmessage.Resource, dnsStatusCode dnsmessage.RCode) {
	var err error
	dnsStatusCode = dnsmessage.RCodeSuccess
	cacheKey := host + requestType.String()
	ok := false
	answers, ok := svc.cache.get(cacheKey)
	if !ok {
		log.Printf("no cached record for %s, fetching...\n", host)
		answers, dnsStatusCode, err = svc.provider.Query(host, requestType)
		log.Printf("fetched result from forwarder: status[%d](%+v)", dnsStatusCode, answers)
		if err == nil {
			svc.cache.set(cacheKey, *answers)
		}
	}
	if dnsStatusCode == dnsmessage.RCodeSuccess {
		for _, answer := range *answers {
			resource, err := answer.ToResource()
			if err == nil {
				answerResources = append(answerResources, resource)
			} else {
				log.Printf("error: %+v\n", err)
			}
		}

	}

	return answerResources, dnsStatusCode
}

func (svc *DNSService) Listen() {
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

		go func(m dnsmessage.Message) {
			question := m.Questions[0]
			questionName := question.Name.String()
			requestType := question.Type

			answerResources, responseStatusCode := svc.Query(questionName, requestType)
			log.Printf("%s %s results: %s %+v\n", questionName, requestType.String(), responseStatusCode, answerResources)
			m.Header.RCode = responseStatusCode
			m.Header.Response = true

			svc.sendPacket(Packet{
				addr: *addr,
				message: dnsmessage.Message{
					Header:      m.Header,
					Questions:   m.Questions,
					Answers:     answerResources,
					Authorities: m.Authorities,
					Additionals: m.Additionals,
				},
			})
		}(m)
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
