package main

import (
	"github.com/micro-plat/hydra/conf/server/acl/blacklist"
	"github.com/micro-plat/hydra/conf/server/api"
	ccron "github.com/micro-plat/hydra/conf/server/cron"
	crpc "github.com/micro-plat/hydra/conf/server/rpc"
	"github.com/micro-plat/hydra/conf/server/task"

	"github.com/micro-plat/hydra"
	"github.com/micro-plat/hydra/context"
	"github.com/micro-plat/hydra/hydra/servers/cron"
	"github.com/micro-plat/hydra/hydra/servers/http"
	"github.com/micro-plat/hydra/hydra/servers/rpc"
)

var app = hydra.NewApp(
	hydra.WithServerTypes(http.API, rpc.RPC, cron.CRON),
	hydra.WithPlatName("taosytest"),
	hydra.WithSystemName("test-confcache"),
	hydra.WithClusterName("taosy"),
	hydra.WithRegistry("zk://192.168.0.101"),
)

func init() {
	hydra.Conf.API(":8070", api.WithTimeout(10, 10)).BlackList(blacklist.WithEnable(), blacklist.WithIP("192.168.0.101"))
	hydra.Conf.RPC(":8888", crpc.WithTimeout(10, 10))
	hydra.Conf.CRON(ccron.WithTimeout(10)).Task(task.NewTask("@every 10s", "/taosy/testcron"))
	app.API("/taosy/testapi", func(ctx context.IContext) (r interface{}) {
		ctx.Log().Info("api 接口服务测试")
		return "success"
	})
	app.RPC("/taosy/testrpc", func(ctx context.IContext) (r interface{}) {
		ctx.Log().Info("rpc 接口服务测试")
		return "success"
	})
	app.CRON("/taosy/testcron", func(ctx context.IContext) (r interface{}) {
		ctx.Log().Info("cron 接口服务测试")
		return "success"
	})
}

func main() {
	app.Start()
}