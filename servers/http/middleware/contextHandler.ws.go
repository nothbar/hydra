package middleware

import (
	x "net/http"

	"github.com/gin-gonic/gin"
	"github.com/micro-plat/hydra/context"
	"github.com/micro-plat/hydra/servers"
)

//WSContextHandler api请求处理程序
func WSContextHandler(exhandler interface{}, name string, engine string, service string, mSetting map[string]string) gin.HandlerFunc {
	handler, ok := exhandler.(servers.IExecuter)
	if !ok {
		panic("不是有效的servers.IExecuter接口")
	}
	ctn, _ := exhandler.(context.IContainer)
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			getLogger(c).Error(err)
			c.AbortWithStatus(x.StatusNotAcceptable)
			return
		}
		srvhandler := &wxHandler{conn: conn, send: make(chan []byte, 256)}
		go srvhandler.readPump(c, conn, handler, ctn, name, engine, service, mSetting)
		srvhandler.writePump()
	}
}