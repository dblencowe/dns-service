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
	// SOA  requestSOA
	// MX   requestMX
	// SRV  requestSRV
}

// type requestSOA struct {
// 	NS      string
// 	MBox    string
// 	Serial  uint32
// 	Refresh uint32
// 	Retry   uint32
// 	Expire  uint32
// 	MinTTL  uint32
// }

// type requestMX struct {
// 	Pref uint16
// 	MX   string
// }

// type requestSRV struct {
// 	Priority uint16
// 	Weight   uint16
// 	Port     uint16
// 	Target   string
// }

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
	// case "SOA":
	// 	fmt.Printf("\nHERE! %+v\n", req)
	// 	resourceType = dnsmessage.TypeSOA
	// 	soa := req.SOA
	// 	soaNS, err := dnsmessage.NewName(soa.NS)
	// 	if err != nil {
	// 		return none, err
	// 	}
	// 	soaMBox, err := dnsmessage.NewName(soa.MBox)
	// 	if err != nil {
	// 		return none, err
	// 	}
	// 	body = &dnsmessage.SOAResource{NS: soaNS, MBox: soaMBox, Serial: soa.Serial, Refresh: soa.Refresh, Retry: soa.Retry, Expire: soa.Expire}
	case "PTR":
		resourceType = dnsmessage.TypePTR
		ptr, err := dnsmessage.NewName(req.Data)
		if err != nil {
			return none, err
		}
		body = &dnsmessage.PTRResource{PTR: ptr}
	// case "MX":
	// 	resourceType = dnsmessage.TypeMX
	// 	mxName, err := dnsmessage.NewName(req.Data)
	// 	if err != nil {
	// 		return none, err
	// 	}
	// 	body = &dnsmessage.MXResource{Pref: req.MX.Pref, MX: mxName}
	case "AAAA":
		resourceType = dnsmessage.TypeAAAA
		ip := net.ParseIP(req.Data)
		if ip == nil {
			return none, errors.New("invalid ip suppled")
		}
		var ipV6 [16]byte
		copy(ipV6[:], ip)
		body = &dnsmessage.AAAAResource{AAAA: ipV6}
	// case "SRV":
	// 	resourceType = dnsmessage.TypeSRV
	// 	srv := req.SRV
	// 	srvTarget, err := dnsmessage.NewName(srv.Target)
	// 	if err != nil {
	// 		return none, err
	// 	}
	// 	body = &dnsmessage.SRVResource{Priority: srv.Priority, Weight: srv.Weight, Port: srv.Port, Target: srvTarget}
	case "TXT":
		fallthrough
	case "OPT":
		fallthrough
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
