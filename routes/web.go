package routes

import (
	"eqray-flavor/app/facades"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/path"
)

func Web() {
	facades.Route().Get("/", func(ctx http.Context) http.Response {
		return ctx.Response().Success().String("api is running")
	})

	// 上传文件落在 public 磁盘（storage/app/public），通过 /storage 前缀对外提供访问
	facades.Route().Static("storage", path.Storage("app/public"))
}
