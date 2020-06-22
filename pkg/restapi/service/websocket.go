package service

// WebsocketKeepAlive is a simple keepAlive heartbeat to be used
// to keep a long lived websocket connection going
type WebsocketKeepAlive struct {
	Type string `json:"type"`
}

// NewWebsocketKeepAlive returns a default heartbeat websocket keepAlive struct
func NewWebsocketKeepAlive() *WebsocketKeepAlive {
	return &WebsocketKeepAlive{
		Type: "heartbeat",
	}
}
