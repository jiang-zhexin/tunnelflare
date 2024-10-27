package outbounds

import (
	"crypto/tls"
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

const x25519Kyber768Draft00 tls.CurveID = 0x6399 // X25519Kyber768Draft00
