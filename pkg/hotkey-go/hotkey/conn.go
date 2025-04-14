package hotkey

import (
	"encoding/binary"
	"encoding/json"
	"fansX/pkg/hotkey-go/model"
	"io"
	"log/slog"
	"time"
)

func (c *conn) write(msg []byte) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(len(msg)))

	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, _ = c.conn.Write(append(buf, msg...))

}

func (c *conn) read() ([]byte, error) {
	head := make([]byte, 4)
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

		if msg.Type == model.Ping {
			c.last = time.Now().Unix()
			c.write(model.ClientPongMessage)
		} else if msg.Type == model.Pong {
			c.last = time.Now().Unix()
			continue
		} else if msg.Type == model.AddKey {
			c.core.addProcess(c.core, msg)
		} else if msg.Type == model.DelKey {
			c.core.delProcess(c.core, msg)
		} else {
			slog.Error("unKnow message type")
		}
	}
}
