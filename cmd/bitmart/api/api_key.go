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
	Access   string `json:"access"`
	Secret   string `json:"secret"`
	Memo     string `json:"memo"`
	Endpoint string `json:"endpoint,omitempty"`
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
	message := fmt.Sprintf("%s:%s:%s", strings.ToLower(ak.Access), strings.ToLower(ak.Secret), ak.Memo)
	return HMACSign(ak.Secret, message), nil
}

// Token is an access token for bitmart
type Token struct {
	Access         string  `json:"access_token"`
	ExpiryDuration float64 `json:"expires_in"`
	expiry         time.Time
}

// SubsURL might update the provided URL with a replacement scheme and host.
//
// API keys may optionally have endpoint overrides embedded. If so, Subs will
// update the provided URL appropriately. If there is no override defined, or
// if any errors are encountered, Subs does not modify the URL.
//
// Returns the host string appropriate for substituting into a request Host field.
func (ak APIKey) SubsURL(purl *url.URL) (host string) {
	host = purl.Host
	if len(ak.Endpoint) == 0 {
		return
	}
	eurl, err := url.Parse(ak.Endpoint)
	if err != nil {
		return
	}
	purl.Scheme = eurl.Scheme
	purl.Host = eurl.Host
	return eurl.Host
}

// Subs might update the provdied url with a replacement scheme and endpoint.
//
// API keys may optionally have endpoint overrides embedded. If so, Subs will
// update the provided URL appropriately. If there is no override defined, or
// if any errors are encountered parsing either URL, Subs returns the original url.
func (ak APIKey) Subs(urlpath string) string {
	purl, err := url.Parse(urlpath)
	if err != nil {
		return urlpath
	}
	ak.SubsURL(purl)
	return purl.String()
}

// ParseToken creates a token from a json message
func ParseToken(message string) (*Token, error) {
	t := new(Token)
	err := json.Unmarshal([]byte(message), &t)
	if err != nil {
		return nil, errors.Wrap(err, "json unmarshal")
	}
	// tokens expire after the specified number of seconds
	// OR midnight UTC, whichever comes first.
	now := time.Now().UTC()
	t.expiry = now.Add(time.Duration(t.ExpiryDuration * float64(time.Second)))
	if t.expiry.YearDay() > now.YearDay() || (t.expiry.YearDay() == 1 && now.YearDay() >= 365) {
		h, m, s := t.expiry.Clock()
		t.expiry = t.expiry.Add(time.Duration(-h) * time.Hour)
		t.expiry = t.expiry.Add(time.Duration(-m) * time.Minute)
		t.expiry = t.expiry.Add(time.Duration(-s) * time.Second)
	}

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

	url := ak.Subs(APIAuth)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", buf)
	if err != nil {
		return nil, errors.Wrap(err, "requesting grant")
	}
	defer resp.Body.Close()

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
