package service

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

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

func DoForwarderRequest(host string) (*request, dnsmessage.RCode, error) {
	transport := &http2.Transport{}
	client := &http.Client{
		Transport: transport,
	}
	req, err := http.NewRequest("GET", "https://1.1.1.1/dns-query?name="+host, nil)
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
	if len(answer.Answer) == 0 {
		return &request{
			Host: answer.Question[0].Name + ".",
			Type: "A",
		}, dnsmessage.RCodeNameError, errors.New("no results from forwarder")
	}
	return &request{
		Host: answer.Answer[0].Name + ".",
		Type: "A",
		TTL:  answer.Answer[0].TTL,
		Data: answer.Answer[0].Data,
	}, dnsmessage.RCodeSuccess, nil
}
