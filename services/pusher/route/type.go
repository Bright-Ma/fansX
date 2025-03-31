package route

import (
	"github.com/gorilla/websocket"
	svc "puhser/internal/context"
	"sync"
)

// Client 客户端，每个连接代表一个客户端
type Client struct {
	Session string
	userId  string
	//指定当前连接放在哪个桶中
	bucketId int64
	conn     *websocket.Conn
}

type Message struct {
	//uuid，用于识别重复消息
	UUId string `json:"uuid"`
	//经过编码的消息体
	PayLoad string `json:"payLoad"`
	//用于指定payload使用什么方式进行编解码
	EncodeType string `json:"encodeType"`
}

var Ctx *svc.Context
var UpGrader websocket.Upgrader

// Bucket 每个bucket中存放一定的连接
var Bucket = make([]sync.Map, 100)
