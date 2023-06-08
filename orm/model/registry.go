// Package orm create by chencanhua in 2023/5/16
package model

import (
	"orm_framework/orm/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

type Registry interface {
	Get(val any) (*Model, error)
	Register(val any, opts ...ModelOpt) (*Model, error)
}

var _ Registry = &registry{}

type registry struct {
	models sync.Map
}

func NewRegistry() Registry {
	return &registry{}
}

// Get 获取model数据
func (r *registry) Get(val any) (*Model, error) {
	tOf := reflect.TypeOf(val)
	m, ok := r.models.Load(tOf)
	if ok {
		return m.(*Model), nil
	}
	return r.Register(val)
}

// Register 元数据注册
func (r *registry) Register(val any, opts ...ModelOpt) (*Model, error) {
	model, err := r.parseModel(val)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		err = opt(model)
		if err != nil {
			return nil, err
		}
	}
	r.models.Store(reflect.TypeOf(val), model)
	return model, nil
}

// parseTag 获取标签
func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag := tag.Get("orm")
	if ormTag == "" {
		return map[string]string{}, nil
	}
	// 这个初始化容量就是我们支持的 key 的数量，
	// 现在只有一个，所以我们初始化为 1
	res := make(map[string]string, 1)

	pairs := strings.Split(ormTag, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, errs.NewErrInvalidTagContent(pair)
		}
		res[kv[0]] = kv[1]
	}
	return res, nil
}

// parseModel 根据输入的entity，返回model数据
func (r *registry) parseModel(entity any) (*Model, error) {
	tOf := reflect.TypeOf(entity)
	// 只允许一级指针结构体
	if tOf.Kind() != reflect.Pointer || tOf.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointOnly
	}
	tOf = tOf.Elem()
	numField := tOf.NumField()
	fieldMap := map[string]*Field{}
	columnMap := map[string]*Field{}
	for index := 0; index < numField; index++ {
		f := tOf.Field(index)
		tagMap, err := r.parseTag(f.Tag)
		if err != nil {
			return nil, err
		}
		columnName := tagMap[tagKeyColumn]
		if columnName == "" {
			columnName = underscoreName(f.Name)
		}
		fieldInfo := &Field{
			ColName: columnName,
			GoName:  f.Name,
			Type:    f.Type,
			Offset:  f.Offset,
		}
		fieldMap[f.Name] = fieldInfo
		columnMap[columnName] = fieldInfo
	}

	// 自定义表名
	var tableName string
	if tn, ok := entity.(TableName); ok {
		tableName = tn.TableName()
	}

	if tableName == "" {
		tableName = underscoreName(tOf.Name())
	}

	return &Model{
		TableName: tableName,
		FieldMap:  fieldMap,
		ColumnMap: columnMap,
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
