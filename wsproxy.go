package main

import (
	"log"
	"net/http"
	"github.com/akamensky/argparse"
	"os"
	"fmt"
)



func main() {
	parser := argparse.NewParser("wsproxy", "Websocket to TCP/UDP socket proxy")

	listenaddr := parser.String("l", "listen", &argparse.Options{
		Required: true,
		Help: "<Address>:<port> to accept Websocket connections from",
	})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	http.HandleFunc("/ws", HandleWebsocket)
	log.Printf("starting proxy on %v", *listenaddr)
	http.ListenAndServe(*listenaddr, nil)
}