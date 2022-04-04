package providers

import (
	"github.com/dblencowe/dns-service/request"
	"golang.org/x/net/dns/dnsmessage"
)

type Provider interface {
	Init()
	Query(hostname string, recordType dnsmessage.Type) (*[]request.Request, error)
}

type DNSProvider struct {
	Name string
}
