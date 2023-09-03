package goods

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/idMiFeng/goods_service/dao/mysql"
	"github.com/idMiFeng/goods_service/proto"
)

func GetGoodsByRoomId(ctx context.Context, roomId int64) (*proto.GoodsListResp, error) {
	// 先查es
	// 再查MySQL
	// 更新redis

	// 1. 先去 xx_room_goods 表 根据 room_id 查询出所有的 goods_id
	objList, err := mysql.GetGoodsByRoomId(ctx, roomId)
	if err != nil {
		return nil, err
	}
	// 处理数据
	// 1. 拿出所有的商品id
	// 2. 记住当前正在讲解的商品id
	var (
		currGoodsId int64
		idList      = make([]int64, 0, len(objList))
	)

	for _, obj := range objList {
		fmt.Printf("obj:%#v\n", obj)
		idList = append(idList, obj.GoodsId)
		if obj.IsCurrent == 1 {
			currGoodsId = obj.GoodsId
		}
	}
	// 2. 再拿上面获取到的 goods_id 去 xx_goods 表查询所有的商品详细信息
	goodsList, err := mysql.GetGoodsById(ctx, idList)
	if err != nil {
		return nil, err
	}
	// 拼装响应数据
	data := make([]*proto.GoodsInfo, 0, len(goodsList))
	for _, goods := range goodsList {
		var headImgs []string
		json.Unmarshal([]byte(goods.HeadImgs), &headImgs)
		data = append(data, &proto.GoodsInfo{
			GoodsId:     goods.GoodsId,
			CategoryId:  goods.CategoryId,
			Status:      int32(goods.Status),
			Title:       goods.Title,
			MarketPrice: fmt.Sprintf("%.2f", float64(goods.MarketPrice/100)),
			Price:       fmt.Sprintf("%.2f", float64(goods.Price/100)),
			Brief:       goods.Brief,
			HeadImgs:    headImgs,
		})
	}
	resp := &proto.GoodsListResp{
		CurrentGoodsId: currGoodsId,
		Data:           data,
	}
	return resp, nil
}
