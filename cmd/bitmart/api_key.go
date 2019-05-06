package bitmart

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// APIKey stores a bitmart API key
type APIKey struct {
	Access string `json:"access"`
	Secret string `json:"secret"`
	Memo   string `json:"memo"`
}

// AccessBytes returns the bytes of the access key
func (ak APIKey) AccessBytes() ([]byte, error) {
	return hex.DecodeString(ak.Access)
}

// SecretBytes returns the bytes of the secret key
func (ak APIKey) SecretBytes() ([]byte, error) {
	return hex.DecodeString(ak.Secret)
}

// HMACSign creates an HMAC signature of the given message using the provided key.
//
// The output string is hex-encoded.
func HMACSign(key, message string) string {
	// key: ascii bytes of lowercase hex of a real byte string
	mac := hmac.New(sha256.New, []byte(strings.ToLower(key)))
	_, err := mac.Write([]byte(message))
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(mac.Sum(nil))
}

// ClientSecret returns the client secret: hex-encoded HMAC of the SHA256 hash of the message
func (ak APIKey) ClientSecret() (s string, err error) {
	message := strings.ToLower(fmt.Sprintf("%s:%s:%s", ak.Access, ak.Secret, ak.Memo))
	return HMACSign(ak.Secret, message), nil
}

// Token is an access token for bitmart
type Token struct {
	Access         string  `json:"access_token"`
	ExpiryDuration float64 `json:"expires_in"`
	expiry         time.Time
}

// ParseToken creates a token from a json message
func ParseToken(message string) (*Token, error) {
	t := new(Token)
	err := json.Unmarshal([]byte(message), &t)
	if err != nil {
		return nil, errors.Wrap(err, "json unmarshal")
	}
	t.expiry = time.Now().Add(time.Duration(t.ExpiryDuration * float64(time.Second)))
	return t, nil
}

// Authenticate with bitmart to get a temporary access token
func (ak APIKey) Authenticate() (*Token, error) {
	secret, err := ak.ClientSecret()
	if err != nil {
		return nil, errors.Wrap(err, "creating client secret")
	}

	message := url.Values{}
	message.Set("grant_type", "client_credentials")
	message.Set("client_id", ak.Access)
	message.Set("client_secret", secret)
	buf := bytes.NewBuffer([]byte(message.Encode()))

	resp, err := http.Post(AuthAPI, "application/x-www-form-urlencoded", buf)
	if err != nil {
		return nil, errors.Wrap(err, "requesting grant")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading response body")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s: %s", resp.Status, body)
	}

	return ParseToken(string(body))
}

// LoadAPIKey loads an APIKey from a .apikey.json file
func LoadAPIKey(path string) (key APIKey, err error) {
	var f *os.File
	f, err = os.Open(path)
	if err != nil {
		err = errors.Wrap(err, "opening")
		return
	}
	defer f.Close()
	var data []byte
	data, err = ioutil.ReadAll(f)
	if err != nil {
		err = errors.Wrap(err, "reading")
		return
	}
	err = json.Unmarshal(data, &key)
	if err != nil {
		err = errors.Wrap(err, "parsing")
		return
	}

	return
}
