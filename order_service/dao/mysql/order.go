package mysql

import (
	"context"
	"github.com/idMiFeng/order_service/model"

	"gorm.io/gorm"
)

func QueryOrder(ctx context.Context, orderId int64) (model.Order, error) {
	var data model.Order
	err := db.WithContext(ctx).
		Model(&model.Order{}).
		Where("order_id = ?", orderId).
		First(&data).Error
	return data, err
}

func UpdateOrder(ctx context.Context, data model.Order) error {
	return db.WithContext(ctx).
		Model(&model.Order{}).
		Where("order_id = ?", data.OrderId).
		Updates(&data).Error
}

func CreateOrder(ctx context.Context, data *model.Order) error {
	return db.WithContext(ctx).
		Model(&model.Order{}).
		Save(data).Error
}

func CreateOrderDetail(ctx context.Context, data *model.OrderDetail) error {
	return db.WithContext(ctx).
		Model(&model.OrderDetail{}).
		Save(data).Error
}

// CreateOrderWithTransation 创建订单事务处理
func CreateOrderWithTransation(ctx context.Context, order *model.Order, orderDetail *model.OrderDetail) error {
	return db.WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
			if err := tx.Create(order).Error; err != nil {
				// 返回任何错误都会回滚事务
				return err
			}

			if err := tx.Create(orderDetail).Error; err != nil {
				return err
			}
			// 返回 nil 提交事务
			return nil
		})
}
