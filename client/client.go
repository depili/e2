package client

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

type NotFound struct {
	ID int
}

func (err NotFound) Error() string {
	return fmt.Sprintf("Not Found: %v", err.ID)
}

type Options struct {
	Address  string        `long:"e2-address" value-name:"HOST"`
	JSONPort string        `long:"e2-jsonrpc-port" value-name:"PORT" default:"9999"`
	XMLPort  string        `long:"e2-xml-port" value-name:"PORT" default:"9876"`
	Timeout  time.Duration `long:"e2-timeout" default:"10s"`

	ReadKeepalive bool // return keepalive updates from XMLClient.Read()
}

// Returns something sufficient to identify matching Options for the same System
func (options Options) String() string {
	return options.Address
}

func (options Options) Client() (*Client, error) {
	if options.Address == "" {
		return nil, fmt.Errorf("No Address given")
	}

	client := &Client{
		options: options,
		rpcURL: url.URL{
			Scheme: "http",
			Host:   net.JoinHostPort(options.Address, options.JSONPort),
		},
		httpClient: &http.Client{
			Timeout: options.Timeout,
		},
	}

	return client, nil
}

type Client struct {
	options Options

	// JSON RPC
	rpcURL     url.URL
	httpClient *http.Client
	seq        int
}

func (client *Client) String() string {
	return client.options.Address
}
