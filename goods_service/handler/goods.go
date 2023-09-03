package handler

import (
	"context"
	"fmt"
	"github.com/idMiFeng/goods_service/biz/goods"
	"github.com/idMiFeng/goods_service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// handler -> biz -> dao
type GoodsSrv struct {
	proto.UnimplementedGoodsServer
}

func (GoodsSrv) GetGoodsByRoom(ctx context.Context, req *proto.GetGoodsByRoomReq) (*proto.GoodsListResp, error) {
	fmt.Println(req.RoomId)
	if req.GetRoomId() <= 0 {
		// 无效的请求
		return nil, status.Error(codes.InvalidArgument, "请求参数有误")
	}
	// 去查询数据并封装返回的响应数据 --> 业务逻辑
	data, err := goods.GetGoodsByRoomId(ctx, req.GetRoomId())
	if err != nil {
		return nil, status.Error(codes.Internal, "内部错误")
	}
	return data, nil
}
func (GoodsSrv) GetGoodsDetail(context.Context, *proto.GetGoodsDetailReq) (*proto.GoodsDetail, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGoodsDetail not implemented")
}
