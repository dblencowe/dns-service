package service

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/net/http2"
)

// https://developers.cloudflare.com/1.1.1.1/encryption/dns-over-https/make-api-requests/dns-json/
type forwarderResponse struct {
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

func DoForwarderRequest(host string, requestType dnsmessage.Type) (*request, dnsmessage.RCode, error) {
	transport := &http2.Transport{}
	client := &http.Client{
		Transport: transport,
	}
	trimmedRequestType := strings.TrimPrefix(requestType.String(), "Type")
	req, err := http.NewRequest("GET", "https://1.1.1.1/dns-query?name="+host+"&type="+trimmedRequestType, nil)
	if err != nil {
		return nil, dnsmessage.RCodeServerFailure, err
	}
	req.Header.Set("Accept", "application/dns-json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, dnsmessage.RCodeServerFailure, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, dnsmessage.RCodeServerFailure, err
	}
	var answer forwarderResponse
	json.Unmarshal([]byte(body), &answer)
	log.Printf("cloudflare response: %+v\n", answer)
	if len(answer.Answer) == 0 {
		return &request{
			Host: answer.Question[0].Name + ".",
			Type: trimmedRequestType,
		}, dnsmessage.RCodeNameError, errors.New("no results from forwarder")
	}

	request := &request{
		Host: answer.Answer[0].Name + ".",
		Type: trimmedRequestType,
		TTL:  answer.Answer[0].TTL,
		Data: answer.Answer[0].Data,
	}

	log.Printf("built request: %+v\n", request)
	return request, dnsmessage.RCodeSuccess, nil
}
