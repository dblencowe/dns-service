package main

import service "github.com/dblencowe/dns-service/service"

func main() {
	service.DNSServer.Listen(&service.DNSService{})
}
