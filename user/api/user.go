// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package main

import (
	"flag"
	"fmt"
	"net/http"
	"user/api/internal/config"
	"user/api/internal/handler"
	"user/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "api/etc/user-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// WithUnauthorizedCallback 是服务级选项：go-zero 内置的 JWT 中间件默认直接把 401
	// 裸写在 ResponseWriter 上、不经 httpx.Error，这里接管过来，让它也输出统一信封。
	server := rest.MustNewServer(c.RestConf,
		rest.WithUnauthorizedCallback(handler.UnauthorizedCallback))
	defer server.Stop()

	// 在启动 HTTP 服务器之前注册统一错误处理器
	handler.RegisterErrorHandler()
	server.Use(middleware)

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

// 自定义的中间件
func middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Middleware", "static-middleware")
		next(w, r)
	}
}
