// Package valuer create by chencanhua in 2023/6/7
package valuer

import (
	"database/sql"
	"orm_framework/orm/model"
)

type Value interface {
	Field(name string) (any, error)
	SetColumns(rows *sql.Rows) error
}

type Creator func(entity any, meta *model.Model) Value
