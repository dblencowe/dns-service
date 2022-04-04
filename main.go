package main

import (
	"github.com/dblencowe/dns-service/providers"
	service "github.com/dblencowe/dns-service/service"
)

func main() {
	provider := providers.InitCloudflareHttpsDNSProvider()
	svc := service.InitDNSService(*provider)
	svc.Listen()
}
