package model

// ORM
// struct -> table

type Stock struct {
	BaseModel // 嵌入默认的7个字段

	GoodsId int64
	Num     int64
	Lock    int64
}

// TableName 声明表名
func (Stock) TableName() string {
	return "xx_stock"
}
