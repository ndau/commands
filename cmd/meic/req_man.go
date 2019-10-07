package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// RequestManager simplifies making HTTP requests to the ndau api
type RequestManager struct {
	client  *http.Client
	baseurl string
}

// NewRequestManager creates a new RequestManager with sensible defaults
func NewRequestManager(base string) *RequestManager {
	r := RequestManager{
		client: &http.Client{
			Timeout: time.Second * 10,
		},
		baseurl: base,
	}
	return &r
}

// Get fetches the path from the baseurl and JSON-decodes it into interface,
// which must be a pointer.
func (r *RequestManager) Get(path string, result interface{}) error {
	resp, err := r.client.Get(r.baseurl + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Print(string(body))
		return fmt.Errorf("Got bad status '%s' from %s", resp.Status, path)
	}
	if result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			return err
		}
	}
	return nil
}

// Post JSON-encodes the payload and sends it to the baseurl+path,
// then JSON-decodes the response into interface,
// which must be a pointer.
func (r *RequestManager) Post(path string, payload interface{}, result interface{}) error {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(payload)
	resp, err := r.client.Post(r.baseurl+path, "application/json", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%s: %s (%s)", path, resp.Status, string(body))
	}
	if result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			return err
		}
	}
	return nil
}
