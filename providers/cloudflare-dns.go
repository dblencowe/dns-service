package providers

import (
	"log"
	"net"

	"github.com/dblencowe/dns-service/request"
	"golang.org/x/net/dns/dnsmessage"
)

type CloudflareDNSProvider struct {
	DNSProvider
	conn *net.UDPConn
}

const (
	packetLen int = 512
)

func (provider *CloudflareDNSProvider) Query(hostname string, recordType dnsmessage.Type) (*[]request.Request, dnsmessage.RCode, error) {
	var m dnsmessage.Message
	queryName, err := dnsmessage.NewName(hostname)
	if err != nil {
		return &[]request.Request{}, dnsmessage.RCodeFormatError, err
	}
	provider.sendPacket(&net.UDPAddr{IP: net.IPv4(1, 1, 1, 1), Port: 53}, dnsmessage.Message{
		Questions: []dnsmessage.Question{
			{
				Name:  queryName,
				Type:  recordType,
				Class: dnsmessage.ClassANY,
			},
		},
	})
	for {
		buf := make([]byte, packetLen)
		_, addr, err := provider.conn.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		if addr.IP.String() != "1.1.1.1" {
			log.Println("[DEBUG] skipping non cloudflare packet from %s\n", addr.IP.String())
			continue
		}
		err = m.Unpack(buf)
		if err != nil {
			log.Println(err)
			continue
		}

		if len(m.Questions) == 0 {
			continue
		}

		if m.Header.Response {
			log.Printf("Response: %+v", m)
			break
		}
	}

	var answers []request.Request
	for _, answer := range m.Answers {
		answers = append(answers, request.Request{
			Host: answer.Header.Name.String(),
			Type: answer.Header.Type.String(),
			TTL:  answer.Header.TTL,
			Data: answer.Body.GoString(),
		})
	}
	log.Printf("built request: %+v\n", answers)
	return &answers, dnsmessage.RCodeSuccess, nil
}

func (provider *CloudflareDNSProvider) sendPacket(addr *net.UDPAddr, message dnsmessage.Message) {
	packed, err := message.Pack()
	if err != nil {
		log.Println("error packing", err)
		return
	}
	_, err = provider.conn.WriteToUDP(packed, addr)
	if err != nil {
		log.Println("error sending to socket", err)
	}
}
