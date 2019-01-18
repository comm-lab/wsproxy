# wsproxy
Websocket to TCP/UDP socket proxy.

### Basic usage
    
    wsproxy -l :8000

just replace :8000 with the listening address and port you want

If the connection initialization cannot be completed, proxy
will send a JSON message containing _error_
and a brief explanation, closing the WebSocket.

### Websocket URL format:

    ws://<addr>:<port>/ws/...

with the following parameters:

- proto   [tcp, udp]
- addr    
- port
- format  [text, bin]

### Static fileserver URL format (optional):

    http://<addr>:<port>/www/...

