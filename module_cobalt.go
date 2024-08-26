package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strings"
)

type CobaltResponse struct {
	Status string     `json:"status"` //Will be error / redirect / stream / success / rate-limit / picker.
	Picker []struct { //array of picker items
		Type  string `json:"type"`
		URL   string `json:"url"`
		Thumb string `json:"thumb"`
	} `json:"picker"`
	URL  string   `json:"url"`  //Returns the download link. If the status is picker this field will be empty. Direct link to a file or a link to cobalt's live render.
	Text string   `json:"text"` //Various text, mostly used for errors.
	URLs []string //If the status is picker all the urls will go here.
}

var (
	useragent = fmt.Sprintf("Mozilla/4.0 (%v; %v); retrobalt/v1.0.1 (%v; %v); +(https://github.com/lostdusty/retrobalt)", runtime.GOOS, runtime.GOARCH, runtime.Compiler, runtime.Version())
	client    = http.Client{Transport: disableTls()} //Disables TLS certificate verification & allows the client to be re-used.
)

func disableTls() http.RoundTripper {
	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return tr
}

func doCobaltRequest(opts Settings, api string) (*CobaltResponse, error) {
	payload, _ := json.Marshal(opts)

	cobaltRes, err := genericRequest(api+"/api/json", "POST", strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	jsonbody, err := ioutil.ReadAll(cobaltRes.Body)
	if err != nil {
		return nil, err
	}
	defer cobaltRes.Body.Close()

	var media CobaltResponse
	err = json.Unmarshal(jsonbody, &media)
	if err != nil {
		return nil, err
	}

	if media.Status == "error" || media.Status == "rate-limit" {
		return nil, fmt.Errorf("cobalt error: %v", media.Text)
	}

	log.Println("[info] (module: cobalt) Cobalt responded our request! Starting to download...")

	if media.Status == "picker" {
		for _, p := range media.Picker {
			media.URLs = append(media.URLs, p.URL)
		}
	} else if media.Status == "stream" {
		media.URLs = append(media.URLs, media.URL)
	}

	return &CobaltResponse{
		Status: media.Status,
		URL:    media.URL,
		Text:   "ok", //Cobalt doesn't return any text if it is ok
		URLs:   media.URLs,
	}, nil
}

func genericRequest(url, method string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", useragent)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
