package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/idMiFeng/order_service/config"
	"github.com/idMiFeng/order_service/dao/mq"
	"github.com/idMiFeng/order_service/dao/mysql"
	"github.com/idMiFeng/order_service/handler"
	"github.com/idMiFeng/order_service/logger"
	"github.com/idMiFeng/order_service/proto"
	"github.com/idMiFeng/order_service/registry"
	"github.com/idMiFeng/order_service/rpc"
	"github.com/idMiFeng/order_service/third_party/snowflake"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime" // !!!!
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	// 可以把启动时的一些列操作放到 bootstap/init 包

	var cfn string
	// 0.从命令行获取可能的conf路径
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

	// 5.初始化RPC客户端
	err = rpc.InitSrvClient()
	if err != nil {
		panic(err)
	}

	// 6. 初始化snowflake
	err = snowflake.Init(config.Conf.StartTime, config.Conf.MachineID)
	if err != nil {
		panic(err)
	}

	// 7. 初始化rocketmq
	err = mq.Init()
	if err != nil {
		panic(err)
	}
	// 监听订单超时的消息
	c, _ := rocketmq.NewPushConsumer(
		consumer.WithGroupName("order_srv_1"),
		consumer.WithNsResolver(primitive.NewPassthroughResolver([]string{"127.0.0.1:9876"})),
	)
	// 订阅topic
	err = c.Subscribe("xx_pay_timeout", consumer.MessageSelector{}, handler.OrderTimeouthandle)
	if err != nil {
		fmt.Println(err.Error())
	}
	// Note: start after subscribe
	err = c.Start()
	if err != nil {
		panic(err)
	}
	// 监听端口
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Conf.Port))
	if err != nil {
		panic(err)
	}
	// 创建gRPC服务
	s := grpc.NewServer()
	// 健康检查
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())
	proto.RegisterOrderServer(s, &handler.OrderSrv{})
	// 商品服务注册RPC服务
	// 启动gRPC服务
	go func() {
		err = s.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()

	// 注册服务到consul
	registry.Reg.RegisterService(config.Conf.Name, config.Conf.IP, config.Conf.Port, nil)

	zap.L().Info("service start...")

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext( // RPC客户端
		context.Background(),
		fmt.Sprintf("%s:%d", config.Conf.IP, config.Conf.Port),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	// Register Greeter
	err = proto.RegisterOrderHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    ":8093",
		Handler: gwmux,
	}

	zap.L().Info("Serving gRPC-Gateway on http://0.0.0.0:8093")
	go func() {
		err := gwServer.ListenAndServe()
		if err != nil {
			log.Printf("gwServer.ListenAndServe failed, err: %v", err)
			return
		}
	}()

	// 服务退出时要注销服务
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit // 正常会hang在此处
	// 退出时注销服务
	serviceId := fmt.Sprintf("%s-%s-%d", config.Conf.Name, config.Conf.IP, config.Conf.Port)
	registry.Reg.Deregister(serviceId)
}
