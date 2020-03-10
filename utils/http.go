package utils

import (
	"github.com/prometheus/common/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func HTTPGet(link string, paraMap map[string]string) {
	params := url.Values{}
	Url, err := url.Parse(link)
	if err != nil {
		return
	}
	for k, v := range paraMap {
		params.Set(k, v)
	}

	Url.RawQuery = params.Encode()
	urlPath := Url.String()

	client := http.Client{
		Timeout: time.Second * 1,
	}
	resp, err := client.Get(urlPath)

	if err != nil {
		log.Error(err)
		return
	} else {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		log.Infof("HTTPGet: %s, response: %s", urlPath, string(body))
	}
}
