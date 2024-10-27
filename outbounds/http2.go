package outbounds

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/net/http2"
)

type Http2Client struct {
	h2Transport *http2.Transport
	request     *http.Request
}

func NewHttp2Client(h2url *url.URL, ech string) (*Http2Client, error) {
	query := h2url.Query()
	serverIP := query.Get("resolve")
	colo := query.Get("colo")

	h2Transport := &http2.Transport{
		TLSClientConfig: &tls.Config{
			CurvePreferences: []tls.CurveID{x25519Kyber768Draft00},
			MinVersion:       tls.VersionTLS13,
			NextProtos:       []string{"h2", "http/1.1"},
		},
		DialTLSContext: func(ctx context.Context, network, addr string, config *tls.Config) (net.Conn, error) {
			if len(ech) > 0 {
				echConfigListBytes, err := base64.StdEncoding.DecodeString(ech)
				if err != nil {
					return nil, err
				}
				config.EncryptedClientHelloConfigList = echConfigListBytes
			}
			if serverIP != "" {
				_, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				addr = net.JoinHostPort(serverIP, port)
			}
			return tls.Dial(network, addr, config)
		},
		DisableCompression: true,
	}

	server := &url.URL{
		Scheme: "https",
		Host:   h2url.Host,
		Path:   h2url.Path,
	}
	request, err := http.NewRequest("POST", server.String(), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("content-type", "application/grpc")
	if auth := h2url.User.String(); len(auth) > 0 {
		request.Header.Set("Proxy-Authorization", "basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	}
	if len(colo) > 0 {
		request.Header.Set("Target-Colo", colo)
	}
	return &Http2Client{
		h2Transport: h2Transport,
		request:     request,
	}, nil
}

func (hc *Http2Client) Relay(target string, r io.ReadCloser) (io.ReadCloser, error) {
	request := hc.request.Clone(context.Background())
	request.Header.Set("Target", target)
	request.Body = r
	response, err := hc.h2Transport.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Need status code 200, but get %d", response.StatusCode))
	}
	return response.Body, nil
}
