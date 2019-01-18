package main

import (
	"log"
	"net/http"
	"github.com/akamensky/argparse"
	"os"
	"fmt"
)

var helpDescr = `
wsproxy - Websocket to TCP/UDP socket proxy.

Websocket URL format:
    ws://<addr>:<port>/ws/...
with the following parameters:
    - proto   (tcp, udp)
    - addr
    - port
    - format  (text, bin)

Static fileserver URL format (optional):
    http://<addr>:<port>/www/...
`


func main(){
	parser := argparse.NewParser("wsproxy", helpDescr)

	listenaddr := parser.String("l", "listen", &argparse.Options{
		Required: true,
		Help: "<Address>:<port> to accept Websocket connections from",
	})

	wwwdir := parser.String("w", "www", &argparse.Options {
		Required: false,
		Help: "www directory to serve HTML files from",
	})

	workers := parser.Int("n", "max-workers", &argparse.Options{
		Required: false,
		Help: "number of maximum simultaneous proxy connections",
		Default: 5,
	})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}


	SetMaxWorkers(*workers)

	if wwwdir != nil {
		http.Handle("/www/", http.StripPrefix("/www/", http.FileServer(http.Dir(*wwwdir))))
		log.Printf("starting www server on %v at /www", *listenaddr)
	}

	http.HandleFunc("/ws", HandleWebsocket)
	log.Printf("starting proxy on %v at /ws", *listenaddr)
	http.ListenAndServe(*listenaddr, nil)
}