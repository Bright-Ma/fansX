package route

import (
	cshash "OutGetWay/consistenthash"
	svc "OutGetWay/context"
	"OutGetWay/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func Init(ctx *svc.Context) {
	cshash.Init(ctx.Config.VirtualNums)
	service.InitService(ctx)
	engine := gin.Default()
	engine.GET("/GetHost", GetHost)
	if err := engine.Run(":" + ctx.Config.Port); err != nil {
		panic(err.Error())
	}
}

func GetHost(c *gin.Context) {
	if service.Model == 1 {
		c.String(http.StatusOK, service.SelectService())
		return
	}

	session, err := c.Cookie("session")
	if err != nil {
		c.String(http.StatusOK, "not found session")
	}
	id, err := getId(session)
	if err != nil {
		c.String(http.StatusOK, "not found id")
	}
	c.String(http.StatusOK, cshash.Get([]string{strconv.FormatInt(id, 10)})[0])
	return
}

func getId(session string) (int64, error) {
	return 1, nil
}
