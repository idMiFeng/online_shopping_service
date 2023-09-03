package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/idMiFeng/order_service/biz/order"
	"github.com/idMiFeng/order_service/config"
	"github.com/idMiFeng/order_service/dao/mq"
	"github.com/idMiFeng/order_service/dao/mysql"
	"github.com/idMiFeng/order_service/model"
	"github.com/idMiFeng/order_service/proto"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrderSrv struct {
	proto.UnimplementedOrderServer
}

// CreateOrder 创建订单
// 生成订单号 查询商品信息（营销中心算价） 扣库存 生成支付信息 调用收货地址 通知商家
// 简化版：生成订单号 查询商品信息 扣库存
// 1. 生成订单号 2.查询商品信息 3.扣库存
func (s *OrderSrv) CreateOrder(ctx context.Context, req *proto.OrderReq) (*emptypb.Empty, error) {
	fmt.Println("in CreateOrder ... ")
	// 参数处理
	if req.GetUserId() <= 0 {
		// 无效的请求
		return nil, status.Error(codes.InvalidArgument, "请求参数有误")
	}
	// 业务处理
	err := order.Create(ctx, req)
	if err != nil {
		zap.L().Error("order.Create failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "内部错误")
	}

	return &emptypb.Empty{}, nil
}

// OrderTimeouthandle 处理 订单超时事件
func OrderTimeouthandle(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for i := range msgs {
		var data model.OrderGoodsStockInfo
		err := json.Unmarshal(msgs[i].Body, &data)
		if err != nil {
			zap.L().Error("json.Unmarshal RollbackMsg failed", zap.Error(err))
			continue
		}
		// 查订单表
		// 1. 如果订单为已支付状态则不处理
		// 2. 如果订单为未支付状态则发送一条回滚库存的消息
		o, err := mysql.QueryOrder(ctx, data.OrderId)
		if err != nil {
			zap.L().Error("mysql.QueryOrder failed", zap.Error(err))
			return consumer.ConsumeRetryLater, nil // 稍后再试
		}
		if o.OrderId == data.OrderId && o.Status == 100 { // 待支付
			msg := &primitive.Message{
				Topic: config.Conf.RocketMqConfig.Topic.StockRollback,
				Body:  msgs[i].Body,
			}
			_, err = mq.Producer.SendSync(context.Background(), msg)
			if err != nil {
				zap.L().Error("send rollback msg failed", zap.Error(err))
				return consumer.ConsumeRetryLater, nil // 稍后再试
			}
			// 发送回滚库存成功，将订单状态设置为关闭
			o.Status = 300
			mysql.UpdateOrder(ctx, o)
		}
	}
	return consumer.ConsumeSuccess, nil
}
