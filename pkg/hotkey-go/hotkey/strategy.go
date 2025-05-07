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
	// 策略工厂
	msgStrategies map[string]MsgStrategy
)

// GetMsgStrategy 获取type对应的处理策略
func GetMsgStrategy(msgType string) MsgStrategy {
	return msgStrategies[msgType]
}

// MsgRegister 注册策略
func MsgRegister(msgType string, strategy MsgStrategy) {
	msgStrategies[msgType] = strategy
}

// Handle 通知观察者，设置hotkey缓存
func (as *MsgAddStrategy) Handle(msg *model.ServerMessage, conn *conn) {
	conn.core.notify(msg.Keys[0])
	conn.core.Set(msg.Keys[0], []byte{}, conn.core.ttl)
}

// Handle 重置消息接收时间戳，发送Pong
func (ps *MsgPingStrategy) Handle(msg *model.ServerMessage, conn *conn) {
	conn.last = time.Now().Unix()
	conn.write(model.ClientPongMessage)
}

func (ps *MsgPongStrategy) Handle(msg *model.ServerMessage, conn *conn) {
	conn.last = time.Now().Unix()
}
