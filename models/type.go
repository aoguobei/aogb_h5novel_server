package models

// Type 类型模型
type Type struct {
	ID   uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"type:varchar(50);not null;comment:类型名称"`
	Code string `json:"code" gorm:"type:varchar(20);not null;uniqueIndex;comment:类型代码"`
}

// TableName 指定表名
func (Type) TableName() string {
	return "types"
}
