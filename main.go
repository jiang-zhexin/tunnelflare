package main

import (
	"log"
	"net/url"
)

const ocRawURL = "https://114514:1919810@example.com"
const isRawURL = "http://127.0.0.1:8080"

func main() {
	log.SetPrefix("[TunnelFlare]")
	log.SetFlags(0)

	ocURL, _ := url.Parse(ocRawURL)
	oc, _ := NewHttp2Client(ocURL)

	isURL, _ := url.Parse(isRawURL)
	is, _ := NewConnect(oc, isURL)

	log.Println("Now server is running")
	is.Start()
}
