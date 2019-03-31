package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis"
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

// RedisPinger returns a function that sends a Ping to the
// redis service at the given address and expects PONG.
func RedisPinger(address string) func() Eventer {
	redis := redis.NewClient(&redis.Options{
		Addr: address,
	})

	return func() Eventer {
		result, err := redis.Ping().Result()
		if err != nil {
			return NewErrorEvent(Failed, err)
		}
		if result != "PONG" {
			return NewErrorEvent(Failed, fmt.Errorf("ping expected 'PONG', got '%s'", result))
		}
		return OK
	}
}
