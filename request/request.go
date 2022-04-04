package request

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

type Request struct {
	Host string
	TTL  uint32
	Type string
	Data string
}

func (req *Request) ToResource() (dnsmessage.Resource, error) {
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
		if ip != nil {
			//
			body = &dnsmessage.AResource{A: [4]byte{ip[12], ip[13], ip[14], ip[15]}}
			break
		}

		name, err := dnsmessage.NewName(req.Data)
		if err != nil {
			return none, fmt.Errorf("(%+v) invalid IP / DNS Name supplied", req)
		}
		body = &dnsmessage.CNAMEResource{CNAME: name}
	case "NS":
		resourceType = dnsmessage.TypeNS
		ns, err := dnsmessage.NewName(req.Data)
		if err != nil {
			return none, err
		}
		body = &dnsmessage.NSResource{NS: ns}
	case "CNAME":
		resourceType = dnsmessage.TypeCNAME
		cname, err := dnsmessage.NewName(req.Data)
		if err != nil {
			return none, err
		}
		body = &dnsmessage.CNAMEResource{CNAME: cname}
	case "SOA":
		resourceType = dnsmessage.TypeSOA
		soa := strings.Split(req.Data, " ")
		soaNS, err := dnsmessage.NewName(soa[0])
		if err != nil {
			return none, err
		}
		soaMbox, err := dnsmessage.NewName(soa[1])
		if err != nil {
			return none, err
		}
		serial, err := strconv.ParseUint(soa[2], 10, 32)
		if err != nil {
			return none, err
		}
		refresh, err := strconv.ParseUint(soa[3], 10, 32)
		if err != nil {
			return none, err
		}
		retry, err := strconv.ParseUint(soa[4], 10, 32)
		if err != nil {
			return none, err
		}
		expire, err := strconv.ParseUint(soa[5], 10, 32)
		if err != nil {
			return none, err
		}

		body = &dnsmessage.SOAResource{
			NS:      soaNS,
			MBox:    soaMbox,
			Serial:  uint32(serial),
			Refresh: uint32(refresh),
			Retry:   uint32(retry),
			Expire:  uint32(expire),
		}
	case "PTR":
		resourceType = dnsmessage.TypePTR
		ptr, err := dnsmessage.NewName(req.Data)
		if err != nil {
			return none, err
		}
		body = &dnsmessage.PTRResource{PTR: ptr}
	case "MX":
		resourceType = dnsmessage.TypeMX
		parts := strings.Split(req.Data, " ")
		mxName, err := dnsmessage.NewName(parts[1])
		if err != nil {
			return none, err
		}
		pref, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			return none, err
		}
		body = &dnsmessage.MXResource{MX: mxName, Pref: uint16(pref)}
	case "AAAA":
		resourceType = dnsmessage.TypeAAAA
		ip := net.ParseIP(req.Data)
		if ip != nil {
			var ipV6 [16]byte
			copy(ipV6[:], ip)
			body = &dnsmessage.AAAAResource{AAAA: ipV6}
			break
		}

		name, err := dnsmessage.NewName(req.Data)
		if err != nil {
			return none, fmt.Errorf("(%+v) invalid IP / DNS Name supplied", req)
		}
		body = &dnsmessage.CNAMEResource{CNAME: name}
	case "SRV":
		resourceType = dnsmessage.TypeSRV
		srv := strings.Split(req.Data, " ")
		priority, err := strconv.ParseUint(srv[0], 10, 32)
		if err != nil {
			return none, err
		}
		weight, err := strconv.ParseUint(srv[1], 10, 32)
		if err != nil {
			return none, err
		}
		port, err := strconv.ParseUint(srv[2], 10, 32)
		if err != nil {
			return none, err
		}
		target, err := dnsmessage.NewName(srv[3])
		if err != nil {
			return none, err
		}
		body = &dnsmessage.SRVResource{Priority: uint16(priority), Weight: uint16(weight), Port: uint16(port), Target: target}
	case "TXT":
		fallthrough
	case "OPT":
		fallthrough
	default:
		return none, fmt.Errorf("(%+v) unsupported record type", req)
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
