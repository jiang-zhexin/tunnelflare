package outbounds

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	dialer  websocket.Dialer
	headers http.Header
	server  *url.URL
}

func NewWsClient(wurl *url.URL, ech string) (*WsClient, error) {
	query := wurl.Query()
	serverIP := query.Get("resolve")
	colo := query.Get("colo")

	dialer := *websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
	if len(ech) > 0 {
		echConfigListBytes, err := base64.StdEncoding.DecodeString(ech)
		if err != nil {
			return nil, err
		}
		dialer.TLSClientConfig.EncryptedClientHelloConfigList = echConfigListBytes
	}
	dialer.NetDialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if serverIP != "" {
			_, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			addr = net.JoinHostPort(serverIP, port)
		}
		netDialer := &net.Dialer{}
		return netDialer.DialContext(ctx, network, addr)
	}

	headers := http.Header{}

	if auth := wurl.User.String(); len(auth) > 0 {
		headers.Add("Proxy-Authorization", "basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	}
	if len(colo) > 0 {
		headers.Set("Target-Colo", colo)
	}
	server := &url.URL{
		Scheme: "wss",
		Host:   wurl.Host,
		Path:   wurl.Path,
	}
	return &WsClient{
		dialer:  dialer,
		headers: headers,
		server:  server,
	}, nil
}

func (w *WsClient) Relay(target string, r io.ReadCloser) (io.ReadCloser, error) {
	headers := w.headers.Clone()
	headers.Add("Target", target)
	conn, _, err := w.dialer.Dial(w.server.String(), headers)
	if err != nil {
		return nil, err
	}

	rr, ww := io.Pipe()
	go stdin2Ws(conn, r)
	go ws2Stdout(conn, ww)
	return rr, nil
}

func ws2Stdout(conn *websocket.Conn, w io.WriteCloser) {
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			return
		}
		switch messageType {
		case websocket.BinaryMessage:
			w.Write(message)
		case websocket.TextMessage:
		}
	}
}

func stdin2Ws(conn *websocket.Conn, r io.ReadCloser) {
	// WebSocket messages received by a Worker have a size limit of 1 MiB.
	buffer := make([]byte, 1<<20)
	for {
		n, err := r.Read(buffer)
		if err != nil {
			conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(5*time.Second))
			return
		}
		err = conn.WriteMessage(websocket.BinaryMessage, buffer[:n])
		if err != nil {
			return
		}
	}
}
