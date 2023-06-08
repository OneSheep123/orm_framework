// Package model Package orm create by chencanhua in 2023/5/14
package model

import (
	"orm_framework/orm/internal/errs"
	"reflect"
)

type ModelOpt func(m *Model) error

// Model 元数据
type Model struct {
	TableName string
	FieldMap  map[string]*Field
	ColumnMap map[string]*Field
}

type Field struct {
	ColName string
	GoName  string
	Type    reflect.Type
	Offset  uintptr
}

// ModelWithColumnName 支持自定义字段名
func ModelWithColumnName(field string, name string) ModelOpt {
	return func(m *Model) error {
		f, ok := m.FieldMap[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}
		f.ColName = name
		return nil
	}
}

// ModelWithTableName 支持自定义表名
func ModelWithTableName(tableName string) ModelOpt {
	return func(m *Model) error {
		m.TableName = tableName
		return nil
	}
}

// 我们支持的全部标签上的 key 都放在这里
// 方便用户查找，和我们后期维护
const (
	tagKeyColumn = "column"
)

type TableName interface {
	TableName() string
}
