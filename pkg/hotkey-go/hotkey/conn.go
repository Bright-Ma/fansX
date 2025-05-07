package hotkey

import (
	"encoding/binary"
	"encoding/json"
	"fansX/pkg/hotkey-go/model"
	"io"
	"log/slog"
	"time"
)

// 发送消息，tcp报文头为4字节的长度字段，使用大端序发送(长度字段)
func (c *conn) write(msg []byte) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(len(msg)))

	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, _ = c.conn.Write(append(buf, msg...))
}

// 读消息，每次读一个包
func (c *conn) read() ([]byte, error) {
	head := make([]byte, 4)
	//获取长度头
	_, err := io.ReadFull(c.conn, head)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(head)
	body := make([]byte, length)

	_, err = io.ReadFull(c.conn, body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *conn) process() {
	// Ping
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for !c.closed.Load() {
			select {
			case <-ticker.C:
				c.write(model.ClientPingMessage)
			}
		}
		return
	}()

	for !c.closed.Load() {
		body, err := c.read()
		if err != nil {
			continue
		}

		msg := &model.ServerMessage{}
		err = json.Unmarshal(body, msg)
		if err != nil {
			continue
		}

		s := GetMsgStrategy(msg.Type)
		if s == nil {
			slog.Error("unknow strategy:", msg.Type)
			continue
		}
		s.Handle(msg, c)
	}
}
