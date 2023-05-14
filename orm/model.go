// Package orm create by chencanhua in 2023/5/14
package orm

import (
	"orm_framework/orm/internal/errs"
	"reflect"
	"unicode"
)

type model struct {
	tableName string
	fieldMap  map[string]*field
}

type field struct {
	colName string
}

// parseModel 根据输入的entity，返回model数据
func parseModel(entity any) (*model, error) {
	tOf := reflect.TypeOf(entity)
	// 只允许一级指针结构体
	if tOf.Kind() != reflect.Pointer || tOf.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointOnly
	}
	tOf = tOf.Elem()
	numField := tOf.NumField()
	fieldMap := map[string]*field{}

	for index := 0; index < numField; index++ {
		f := tOf.Field(index)
		fieldMap[f.Name] = &field{
			colName: underscoreName(f.Name),
		}
	}

	return &model{
		tableName: underscoreName(tOf.Name()),
		fieldMap:  fieldMap,
	}, nil
}

// underscoreName 驼峰转字符串命名
// eg: TestModel => test_model
// ID => i_d
func underscoreName(name string) string {
	var buf []byte
	for i, v := range name {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}

	}
	return string(buf)
}
