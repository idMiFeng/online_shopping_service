package model

// ORM
// struct -> table

type Order struct {
	BaseModel // 嵌入默认的7个字段

	OrderId   int64
	UserId    int64
	PayAmount int64
	Status    int32

	ReceiveAddress string
	ReceiveName    string
	ReceivePhone   string
}

// TableName 声明表名
func (Order) TableName() string {
	return "xx_order"
}
