package mysql

import (
	"context"
	"github.com/idMiFeng/goods_service/errno"
	"github.com/idMiFeng/goods_service/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetGoodsByRoomId 根据roomID查直播间绑定的所有的商品 Id
func GetGoodsByRoomId(ctx context.Context, roomId int64) ([]*model.RoomGoods, error) {
	// 通过gorm去数据库中获取数据
	var data []*model.RoomGoods
	err := db.WithContext(ctx).
		Model(&model.RoomGoods{}).
		Where("room_id = ?", roomId).
		Order("weight").
		Find(&data).Error
	// 如果查询出错且不是空数据的错
	if err != nil && err != gorm.ErrEmptySlice {
		return nil, errno.ErrQueryFailed
	}
	return data, nil
}

// GetGoodsById 根据id查询商品信息
func GetGoodsById(ctx context.Context, idList []int64) ([]*model.Goods, error) {
	var data []*model.Goods
	//Expression: clause.Expr{SQL: "FIELD(goods_id,?)", Vars: []interface{}{idList}, WithoutParentheses: true}
	//这部分表示排序的 SQL 表达式，使用了 FIELD 函数来实现按照 idList 中的顺序进行排序。Vars 字段是一个参数，用于将 idList 的值传递给 SQL 表达式。
	err := db.WithContext(ctx).
		Model(&model.Goods{}).
		Where("goods_id in ?", idList). // 会按照idList顺序返回吗?
		Clauses(clause.OrderBy{
			Expression: clause.Expr{SQL: "FIELD(goods_id,?)", Vars: []interface{}{idList}, WithoutParentheses: true},
		}).
		Find(&data).Error
	if err != nil && err != gorm.ErrEmptySlice {
		return nil, errno.ErrQueryFailed
	}
	return data, nil
}
