package transport

type Proxy struct {
	HTTP *HTTPServer
	// Add gRCP, PubSub consumers, etc
}

func NewProxy(server *HTTPServer) *Proxy {
	return &Proxy{
		HTTP: server,
	}
}
