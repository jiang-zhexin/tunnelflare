package inbounds

import (
	"io"
	"net"
	"net/url"
	"os"

	"tunnelflare/outbounds"
)

type ProxyCommand struct {
	server outbounds.ProxyClient
	target string
}

func NewProxyCommand(client outbounds.ProxyClient, u *url.URL) (*ProxyCommand, error) {
	query := u.Query()
	host := query.Get("host")
	port := query.Get("port")

	return &ProxyCommand{
		server: client,
		target: net.JoinHostPort(host, port),
	}, nil
}

func (pc *ProxyCommand) Start() error {
	r, err := pc.server.Relay(pc.target, os.Stdin)
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = io.Copy(os.Stdout, r)
	return err
}
