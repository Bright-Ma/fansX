package hotkey

import (
	"fansX/pkg/hotkey-go/model"
	"time"
)

func init() {
	msgStrategies = make(map[string]MsgStrategy)
	MsgRegister(model.Ping, &MsgPingStrategy{})
	MsgRegister(model.Pong, &MsgPongStrategy{})
	MsgRegister(model.AddKey, &MsgAddStrategy{})
}

var (
	msgStrategies map[string]MsgStrategy
)

func GetMsgStrategy(msgType string) MsgStrategy {
	return msgStrategies[msgType]
}

func MsgRegister(msgType string, strategy MsgStrategy) {
	msgStrategies[msgType] = strategy
}

func (as *MsgAddStrategy) Handle(msg *model.ServerMessage, conn *conn) {
	conn.core.notify(msg.Keys[0])
	conn.core.Set(msg.Keys[0], []byte{}, conn.core.ttl)
}

func (ps *MsgPingStrategy) Handle(msg *model.ServerMessage, conn *conn) {
	conn.last = time.Now().Unix()
	conn.write(model.ClientPongMessage)
}

func (ps *MsgPongStrategy) Handle(msg *model.ServerMessage, conn *conn) {
	conn.last = time.Now().Unix()
}
