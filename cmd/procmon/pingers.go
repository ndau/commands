package main

import (
	"fmt"
	"net/http"
	"time"
)

// HTTPPinger returns a function compatible with the Monitor's Test
// parameter that pings an HTTP address with a timeout.
func HTTPPinger(u string, timeout time.Duration) func() Eventer {
	client := http.Client{Timeout: time.Duration(timeout)}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		panic(err)
	}
	return func() Eventer {
		resp, err := client.Do(req)
		if err != nil {
			return NewErrorEvent(Failed, err)
		}
		resp.Body.Close()
		if resp.StatusCode > 299 || resp.StatusCode < 200 {
			return NewErrorEvent(Failed, fmt.Errorf("Got status code %d (%s) from %s",
				resp.StatusCode, resp.Status, u))
		}
		return OK
	}
}
