package inbounds

import (
	"errors"
	"net/url"

	"tunnelflare/outbounds"
)

func NewServer(clientURL *url.URL, server outbounds.ProxyClient) (ProxyClient, error) {
	switch clientURL.Scheme {
	case "proxycommand":
		return NewProxyCommand(server, clientURL)
	case "http":
		return NewConnect(server, clientURL)
	default:
		return nil, errors.New("Unkown protocol")
	}
}

type ProxyClient interface {
	Start() error
}
