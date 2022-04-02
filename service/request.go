package service

import (
	"errors"
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

type request struct {
	Host string
	TTL  uint32
	Type string
	Data string
}

func toRequest(res dnsmessage.Resource) (request, error) {
	return request{
		Host: res.Header.Name.String(),
		TTL:  res.Header.TTL,
		Type: res.Header.Type.String(),
		Data: res.Body.GoString(),
	}, nil
}

func toResource(req *request) (dnsmessage.Resource, error) {
	name, err := dnsmessage.NewName(req.Host)
	none := dnsmessage.Resource{}
	if err != nil {
		return none, err
	}
	var resourceType dnsmessage.Type
	var body dnsmessage.ResourceBody
	switch req.Type {
	case "A":
		resourceType = dnsmessage.TypeA
		ip := net.ParseIP(req.Data)
		if ip == nil {
			return none, errors.New("invalid IP supplied")
		}
		body = &dnsmessage.AResource{A: [4]byte{ip[12], ip[13], ip[14], ip[15]}}
	default:
		return none, errors.New("unsupported record type")
	}

	return dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  name,
			Type:  resourceType,
			Class: dnsmessage.ClassINET,
			TTL:   req.TTL,
		},
		Body: body,
	}, nil
}
