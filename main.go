package main

import (
	"os"

	"github.com/dblencowe/dns-service/providers"
	service "github.com/dblencowe/dns-service/service"
)

func main() {
	provider := providers.InitCloudflareHttpsDNSProvider()
	filterFile := os.Getenv("FILTER_FILE")
	filterService := service.InitFilterService(filterFile)
	svc := service.InitDNSService(*provider, filterService)
	svc.Listen()
}
