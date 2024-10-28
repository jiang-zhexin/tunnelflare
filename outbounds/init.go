package outbounds

import (
	"errors"
	"io"
	"net/url"
)

func NewClient(serverURL *url.URL, ech string) (ProxyClient, error) {
	switch serverURL.Scheme {
	case "h2":
		return NewHttp2Client(serverURL, ech)
	case "ws", "wss":
		return NewWsClient(serverURL, ech)
	default:
		return nil, errors.New("Unkown protocol")
	}
}

type ProxyClient interface {
	Relay(target string, r io.ReadCloser) (io.ReadCloser, error)
}
