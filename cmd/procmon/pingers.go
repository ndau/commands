package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/go-redis/redis"
)

// HTTPPinger returns a function compatible with the Monitor's Test
// parameter that pings an HTTP address with a timeout.
func HTTPPinger(u string, timeout time.Duration, logger logrus.FieldLogger) func() Eventer {
	client := http.Client{Timeout: time.Duration(timeout)}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		panic(err)
	}
	return func() Eventer {
		logger.WithField("url", u).WithField("pinger", "HTTPPinger").Debug("pinging")
		resp, err := client.Do(req)
		if err != nil {
			logger.WithField("url", u).WithField("pinger", "HTTPPinger").Debug("error from request")
			return NewErrorEvent(Failed, err)
		}
		resp.Body.Close()
		if resp.StatusCode > 299 || resp.StatusCode < 200 {
			logger.WithField("url", u).WithField("pinger", "HTTPPinger").
				WithField("status", resp.StatusCode).Debug("got bad status")
			return NewErrorEvent(Failed, fmt.Errorf("Got status code %d (%s) from %s",
				resp.StatusCode, resp.Status, u))
		}
		return OK
	}
}

// RedisPinger returns a function that sends a Ping to the
// redis service at the given address and expects PONG.
func RedisPinger(address string, logger logrus.FieldLogger) func() Eventer {
	redis := redis.NewClient(&redis.Options{
		Addr: address,
	})

	return func() Eventer {
		logger.WithField("addr", address).WithField("pinger", "RedisPinger").Debug("pinging")
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

// PortAvailable returns an Eventer that tests to see if
// a given port on the local machine is available to be claimed by
// a process.
// The port parameter may be a space-separated list of ports; this
// will check all of them.
func PortAvailable(port string) func() Eventer {
	ports := strings.Fields(port)
	return func() Eventer {
		for _, p := range ports {
			ln, err := net.Listen("tcp", ":"+p)

			if err != nil {
				return NewErrorEvent(Failed, err)
			}

			_ = ln.Close()
		}
		return OK
	}
}

// PortInUse returns an Eventer that tests to see if
// a given port on the local machine is being serviced by a process.
// The port parameter may be a space-separated list of ports; this
// will check all of them.
func PortInUse(port string, timeout time.Duration, logger logrus.FieldLogger) func() Eventer {
	ports := strings.Fields(port)
	return func() Eventer {
		for _, p := range ports {
			conn, err := net.DialTimeout("tcp", net.JoinHostPort("", p), timeout)
			if err != nil {
				logger.WithField("port", p).WithError(err).Debug("port is not being serviced")
				return NewErrorEvent(Failed, err)
			}
			if conn != nil {
				logger.WithField("port", p).Debug("port is being serviced")
				conn.Close()
			}
		}
		return OK
	}
}

// EnsureDir ensures that a given path name exists as a directory
// on the filesystem; if it already exists, nothing happens, but if it
// doesn't exist, it tries to create it with the given permissions.
func EnsureDir(name string, permissions os.FileMode) func() Eventer {
	return func() Eventer {
		err := os.MkdirAll(name, permissions)
		if err != nil {
			return NewErrorEvent(Failed, err)
		}
		return OK
	}
}
