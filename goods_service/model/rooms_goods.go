package model

// RoomGoods 直播间与商品对应
type RoomGoods struct {
	BaseModel
	RoomId    int64
	GoodsId   int64
	Weight    int64
	IsCurrent int8 `gorm:"is_current"`
}

// TableName 声明表名
func (RoomGoods) TableName() string {
	return "xx_room_goods"
}
