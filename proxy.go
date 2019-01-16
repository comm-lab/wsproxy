package wsproxy

import (
	"net"
	"net/http"
	"github.com/gorilla/websocket"
	"bufio"
	"sync"
	"log"
)



var pool = make(map[string]proxyworker)
var lock = sync.Mutex{}

var MAX_PROXY_WORKERS int = 10


var upgrader = websocket.Upgrader {
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

func HandleWebsocket(w http.ResponseWriter, r* http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		//upgrade failed send 400
		w.WriteHeader(http.StatusBadRequest)
	}

	if len(pool) >= MAX_PROXY_WORKERS {
		ws.WriteJSON(map[string]string{
			"error" : "too many connections",
		})
		ws.Close()
		return
	}
	
	// get fields from URL
	q := r.URL.Query()
	proto, addr, port, msgformat := q.Get("proto"), q.Get("addr"), q.Get("port"), q.Get("format")

	if len(proto) * len(port) * len(addr) == 0 {
		// not enough args
		ws.WriteJSON(map[string]string{
			"error" : "specify URL arguments",
		})
		ws.Close()
		return
	}

	var format int
	switch msgformat {
	case "text": format = websocket.TextMessage
	case "bin": format = websocket.BinaryMessage
	default: format = websocket.TextMessage
	}
	
	// try to establish socket connection
	sock, err := net.Dial(proto, addr + ":" + port)
	if err != nil {
		ws.WriteJSON(map[string]string{
			"error" : "socket connection failed",
		})
		ws.Close()
		return
	}

	proxy := proxyworker{r.RemoteAddr, format, ws, sock}
	log.Printf("new proxy instance: [%v, from: %v, to: %v]. slots available: %v",
		msgformat,
		r.RemoteAddr,
		proxy.key,
		MAX_PROXY_WORKERS - len(pool) - 1,
	)

	lock.Lock()
	pool[proxy.key] = proxy
	pool[proxy.key].start()
	lock.Unlock()
}

type proxyworker struct {
	key string
	format int
	ws *websocket.Conn
	sock net.Conn
}

func (p proxyworker) start() {
	go p.upstream()
	go p.downstream()
}

func (p proxyworker) destroy() {
	p.sock.Close()
	p.ws.Close()

	lock.Lock()
	delete(pool, p.key)
	lock.Unlock()
}

// Socket to Websocket channel
func (p *proxyworker) upstream() {
	reader := bufio.NewReader(p.sock)
	buf := make([]byte, 1024)
	for {
		// Read from Socket
		n, err := reader.Read(buf)
		if err != nil {
			break
		}
		// Write to Websocket
		err = p.ws.WriteMessage(p.format, buf[:n])
		if err != nil {
			break
		}
	}
	p.destroy()
}

// Websocket to socket channel
func (p *proxyworker) downstream() {
	writer := bufio.NewWriter(p.sock)
	for {
		// Read from Websocket
		_, buf, err := p.ws.ReadMessage()
		if err != nil {
			break
		}
		// Write to socket
		n, err := writer.Write(buf)
		if err != nil || n < len(buf) {
			break
		}
		writer.Flush()
	}
	p.destroy()
}