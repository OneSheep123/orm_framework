// create by chencanhua in 2023/5/16
package orm

import (
	"orm_framework/orm/internal/errs"
	"reflect"
	"sync"
)

type registry struct {
	models sync.Map
}

func newRegistry() *registry {
	return &registry{}
}

// get 获取model数据
func (r *registry) get(val any) (*model, error) {
	tOf := reflect.TypeOf(val)
	m, ok := r.models.Load(tOf)
	if !ok {
		var err error
		m, err = r.parseModel(val)
		if err != nil {
			return nil, err
		}
		r.models.Store(tOf, m)
	}
	return m.(*model), nil
}

// parseModel 根据输入的entity，返回model数据
func (r *registry) parseModel(entity any) (*model, error) {
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
