package main

import (
	"flag"
	"log"
	"net/url"

	"tunnelflare/inbounds"
	"tunnelflare/outbounds"
)

func main() {
	log.SetPrefix("[TunnelFlare]")
	log.SetFlags(0)

	ocRawURL := flag.String("url", "", "server URL")
	isRawURL := flag.String("method", "", "inbound method")
	echConfigListBase64 := flag.String("ech", "", "server ech")
	flag.Parse()

	ocURL, err := url.Parse(*ocRawURL)
	if err != nil {
		log.Fatal("Init fail: ", err.Error())
	}
	oc, err := outbounds.NewClient(ocURL, *echConfigListBase64)
	if err != nil {
		log.Fatal("Init fail: ", err.Error())
	}

	isURL, err := url.Parse(*isRawURL)
	if err != nil {
		log.Fatal("Init fail: ", err.Error())
	}
	is, err := inbounds.NewServer(isURL, oc)
	if err != nil {
		log.Fatal("Init fail: ", err.Error())
	}

	log.Println("Now server is running")
	err = is.Start()
	if err != nil {
		log.Fatal("Relay fail: ", err.Error())
	}
}
