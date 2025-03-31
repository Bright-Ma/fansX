package route

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	svc "puhser/internal/context"
)

func Init(ctx *svc.Context) {
	Ctx = ctx
	UpGrader = websocket.Upgrader{
		ReadBufferSize:  ctx.Config.Websocket.ReadBufferSize,
		WriteBufferSize: ctx.Config.Websocket.WriteBufferSize,
	}

	//gin.SetMode(gin.ReleaseMode)
	//gin.DefaultWriter = ioutil.Discard
	engine := gin.Default()
	engine.GET("/ToWebSocket", ToWebSocket)
	err := engine.Run(":" + ctx.Config.Websocket.Port)
	if err != nil {
		panic(err)
	}
}

func ToWebSocket(ctx *gin.Context) {
	if !websocket.IsWebSocketUpgrade(ctx.Request) {
		return
	}

	session, err := ctx.Cookie("session_id")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	conn, err := UpGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	NewClient(conn, session)
}
