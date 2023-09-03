package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/idMiFeng/stock_service/biz/stock"
	"github.com/idMiFeng/stock_service/dao/mysql"
	"github.com/idMiFeng/stock_service/model"
	"github.com/idMiFeng/stock_service/proto"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// RPC的入口

type StockSrv struct {
	proto.UnimplementedStockServer
}

// SetStock 设置库存
func (s *StockSrv) SetStock(context.Context, *proto.GoodsStockInfo) (*emptypb.Empty, error) {

	return nil, nil
}

// GetStock 获取商品库存信息
func (s *StockSrv) GetStock(ctx context.Context, req *proto.GoodsStockInfo) (*proto.GoodsStockInfo, error) {
	// 参数处理
	if req.GetGoodsId() <= 0 {
		// 无效的请求
		return nil, status.Error(codes.InvalidArgument, "请求参数有误")
	}
	// 去查询数据并封装返回的响应数据 --> 业务逻辑
	data, err := stock.GetStockByGoodsId(ctx, req.GetGoodsId())
	if err != nil {
		return nil, status.Error(codes.Internal, "内部错误")
	}
	return data, nil
}

// 批量查询库存
func (s *StockSrv) BatchGetStock(ctx context.Context, req *proto.StockInfoList) (*proto.StockInfoList, error) {
	if len(req.GetData()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "请求参数有误")
	}
	data, err := stock.BatchGetStockByGoodsId(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, "内部错误")
	}
	return data, nil

}

// ReduceStock 扣减库存
func (s *StockSrv) ReduceStock(ctx context.Context, req *proto.GoodsStockInfo) (*emptypb.Empty, error) {
	fmt.Printf("in ReduceStock... req:%#v\n", req)
	// 参数处理
	if req.GetGoodsId() <= 0 {
		// 无效的请求
		return nil, status.Error(codes.InvalidArgument, "请求参数有误")
	}
	// 扣减库存
	err := stock.ReduceStockByGoodsId(ctx, req.GetGoodsId(), req.GetNum(), req.GetOrderId())
	if err != nil {
		return nil, status.Error(codes.Internal, "内部错误")
	}
	return &emptypb.Empty{}, nil
}

// 批量扣库存
func (s *StockSrv) BatchReduceStock(ctx context.Context, req *proto.StockInfoList) (*proto.StockInfoList, error) {
	if len(req.GetData()) <= 0 {
		// 无效的请求
		return nil, status.Error(codes.InvalidArgument, "请求参数有误")
	}
	data, err := stock.BatchReduceStock(ctx, req.Data)
	if err != nil {
		return nil, status.Error(codes.Internal, "内部错误")
	}
	return data, nil
}

// RollbackStock 批量归还库存
// func (s *StockSrv) RollbackStock(ctx context.Context, req *proto.StockInfoList) (*emptypb.Empty, error) {
// 	// 参数校验
// 	if len(req.GetData()) < 1 {
// 		// 无效的请求
// 		return nil, status.Error(codes.InvalidArgument, "请求参数有误")
// 	}
// 	// 业务处理
// 	err := stock.RollbackStock(ctx, req.GetData())
// 	if err != nil {
// 		return nil, status.Error(codes.Internal, "回滚库存失败")
// 	}
// 	return &emptypb.Empty{}, nil
// }

// RollbackMsghandle 监听rocketmq消息进行库存回滚的处理函数
// 需考虑重复归还的问题（幂等性）
// 添加库存扣减记录表
func RollbackMsghandle(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for i := range msgs {
		var data model.OrderGoodsStockInfo
		err := json.Unmarshal(msgs[i].Body, &data)
		if err != nil {
			zap.L().Error("json.Unmarshal RollbackMsg failed", zap.Error(err))
			continue
		}
		// 将库存回滚
		err = mysql.RollbackStockByMsg(ctx, data)
		if err != nil {
			return consumer.ConsumeRetryLater, nil
		}
		return consumer.ConsumeSuccess, nil
	}
	return consumer.ConsumeSuccess, nil
}
