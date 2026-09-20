package main

import (
	"context"
	"flag"
	"fmt"
	"user/rpc/internal/mqs"

	"github.com/zeromicro/go-zero/core/service"

	"user/rpc/internal/config"
	"user/rpc/internal/server"
	"user/rpc/internal/svc"
	"user/rpc/user"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var configFile = flag.String("f", "rpc/etc/user.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		user.RegisterUserServer(grpcServer, server.NewUserServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	// 启动消费者服务组
	serviceGroup := service.NewServiceGroup()
	defer serviceGroup.Stop()
	for _, mq := range mqs.Consumers(c, context.Background(), ctx) {
		serviceGroup.Add(mq)
	}

	s.AddUnaryInterceptors(authInterceptor)
	s.AddStreamInterceptors(loggingStreamInterceptor)

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)

	// 在后台启动消费者
	go serviceGroup.Start()

	s.Start()
}

func authInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (any, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if len(md["token"]) == 0 {
		return nil, status.Error(codes.Unauthenticated, "缺少 token")
	}
	return handler(ctx, req)
}
func loggingStreamInterceptor(srv any, ss grpc.ServerStream,
	info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	logx.Infof("流式调用开始: %s", info.FullMethod)
	err := handler(srv, ss)
	if err != nil {
		logx.Errorf("流式调用错误: %s — %v", info.FullMethod, err)
	}
	return err
}
