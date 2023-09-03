package model

type StockRecord struct {
	BaseModel // 嵌入默认的7个字段

	OrderId int64
	GoodsId int64
	Num     int64
	Status  int32
}

// TableName 声明表名
func (StockRecord) TableName() string {
	return "xx_stock_record"
}
