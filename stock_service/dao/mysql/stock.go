package mysql

import (
	"context"
	"fmt"
	"github.com/idMiFeng/stock_service/dao/redis"
	"github.com/idMiFeng/stock_service/errno"
	"github.com/idMiFeng/stock_service/model"
	"github.com/idMiFeng/stock_service/proto"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// dao 层用来执行数据库相关的操作

// GetGoodsByRoomId 根据roomID查直播间绑定的所有的商品 Id
func GetStockByGoodsId(ctx context.Context, goodsId int64) (*model.Stock, error) {
	// 通过gorm去数据库中获取数据
	var data model.Stock
	err := db.WithContext(ctx).
		Model(&model.Stock{}).
		Where("goods_id = ?", goodsId).
		First(&data).Error
	// 如果查询出错且不是空数据的错
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errno.ErrQueryFailed
	}
	return &data, nil
}

// BatchGetStockByGoodsId 根据goodsIds切片返回商品切片
func BatchGetStockByGoodsId(ctx context.Context, goodsIds []int64) (*[]model.Stock, error) {
	var data []model.Stock
	err := db.WithContext(ctx).
		Model(&model.Stock{}).
		Where("goods_id IN (?)", goodsIds).
		Find(&data).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errno.ErrQueryFailed
	}
	return &data, nil

}

// BatchReduceStock 批量扣库存，需要使用事务，购物车清空要么同时成功要么同时失败
func BatchReduceStock(ctx context.Context, goodsIds, nums []int64) (*[]model.Stock, error) {
	var data []model.Stock
	db.Transaction(func(tx *gorm.DB) error {
		var d model.Stock
		for i, goodsId := range goodsIds {
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				WithContext(ctx).
				Where("goods_id = ?", goodsId).
				First(&d).Error
			if err != nil {
				return err
			}
			if d.Num < nums[i] {
				zap.L().Warn("understock", zap.Int64("goods_id", goodsId), zap.Error(err))
				return errno.ErrUnderstock
			}
			// 执行库存扣减
			d.Num -= nums[i]
			data = append(data, d)
			err = tx.
				WithContext(ctx).
				Save(&data).Error
			if err != nil {
				zap.L().Warn("ReduceStockByGoodsId save failed", zap.Int64("goods_id", goodsId), zap.Error(err))
				return err
			}
		}
		return nil
	})
	return &data, nil
}

// ReduceStockByGoodsId 扣减库存
func ReduceStockByGoodsId(ctx context.Context, goodsId, num int64) error {
	// 先查询库存数据
	var data model.Stock
	//return nil提交事务，return 任何 err都会回滚事务
	db.Transaction(func(tx *gorm.DB) error {
		//Clauses(clause.Locking{Strength: "UPDATE"}) 是用于在 GORM 中设置数据库查询的锁定方式，具体来说是使用了 "UPDATE" 锁定策略。这个锁定策略告诉数据库在查询期间锁定匹配的记录，以防止其他事务同时修改这些记录。
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			WithContext(ctx).
			Model(&model.Stock{}).
			Where("goods_id = ?", goodsId).
			First(&data).Error
		if err != nil {
			return err
		}
		// 判断库存是否充足
		if data.Num < num {
			zap.L().Warn("understock", zap.Int64("goods_id", goodsId), zap.Error(err))
			return errno.ErrUnderstock
		}
		// 执行库存扣减
		data.Num -= num
		err = tx.
			WithContext(ctx).
			Save(&data).Error
		if err != nil {
			zap.L().Warn("ReduceStockByGoodsId save failed", zap.Int64("goods_id", goodsId), zap.Error(err))
			return err
		}
		return nil
	})
	return nil
}

