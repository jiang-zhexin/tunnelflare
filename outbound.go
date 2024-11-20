package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/net/http2"
)

type Http2Client struct {
	h2Transport *http2.Transport
	request     *http.Request
}

func NewHttp2Client(h2url *url.URL) (*Http2Client, error) {
	h2Transport := &http2.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
			NextProtos: []string{"h2", "http/1.1"},
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
	return &Http2Client{
		h2Transport: h2Transport,
		request:     request,
	}, nil
}

func (hc *Http2Client) Relay(target string, r io.ReadCloser) (io.ReadCloser, error) {
	request := hc.request.Clone(context.Background())
	q := request.URL.Query()
	q.Set("target", target)
	request.URL.RawQuery = q.Encode()
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
