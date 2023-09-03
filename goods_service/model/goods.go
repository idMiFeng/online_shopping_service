package model

type Goods struct {
	BaseModel // 嵌入默认的7个字段

	GoodsId     int64
	CategoryId  int64
	BrandName   string
	Code        int64
	Status      int8
	Title       string
	MarketPrice int64
	Price       int64
	Brief       string
	HeadImgs    string
	Videos      string
	Detail      string
	ExtJson     string
}

// TableName 声明表名
func (Goods) TableName() string {
	return "xx_goods"
}
