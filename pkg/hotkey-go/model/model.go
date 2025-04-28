package model

import "encoding/json"

func init() {
	clientPing := ClientMessage{
		Type: Ping,
	}
	ClientPingMessage, _ = json.Marshal(clientPing)
	clientPong := ClientMessage{
		Type: Pong,
	}
	ClientPongMessage, _ = json.Marshal(clientPong)
	serverPing := ServerMessage{
		Type: Ping,
	}
	ServerPingMessage, _ = json.Marshal(serverPing)
	serverPong := ServerMessage{
		Type: Pong,
	}
	ServerPongMessage, _ = json.Marshal(serverPong)
}

type ClientMessage struct {
	Type      string         `json:"type"`
	GroupName string         `json:"group_name"`
	Key       map[string]int `json:"key"`
}

type ServerMessage struct {
	Type string   `json:"type"`
	Keys []string `json:"keys"`
}

const (
	Ping   = "ping"
	Pong   = "pong"
	AddKey = "add_key"
)

var ClientPingMessage []byte
var ClientPongMessage []byte
var ServerPingMessage []byte
var ServerPongMessage []byte
