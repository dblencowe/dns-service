package service

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
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

func DoForwarderRequest(httpClient *http.Client, host string, requestType dnsmessage.Type) (*[]Request, dnsmessage.RCode, error) {
	trimmedRequestType := strings.TrimPrefix(requestType.String(), "Type")
	req, err := http.NewRequest(http.MethodGet, "https://1.1.1.1/dns-query?name="+host+"&type="+trimmedRequestType, nil)
	if err != nil {
		return &[]Request{}, dnsmessage.RCodeServerFailure, err
	}
	req.Header.Set("Accept", "application/dns-json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return &[]Request{}, dnsmessage.RCodeServerFailure, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &[]Request{}, dnsmessage.RCodeServerFailure, err
	}
	var apiResponse forwarderResponse
	json.Unmarshal([]byte(body), &apiResponse)
	log.Printf("cloudflare response: %+v\n", apiResponse)
	if len(apiResponse.Answer) == 0 {
		return &[]Request{{
			Host: apiResponse.Question[0].Name + ".",
			Type: trimmedRequestType,
		}}, dnsmessage.RCodeNameError, errors.New("no results from forwarder")
	}
	var answers []Request
	for _, answer := range apiResponse.Answer {
		answers = append(answers, Request{
			Host: answer.Name + ".",
			Type: trimmedRequestType,
			TTL:  answer.TTL,
			Data: answer.Data,
		})
	}

	log.Printf("built request: %+v\n", answers)
	return &answers, dnsmessage.RCodeSuccess, nil
}
