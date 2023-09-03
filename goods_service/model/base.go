package model

import (
	"time"
)

type BaseModel struct {
	ID       uint `gorm:"primaryKey"`
	CreateAt time.Time
	UpdateAt time.Time
	CreateBy string
	UpdateBy string
	Version  int16
	isDel    int8 `gorm:"index"`
}
