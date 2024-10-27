package inbounds

import (
	"io"
	"log"
	"net/http"
	"net/url"

	"tunnelflare/outbounds"
)

type Connect struct {
	server outbounds.ProxyClient
	client *url.URL
}

func NewConnect(server outbounds.ProxyClient, client *url.URL) (*Connect, error) {
	return &Connect{
		server: server,
		client: client,
	}, nil
}

func (c *Connect) Start() error {
	err := http.ListenAndServe(c.client.Host, c)
	return err
}

func (c *Connect) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodConnect {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
		downConn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			return
		}
		defer downConn.Close()
		r, err := c.server.Relay(req.RequestURI, downConn)
		if err != nil {
			log.Printf("Relay info: connect to %s fail", req.RequestURI)
			log.Println(err.Error())
			return
		}
		defer r.Close()

		log.Printf("Relay info: connect to %s success", req.RequestURI)
		n, err := io.Copy(downConn, r)
		if err != nil {
			return
		}
		log.Printf("Relay info: from %s bytes %d", req.RequestURI, n)
	}
}
