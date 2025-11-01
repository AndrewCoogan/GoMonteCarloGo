package api

import (
	"net/http"
	"net/url"
)

type Connection interface {
	Request(endpoint *url.URL) (*http.Response, error)
}

type ClientHost struct {
	client *http.Client
	host   string
}

type Client struct {
	conn   Connection
	apiKey string
}

func ConnectionFactory(host string) Connection {
	client := &http.Client{
		Timeout: requestTimeout,
	}
	return &ClientHost{
		client: client,
		host:   host,
	}
}

func (conn *ClientHost) Request(endpoint *url.URL) (*http.Response, error) {
	endpoint.Scheme = schemeHttps
	endpoint.Host = conn.host
	targetUrl := endpoint.String()
	return conn.client.Get(targetUrl)
}

func NewClient(host string, apiKey string) *Client {
	client := &http.Client{
		Timeout: requestTimeout,
	}
	
	clientHost := &ClientHost{
		client: client,
		host:   host,
	}

	return &Client{
		conn:   clientHost,
		apiKey: apiKey,
	}
}