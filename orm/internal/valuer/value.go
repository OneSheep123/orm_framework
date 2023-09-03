// Package valuer create by chencanhua in 2023/6/7
package valuer

import (
	"database/sql"
	"orm_framework/orm/model"
)

// Value 是对结构体实例的内部抽象
type Value interface {
	// Field 返回字段对应的值
	Field(name string) (any, error)
	// SetColumns 设置新值
	SetColumns(rows *sql.Rows) error
}

type Creator func(entity any, meta *model.Model) Value
