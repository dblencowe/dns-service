package main

import service "github.com/dblencowe/dns-service/service"

func main() {
	svc := service.InitDNSService()
	svc.Listen()
}
