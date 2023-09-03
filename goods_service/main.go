package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/idMiFeng/goods_service/config"
	"github.com/idMiFeng/goods_service/dao/mysql"
	"github.com/idMiFeng/goods_service/handler"
	"github.com/idMiFeng/goods_service/logger"
	"github.com/idMiFeng/goods_service/proto"
	"github.com/idMiFeng/goods_service/registry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// 可以把启动时的一些列操作放到 bootstap/init 包

func main() {
	// 可以把启动时的一些列操作放到 bootstap/init 包

	var cfn string
	// 0.从命令行获取可能的conf路径
	// goods_service -conf="./conf/config_qa.yaml"
	// goods_service -conf="./conf/config_online.yaml"
	flag.StringVar(&cfn, "conf", "./conf/config.yaml", "指定配置文件路径")
	flag.Parse()
	// 1. 加载配置文件
	err := config.Init(cfn)
	if err != nil {
		panic(err) // 程序启动时加载配置文件失败直接退出
	}
	// 2. 加载日志
	err = logger.Init(config.Conf.LogConfig, config.Conf.Mode)
	if err != nil {
		panic(err) // 程序启动时初始化日志模块失败直接退出
	}
	// 3. 初始化MySQL
	err = mysql.Init(config.Conf.MySQLConfig)
	if err != nil {
		panic(err) // 程序启动时初始化MySQL失败直接退出
	}
	// 4. 初始化Consul
	err = registry.Init(config.Conf.ConsulConfig.Addr)
	if err != nil {
		panic(err) // 程序启动时初始化注册中心失败直接退出
	}
	// 监听端口
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Conf.RpcPort))
	if err != nil {
		panic(err)
	}
	// 创建gRPC服务
	s := grpc.NewServer()
	// 注册健康检查服务，支持consul来对我进行健康检查
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())
	// 商品服务注册RPC服务，指定handler处理方法
	proto.RegisterGoodsServer(s, &handler.GoodsSrv{})
	// 启动gRPC服务, 使用 s.Serve(lis) 启动 gRPC 服务会进入一个阻塞状态，即直到服务停止才会返回。如果不在一个单独的 Go 协程中运行此代码，整个程序会在此处被阻塞，无法执行后续的代码。
	go func() {
		err = s.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()
	// 注册服务到consul
	registry.Reg.RegisterService(config.Conf.Name, config.Conf.IP, config.Conf.RpcPort, nil)

	zap.L().Info(
		"rpc server start",
		zap.String("ip", config.Conf.IP),
		zap.Int("port", config.Conf.RpcPort),
	)
	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	/*
				gRPC-Gateway原理（有的项目只支持JSON格式API的访问，不支持RPC，这个把http请求转换成rpc请求）
			我们先定义好了一套gRPC服务
		1. Gateway生成一个反向代理
			1. 接收外部的HTTP请求  --》启动一个HTTP服务
		2. 把HTTP请求转为RPC请求 --》1. 要有一个RPC client端  2.要有HTTP请求与RPC方法的对应关系
			 1. 要有RPC客户端
			 2. HTTP请求怎么转为RPC方法，对应关系是proto文件中描述的对应关系
	*/
	conn, err := grpc.DialContext( // RPC客户端
		context.Background(),
		fmt.Sprintf("%s:%d", config.Conf.IP, config.Conf.RpcPort),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	//这里创建了一个新的 runtime.ServeMux 实例，它是 gRPC-Gateway 库提供的 HTTP 多路复用器。它用于将 HTTP 请求路由到相应的 gRPC 服务处理函数。
	gwmux := runtime.NewServeMux()
	//这里使用 proto.RegisterGoodsHandler 将 gRPC 服务 Goods 注册到 gwmux 上。conn 是 gRPC 客户端与 gRPC 服务之间的连接，它用于将 HTTP 请求转发到相应的 gRPC 服务方法。
	err = proto.RegisterGoodsHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}
	//这里创建了一个 HTTP 服务器实例 gwServer，用于监听来自 gRPC-Gateway 的 HTTP 请求

	gwServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Conf.HttpPort),
		Handler: gwmux,
	}

	zap.L().Info(
		"gRPC-Gateway HTTP Server start",
		zap.String("ip", config.Conf.IP),
		zap.Int("port", config.Conf.HttpPort),
	)
	// sugar
	zap.S().Infof("Serving gRPC-Gateway on http://0.0.0.0:%d", config.Conf.HttpPort)
	//启动http服务
	go func() {
		err := gwServer.ListenAndServe()
		if err != nil {
			log.Printf("gwServer.ListenAndServe failed, err: %v", err)
			return
		}
	}()

	// 服务退出时要注销服务, 这里创建了一个通道（channel）叫做 quit，用于接收操作系统信号。
	quit := make(chan os.Signal)
	//使用 signal.Notify 函数来告诉操作系统，当接收到 SIGTERM 或者 SIGINT 信号时，将这些信号发送到 quit 通道。SIGTERM 通常用于请求进程终止，SIGINT 通常是在用户按下 Ctrl+C 时发送给进程的信号。
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit // 正常会hang在此处 程序会一直等待，直到操作系统发送退出信号。
	// 退出时注销服务
	serviceId := fmt.Sprintf("%s-%s-%d", config.Conf.Name, config.Conf.IP, config.Conf.RpcPort)
	registry.Reg.Deregister(serviceId)
}
