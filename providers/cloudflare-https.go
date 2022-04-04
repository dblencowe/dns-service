package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dblencowe/dns-service/output"
	"github.com/dblencowe/dns-service/request"
	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/net/http2"
)

type CloudflareHttpsDNSProvider struct {
	DNSProvider
	client *http.Client
}

const urlString string = "https://1.1.1.1/dns-query?name=%s&type=%s"

// https://developers.cloudflare.com/1.1.1.1/encryption/dns-over-https/make-api-requests/dns-json/
type cloudflareResponse struct {
	Status   int  `json:"Status"`
	Tc       bool `json:"TC"`
	Rd       bool `json:"RD"`
	Ra       bool `json:"RA"`
	Ad       bool `json:"AD"`
	Cd       bool `json:"CD"`
	Question []struct {
		Name string `json:"name"`
		Type int    `json:"type"`
	} `json:"Question"`
	Answer []struct {
		Name string `json:"name"`
		Type int    `json:"type"`
		TTL  uint32 `json:"TTL"`
		Data string `json:"data"`
	} `json:"Answer"`
}

func InitCloudflareHttpsDNSProvider() *CloudflareHttpsDNSProvider {
	transport := &http2.Transport{}
	provider := CloudflareHttpsDNSProvider{
		DNSProvider: DNSProvider{
			Name: "CloudflareHTTPS",
		},
		client: &http.Client{
			Transport: transport,
		},
	}

	return &provider
}

func (provider *CloudflareHttpsDNSProvider) Query(hostname string, recordType dnsmessage.Type) (*[]request.Request, dnsmessage.RCode, error) {
	if provider.client == nil {
		return nil, dnsmessage.RCodeServerFailure, fmt.Errorf("cloudflare https provider queried without init")
	}
	trimmedRequestType := strings.TrimPrefix(recordType.String(), "Type")
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(urlString, hostname, trimmedRequestType), nil)
	if err != nil {
		return &[]request.Request{}, dnsmessage.RCodeServerFailure, err
	}
	req.Header.Set("Accept", "application/dns-json")
	req.Header.Set("Content-Type", "application/dns-json")
	resp, err := provider.client.Do(req)
	if err != nil {
		return &[]request.Request{}, dnsmessage.RCodeServerFailure, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &[]request.Request{}, dnsmessage.RCodeServerFailure, err
	}
	var formattedResponse cloudflareResponse
	json.Unmarshal([]byte(body), &formattedResponse)
	output.Println(output.Debug, "cloudflare response: %+v\n", formattedResponse)
	if len(formattedResponse.Answer) == 0 {
		return &[]request.Request{{
			Host: formattedResponse.Question[0].Name + ".",
			Type: trimmedRequestType,
		}}, dnsmessage.RCodeNameError, errors.New("no results from forwarder")
	}

	var answers []request.Request
	for _, answer := range formattedResponse.Answer {
		answers = append(answers, request.Request{
			Host: answer.Name + ".",
			Type: trimmedRequestType,
			TTL:  answer.TTL,
			Data: answer.Data,
		})
	}
	output.Println(output.Debug, "built request: %+v\n", answers)
	return &answers, dnsmessage.RCodeSuccess, nil
}
