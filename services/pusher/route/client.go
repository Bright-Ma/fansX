package route

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	cshash "puhser/consistenthash"
	"strconv"
	"time"
)

// Send 单点消息发送
func (c *Client) Send(message *Message) {
	//设置写超时时间，防止阻塞时间过长
	if err := c.conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(Ctx.Config.Websocket.WriteTimeout))); err != nil {
		return
	}
	msg, err := json.Marshal(*message)
	if err != nil {
		return
	}
	if err = c.conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
		return
	}
	return
}

// SendGlobalMessage 发送全局消息，每次只发送一个桶中的消息，内部调用Send
func SendGlobalMessage(bucketId int64, message *Message) {
	Bucket[bucketId].Range(func(k, v interface{}) bool {
		c := v.(*Client)
		c.Send(message)
		return true
	})
}

func GetUserId(session string) (int64, error) {
	return rand.Int63(), nil
}

func NewClient(conn *websocket.Conn, session string) {
	rdb := Ctx.RDB
	//此处应替换为登录服务
	id, err := GetUserId(session)
	if err != nil {
		fmt.Println(err.Error())
	}

	c := &Client{
		Session:  session,
		userId:   strconv.FormatInt(id, 10),
		bucketId: id % 100,
		conn:     conn,
	}
	//在redis中存储映射关系(random)，并将连接存储到bucket中
	if Ctx.Config.Model == 1 {
		rdb.Set(context.Background(), "pusher:"+c.userId, Ctx.Config.Etcd.Addr, time.Second*time.Duration(Ctx.Config.Redis.TTL))
	}
	Bucket[c.bucketId].Store(c.userId, c)

	go c.HeartCheck()
}

// HeartCheck 心跳检测，这里以空的text消息作为代替（后续更改）
func (c *Client) HeartCheck() {
	rdb := Ctx.RDB
	for {
		//设置read超时时间，超过则认为心跳超时
		_ = c.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(Ctx.Config.Websocket.ReadTimeout)))
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			fmt.Println(err.Error())
			c.CloseConn()
			return
		}
		//重置redis映射时间(random)
		if Ctx.Config.Model == 1 {
			rdb.Expire(context.Background(), "pusher:"+c.userId, time.Second*time.Duration(Ctx.Config.Websocket.ReadTimeout))
		}

	}
}

// CloseConn 关闭连接
func (c *Client) CloseConn() {
	//删除redis中映射关系(random)
	if Ctx.Config.Model == 1 {
		Ctx.RDB.Del(context.Background(), "pusher:"+c.userId)
	}

	//删除桶中的连接
	Bucket[c.bucketId].Delete(c.userId)
	_ = c.conn.Close()
	return
}

// ReConn 一致性hash模式下，关闭不属于当前节点的连接
func ReConn() {
	for i := 0; i < len(Bucket); i++ {
		Bucket[i].Range(func(k, v interface{}) bool {
			c := v.(*Client)
			ids := []string{c.userId}
			if cshash.Get(ids)[0] != Ctx.Config.IP+":"+Ctx.Config.Websocket.Port {
				c.CloseConn()
			}
			return true
		})
	}
}
