package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	url      = `http://www.metal-archives.com`
	labels   = "%s/index/latest-artists/%s"
	artists  = "%s/index/latest-labels/%s"
	modified = "by/modified"
)

func get(url string) (*http.Response, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("http request returned status %d\n", res.Status))
	}

	return res, nil
}

func work() {

}

func main() {
	for {
		select {
		case <-time.Tick(2 * time.Second):
			//get()
		}
	}
}