// RollbackStock 监听rocketmq消息进行库存回滚
func RollbackStockByMsg(ctx context.Context, data model.OrderGoodsStockInfo) error {
	// 先查询库存数据，需要放到事务操作中
	db.Transaction(func(tx *gorm.DB) error {
		var sr model.StockRecord
		err := tx.WithContext(ctx).
			Model(&model.StockRecord{}).
			Where("order_id = ? and goods_id = ? and status = 1", data.OrderId, data.GoodsId).
			First(&sr).Error
		// 没找到记录
		// 压根就没记录或者已经回滚过 不需要后续操作
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		if err != nil {
			zap.L().Error("query stock_record by order_id failed", zap.Error(err), zap.Int64("order_id", data.OrderId), zap.Int64("goods_id", data.GoodsId))
			return err
		}
		// 开始归还库存
		var s model.Stock
		err = tx.WithContext(ctx).
			Model(&model.Stock{}).
			Where("goods_id = ?", data.GoodsId).
			First(&s).Error
		if err != nil {
			zap.L().Error("query stock by goods_id failed", zap.Error(err), zap.Int64("goods_id", data.GoodsId))
			return err
		}
		s.Num += data.Num  // 库存加上
		s.Lock -= data.Num // 锁定的库存减掉
		if s.Lock < 0 {    // 预扣库存不能为负
			return errno.ErrRollbackstockFailed
		}
		err = tx.WithContext(ctx).Save(&s).Error
		if err != nil {
			zap.L().Warn("RollbackStock stock save failed", zap.Int64("goods_id", s.GoodsId), zap.Error(err))
			return err
		}
		// 将库存扣减记录的状态变更为已回滚
		sr.Status = 3
		err = tx.WithContext(ctx).Save(&sr).Error
		if err != nil {
			zap.L().Warn("RollbackStock stock_record save failed", zap.Int64("goods_id", s.GoodsId), zap.Error(err))
			return err
		}
		return nil
	})
	return nil
}

// RollbackStock 回滚库存
func RollbackStock(ctx context.Context, data []*proto.GoodsStockInfo) error {
	// 先查询库存数据，需要放到事务操作中
	db.Transaction(func(tx *gorm.DB) error {
		for _, item := range data {
			var s model.Stock
			err := tx.WithContext(ctx).
				Model(&model.Stock{}).
				Where("goods_id = ?", item.GoodsId).
				First(&s).Error
			if err != nil {
				return err
			}
			// 归还库存
			s.Num += item.Num
			err = tx.WithContext(ctx).Save(&s).Error
			if err != nil {
				zap.L().Warn("RollbackStock save failed", zap.Int64("goods_id", s.GoodsId), zap.Error(err))
				return err
			}
		}
		return nil
	})
	return nil
}

// ReduceStock 扣减库存 基于redis分布式锁版本
func ReduceStock(ctx context.Context, goodsId, num, orderId int64) (*model.Stock, error) {
	// 1. 查询现有库存
	var data model.Stock
	// 创建key
	mutexname := fmt.Sprintf("xx-stock-%d", goodsId)
	// 创建锁
	mutex := redis.Rs.NewMutex(mutexname)
	// 获取锁
	if err := mutex.Lock(); err != nil {
		return nil, errno.ErrReducestockFailed
	}
	// 此时data可能都是旧数据 data.num = 99  实际上数据库中num=97
	defer mutex.Unlock() // 释放锁
	// 获取锁成功
	// 开启事务
	db.Transaction(func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).
			Model(&model.Stock{}).
			Where("goods_id = ?", goodsId).
			First(&data).Error
		if err != nil {
			return err
		}
		// 2. 校验；现有库存数>0 且 大于等于num
		if data.Num-num < 0 {
			return errno.ErrUnderstock
		}
		// 3. 扣减
		data.Num -= num  // 库存-
		data.Lock += num // 预扣库存+
		// 保存
		err = tx.WithContext(ctx).
			Save(&data).Error // save更新所有字段！  97 -> 99 要保证更新的数据是准确的。
		if err != nil {
			zap.L().Error(
				"reduceStock save failed",
				zap.Int64("goods_id", goodsId),
			)
			return err
		}
		// 创建库存记录表
		stockRecord := model.StockRecord{
			OrderId: orderId,
			GoodsId: goodsId,
			Num:     num,
			Status:  1, // 预扣减
		}
		err = tx.WithContext(ctx).
			Model(&model.StockRecord{}).
			Create(&stockRecord).Error
		if err != nil {
			zap.L().Error("create StockRecord failed", zap.Error(err))
			return err
		}
		return nil
	})
	return &data, nil
}
