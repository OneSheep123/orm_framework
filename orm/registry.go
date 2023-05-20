// create by chencanhua in 2023/5/16
package orm

import (
	"orm_framework/orm/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
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
		tagMap, err := r.parseTag(f.Tag)
		if err != nil {
			return nil, err
		}
		columnName := tagMap[tagKeyColumn]
		if columnName == "" {
			columnName = underscoreName(f.Name)
		}
		fieldMap[f.Name] = &field{
			colName: columnName,
		}
	}

	// 自定义表名
	var tableName string
	if tn, ok := entity.(TableName); ok {
		tableName = tn.TableName()
	}

	if tableName == "" {
		tableName = underscoreName(tOf.Name())
	}

	return &model{
		tableName: tableName,
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
