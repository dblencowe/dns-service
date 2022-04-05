package service

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/dblencowe/dns-service/output"
)

type FilterService struct {
	Rewrites map[string]net.IP
}

func InitFilterService(filterListPath string) *FilterService {
	svc := FilterService{}
	svc.Rewrites = make(map[string]net.IP)
	svc.loadFiltersFromFile(filterListPath)

	return &svc
}

func (svc FilterService) Filter(hostname string) (*net.IP, bool) {
	cleanHostname := strings.TrimSuffix(hostname, ".")
	output.Println(output.Debug, "checking filter list for %s", cleanHostname)
	if ip, ok := svc.Rewrites[cleanHostname]; ok {
		return &ip, ok
	}

	return &net.IP{}, false
}

func (svc *FilterService) loadFiltersFromFile(path string) {
	if len(path) == 0 {
		return
	}
	output.Println(output.Debug, "loading filter list from %s", path)
	file, err := os.Open(path)
	if err != nil {
		output.Println(output.Error, "+v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ip, domain, err := extractRecordFromLine(scanner.Text())
		if err != nil {
			output.Println(output.Error, "error extracting line %s: %v", scanner.Text(), err)
			continue
		}
		svc.Rewrites[domain] = ip
		output.Println(output.Debug, "loaded %s for %s from filter file", ip, domain)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	output.Println(output.Info, "loaded %d results from %s for filter list", len(svc.Rewrites), path)
}

const ValidIpAddressRegex string = `(?P<IP>(?:[0-9]{1,3}\.){3}[0-9]{1,3})`
const ValidHostnameRegex string = `(?P<domain>(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9])`

func extractRecordFromLine(line string) (net.IP, string, error) {
	re, err := regexp.Compile(ValidIpAddressRegex + `\s+` + ValidHostnameRegex)
	if err != nil {
		return net.IP{}, "", err
	}
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return net.IP{}, "", fmt.Errorf("unable to match results from line %s: %v", line, matches)
	}
	return net.ParseIP(matches[1]), matches[2], nil
}
