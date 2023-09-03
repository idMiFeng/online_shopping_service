package stock

import (
	"context"
	"github.com/idMiFeng/stock_service/dao/mysql"
	"github.com/idMiFeng/stock_service/proto"
)

// biz层业务代码
// biz -> dao

func GetStockByGoodsId(ctx context.Context, goodsId int64) (*proto.GoodsStockInfo, error) {
	// 先查MySQL数据库
	data, err := mysql.GetStockByGoodsId(ctx, goodsId)
	if err != nil {
		return nil, err
	}
	// 拼装数据
	return &proto.GoodsStockInfo{GoodsId: goodsId, Num: data.Num}, nil
}

// 批量查询库存
func BatchGetStockByGoodsId(ctx context.Context, req *proto.StockInfoList) (*proto.StockInfoList, error) {
	goodsIds := make([]int64, 0)
	for _, stockInfo := range req.Data {
		// 获取每个 *GoodsStockInfo 中的 GoodsId 字段并添加到切片中
		goodsId := stockInfo.GoodsId
		goodsIds = append(goodsIds, goodsId)
	}
	stockList, err := mysql.BatchGetStockByGoodsId(ctx, goodsIds)
	if err != nil {
		return nil, err
	}
	data := make([]*proto.GoodsStockInfo, 0)
	for _, stock := range *stockList {
		GoodsStockInfo := proto.GoodsStockInfo{
			GoodsId: stock.GoodsId,
			Num:     stock.Num,
		}
		data = append(data, &GoodsStockInfo)
	}
	stockInfoList := &proto.StockInfoList{Data: data}
	return stockInfoList, nil

}

// 批量扣库存
func BatchReduceStock(ctx context.Context, req []*proto.GoodsStockInfo) (*proto.StockInfoList, error) {
	goodsIds := make([]int64, 0)
	nums := make([]int64, 0)
	for _, GoodsStockInfo := range req {
		goodsId := GoodsStockInfo.GoodsId
		goodsIds = append(goodsIds, goodsId)
		num := GoodsStockInfo.Num
		nums = append(nums, num)
	}
	data, err := mysql.BatchReduceStock(ctx, goodsIds, nums)
	if err != nil {
		return nil, err
	}
	res := make([]*proto.GoodsStockInfo, 0)
	for _, d := range *data {
		GoodsStockInfo := proto.GoodsStockInfo{
			GoodsId: d.GoodsId,
			Num:     d.Num,
		}
		res = append(res, &GoodsStockInfo)
	}
	stockInfoList := proto.StockInfoList{Data: res}
	return &stockInfoList, nil

}

// ReduceStockByGoodsId 根据商品id扣减库存
func ReduceStockByGoodsId(ctx context.Context, goodsId, num int64, orderId int64) error {
	// 执行数据库操作
	_, err := mysql.ReduceStock(ctx, goodsId, num, orderId)
	return err
}

// RollbackStock 回滚库存
func RollbackStock(ctx context.Context, data []*proto.GoodsStockInfo) error {
	return mysql.RollbackStock(ctx, data)
}
