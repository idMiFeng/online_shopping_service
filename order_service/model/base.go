package model

import (
	"time"
)

type BaseModel struct {
	ID       uint      `gorm:"primaryKey"`
	CreateAt time.Time `gorm:"autoCreateTime"`   // 创建时间
	UpdateAt time.Time `gorm:"autoUpdateTime"`   // 更新时间
	CreateBy string    `gorm:"column:create_by"` // 指定数据库中的列名
	UpdateBy string
	Version  int16
	isDel    int8 `gorm:"index"`
}

// OrderGoodsStockInfo 订单商品信息
type OrderGoodsStockInfo struct {
	OrderId int64
	GoodsId int64
	Num     int64
}
