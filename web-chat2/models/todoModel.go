package models

import (
	"gorm.io/gorm"
)

type Todo struct {
	gorm.Model //gormで定義されている構造体
	UserID     uint
	Content    string
}
