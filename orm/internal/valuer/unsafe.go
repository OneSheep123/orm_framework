// create by chencanhua in 2023/6/8
package valuer

import (
	"database/sql"
	"orm_framework/orm/internal/errs"
	"orm_framework/orm/model"
	"reflect"
	"unsafe"
)

type unsafeValue struct {
	// address 元素的地址
	address unsafe.Pointer
	// meta 元数据
	meta *model.Model
}

var _ Creator = NewUnsafeValue

func NewUnsafeValue(entity interface{}, meta *model.Model) Value {
	return unsafeValue{
		address: unsafe.Pointer(reflect.ValueOf(entity).Pointer()),
		meta:    meta,
	}
}

func (u unsafeValue) Field(name string) (any, error) {
	field, ok := u.meta.FieldMap[name]
	if !ok {
		return nil, errs.NewErrUnknownField(name)
	}
	ptr := unsafe.Pointer(uintptr(u.address) + field.Offset)
	// 这里NewAt是创建了指定地址的类型变量，不会修改对应地址的值
	val := reflect.NewAt(field.Type, ptr).Elem()
	return val.Interface(), nil
}

func (u unsafeValue) SetColumns(rows *sql.Rows) error {
	cs, err := rows.Columns()
	if err != nil {
		return err
	}
	if len(cs) > len(u.meta.FieldMap) {
		return errs.ErrTooManyReturnedColumns
	}

	colValues := make([]any, len(cs))
	for i, c := range cs {
		cm, ok := u.meta.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		// 结构体的地址 + 对应字段在结构体中的偏移量
		ptr := unsafe.Pointer(uintptr(u.address) + cm.Offset)
		// 在特定地址创建值
		val := reflect.NewAt(cm.Type, ptr)
		colValues[i] = val.Interface()
	}
	return rows.Scan(colValues...)
}
