package rpc

import (
	"errors"
	"fmt"
	"github.com/idMiFeng/order_service/config"
	"github.com/idMiFeng/order_service/proto"

	_ "github.com/mbobakov/grpc-consul-resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 正常来说应该在请求的时候访问注册中心得到ip和端口注册rpc客户端调用服务，这里用全局变量来设置客户端，如果出现服务宕机，就不可用了
const ()

// 初始化其他服务的RPC客户端

var (
	GoodsCli proto.GoodsClient // 商品服务
	StockCli proto.StockClient // 库存服务
)

func InitSrvClient() error {
	if len(config.Conf.GoodsService.Name) == 0 {
		return errors.New("invalid GoodsService.Name")
	}
	if len(config.Conf.StockService.Name) == 0 {
		return errors.New("invalid StockService.Name")
	}
	// consul实现服务发现
	// 程序启动的时候请求consul获取一个可以用的商品服务地址
	goodsConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s/%s?wait=14s", config.Conf.ConsulConfig.Addr, config.Conf.GoodsService.Name),
		// 指定round_robin策略
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Printf("dial goods_srv failed, err:%v\n", err)
		return err
	}
	GoodsCli = proto.NewGoodsClient(goodsConn)

	stockConn, err := grpc.Dial(
		// consul服务
		fmt.Sprintf("consul://%s/%s?wait=14s", config.Conf.ConsulConfig.Addr, config.Conf.StockService.Name),
		// 指定round_robin策略
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Printf("dial stock_srv failed, err:%v\n", err)
		return err
	}
	StockCli = proto.NewStockClient(stockConn)
	return nil
}
