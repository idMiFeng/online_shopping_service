package order

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/idMiFeng/order_service/config"
	"github.com/idMiFeng/order_service/dao/mq"
	"github.com/idMiFeng/order_service/dao/mysql"
	"github.com/idMiFeng/order_service/model"
	"github.com/idMiFeng/order_service/proto"
	"github.com/idMiFeng/order_service/rpc"
	"github.com/idMiFeng/order_service/third_party/snowflake"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// biz层业务代码
// biz -> dao

// OrderEntity 自定义结构体，实现了两个方法
// 发送事务消息的时候 RocketMQ 会自动根据情况调用那两个方法
type OrderEntity struct {
	OrderId int64           //订单号
	Param   *proto.OrderReq //订单详细
	err     error           //报错时返回的错误
}

// 当发送prepare(half) message 成功后, 这个方法（本地的事务方法）就会被执行
func (o *OrderEntity) ExecuteLocalTransaction(*primitive.Message) primitive.LocalTransactionState {
	fmt.Println("in ExecuteLocalTransaction...")
	if o.Param == nil {
		zap.L().Error("ExecuteLocalTransaction param is nil")
		o.err = status.Error(codes.Internal, "invalid OrderEntity")
		return primitive.CommitMessageState
	}
	param := o.Param
	ctx := context.Background()
	// 1. 查询商品金额（营销）--> RPC连接 goods_service
	goodsDatail, err := rpc.GoodsCli.GetGoodsDetail(ctx, &proto.GetGoodsDetailReq{
		GoodsId: param.GoodsId,
		UserId:  param.UserId,
	})
	if err != nil {
		zap.L().Error("GoodsCli.GetGoodsDetail failed", zap.Error(err))
		// 库存未扣减
		o.err = status.Error(codes.Internal, err.Error())
		return primitive.RollbackMessageState
	}
	payAmountStr := goodsDatail.Price
	payAmount, _ := strconv.ParseInt(payAmountStr, 10, 64)

	// 2. 库存校验及扣减  --> RPC连接 stock_service
	_, err = rpc.StockCli.ReduceStock(ctx, &proto.GoodsStockInfo{
		OrderId: o.OrderId,
		GoodsId: o.Param.GoodsId,
		Num:     o.Param.Num,
	})
	if err != nil {
		// 库存扣减失败，丢弃half-message
		zap.L().Error("StockCli.ReduceStock failed", zap.Error(err))
		o.err = status.Error(codes.Internal, "ReduceStock failed")
		return primitive.RollbackMessageState
	}
	// 代码能执行到这里说明 扣减库存成功了，
	// 从这里开始如果本地事务执行失败就需要回滚库存
	// 3. 创建订单
	// 生成订单表
	orderData := model.Order{
		OrderId:        o.OrderId,
		UserId:         param.UserId,
		PayAmount:      payAmount,
		ReceiveAddress: param.Address,
		ReceiveName:    param.Name,
		ReceivePhone:   param.Phone,
		Status:         100, // 待支付
	}
	// mysql.CreateOrder(ctx, &orderData)
	orderDetail := model.OrderDetail{
		OrderId: o.OrderId,
		UserId:  param.UserId,
		GoodsId: param.GoodsId,
		Num:     param.Num,
	}
	// mysql.CreateOrderDetail(ctx, &orderDetail)
	// 在本地事务创建订单和订单详情记录
	err = mysql.CreateOrderWithTransation(ctx, &orderData, &orderDetail)
	if err != nil {
		// 本地事务执行失败了，上一步已经库存扣减成功
		// 就需要将库存回滚的消息投递出去，下游根据消息进行库存回滚
		zap.L().Error("CreateOrderWithTransation failed", zap.Error(err))
		return primitive.CommitMessageState // 将之前发送的hal-message commit
	}
	// 发送延迟消息
	// 1s 5s 10s 30s 1m 2m 3m 4m 5m 6m 7m 8m 9m 10m 20m 30m 1h 2h
	data := model.OrderGoodsStockInfo{
		OrderId: o.OrderId,
		GoodsId: param.GoodsId,
		Num:     param.Num,
	}
	b, _ := json.Marshal(data)
	msg := primitive.NewMessage("xx_order_timeout", b)
	msg.WithDelayTimeLevel(3)
	_, err = mq.Producer.SendSync(context.Background(), msg)
	if err != nil {
		// 发送延时消息失败
		zap.L().Error("send delay msg failed", zap.Error(err))
		return primitive.CommitMessageState
	}

	// 走到这里说明 本地事务执行成功
	// 需要将之前的half-message rollback， 丢弃掉
	return primitive.RollbackMessageState
}

// 当 prepare(half) message 没有响应时(一般网络问题)
// broker 会回查本地事务的状态，此时这个方法会被执行
func (o *OrderEntity) CheckLocalTransaction(*primitive.MessageExt) primitive.LocalTransactionState {
	// 检查本地状态是否创建成功订单
	_, err := mysql.QueryOrder(context.Background(), o.OrderId)
	// 需要再查询订单详情表
	if err == gorm.ErrRecordNotFound {
		// 没查询到说明订单创建失败，需要回滚库存
		return primitive.CommitMessageState
	}
	return primitive.RollbackMessageState
}

func Create(ctx context.Context, param *proto.OrderReq) error {
	// 3.1 生成订单号
	orderId := snowflake.GenID()

	orderEntity := &OrderEntity{
		OrderId: orderId,
		Param:   param,
	}
	//创建生产者
	p, err := rocketmq.NewTransactionProducer(
		orderEntity,
		producer.WithNsResolver(primitive.NewPassthroughResolver([]string{config.Conf.RocketMqConfig.Addr})),
		//producer.WithNsResolver(primitive.NewPassthroughResolver(endPoint)),
		producer.WithRetry(2),
		producer.WithGroupName("order_srv_1"), // 生产者组
	)
	if err != nil {
		zap.L().Error("NewTransactionProducer failed", zap.Error(err))
		return status.Error(codes.Internal, "NewTransactionProducer failed")
	}
	err = p.Start()
	if err != nil {
		zap.L().Error("start producer error", zap.Error(err))
		return status.Error(codes.Internal, "start producer error")
	}
	// 封装消息 orderId GoodsId num
	data := model.OrderGoodsStockInfo{
		OrderId: orderId,
		GoodsId: param.GoodsId,
		Num:     param.Num,
	}
	body, _ := json.Marshal(data)
	msg := &primitive.Message{
		Topic: config.Conf.RocketMqConfig.Topic.StockRollback, // xx_stock_rollback
		Body:  body,
	}
	// 发送事务消息，对应RocketMQ事务消息第一步发送Half消息，发送消息给MQserver返回成功后后会根据之前创建生产者传入的结构体自动调用结构体的执行本地事务方法
	res, err := p.SendMessageInTransaction(context.Background(), msg)
	if err != nil {
		zap.L().Error("SendMessageInTransaction failed", zap.Error(err))
		return status.Error(codes.Internal, "create order failed")
	}
	zap.L().Info("p.SendMessageInTransaction success", zap.Any("res", res))
	// 执行到这一步说明生产者事务已有结果，如果回滚库存的消息被投递出去给消费者（commit）说明本地事务执行失败，也就是创建订单失败
	if res.State == primitive.CommitMessageState {
		return status.Error(codes.Internal, "create order failed")
	}
	// 其他内部错误
	if orderEntity.err != nil {
		return orderEntity.err
	}
	return nil
}
