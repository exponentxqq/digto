package client

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ysmood/kit"
)

// Client ...
type Client struct {
	// Scheme default is https
	Scheme string

	// Subdomain ...
	Subdomain string

	// APIScheme to use for api request
	APIScheme string
	// APIHost api host
	APIHost string
	// Host api header host
	APIHeaderHost string

	// Concurrent concurrent request when serving
	Concurrent int

	httpClient *http.Client
}

// New new default client
func New(subdomain string) *Client {
	return &Client{
		Scheme:        "https",
		APIScheme:     "https",
		APIHost:       "digto.org",
		APIHeaderHost: "digto.org",
		Subdomain:     subdomain,
		Concurrent:    2,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

// PublicURL the url exposed to public
func (c *Client) PublicURL() string {
	return c.Scheme + "://" + c.Subdomain + "." + c.APIHost
}

// Next serve only once
func (c *Client) Next() (*http.Request, func(status int, header http.Header, body io.Reader) error, error) {
	apiURL := url.URL{
		Scheme: c.APIScheme,
		Host:   c.APIHost,
		Path:   c.Subdomain,
	}

	senderRes, err := resError(kit.Req(apiURL.String()).Client(c.httpClient).Host(c.APIHeaderHost).Response())
	if err != nil {
		return nil, nil, err
	}

	senderURL, _ := url.Parse(senderRes.Header.Get("Digto-URL"))

	receiverReq := &http.Request{
		URL:    senderURL,
		Method: senderRes.Header.Get("Digto-Method"),
		Header: senderRes.Header,
		Body:   senderRes.Body,
	}

	send := func(status int, header http.Header, body io.Reader) error {
		headerToSend := []string{
			"Digto-ID", senderRes.Header.Get("Digto-ID"),
			"Digto-Status", fmt.Sprint(status),
		}
		if header != nil {
			for k, l := range header {
				for _, v := range l {
					headerToSend = append(headerToSend, k, v)
				}
			}
		}

		_, err = resError(
			kit.Req(apiURL.String()).Post().
				Client(c.httpClient).
				Host(c.APIHeaderHost).
				Header(headerToSend...).Body(body).Response(),
		)
		return err
	}

	return receiverReq, send, nil
}

func resError(res *http.Response, err error) (*http.Response, error) {
	errMsg := res.Header.Get("Digto-Error")
	if errMsg != "" {
		return res, errors.New(errMsg)
	}

	return res, err
}
