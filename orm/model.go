// Package orm create by chencanhua in 2023/5/14
package orm

import (
	"orm_framework/orm/internal/errs"
	"reflect"
)

type ModelOpt func(m *Model) error

// Model 元数据
type Model struct {
	tableName string
	fieldMap  map[string]*field
	columnMap map[string]*field
}

type field struct {
	colName   string
	fieldName string
	tOf       reflect.Type
}

// ModelWithColumnName 支持自定义字段名
func ModelWithColumnName(field string, name string) ModelOpt {
	return func(m *Model) error {
		f, ok := m.fieldMap[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}
		f.colName = name
		return nil
	}
}

// ModelWithTableName 支持自定义表名
func ModelWithTableName(tableName string) ModelOpt {
	return func(m *Model) error {
		m.tableName = tableName
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
