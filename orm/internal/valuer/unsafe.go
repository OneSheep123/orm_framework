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
	address unsafe.Pointer
	meta    *model.Model
}

var _ Creator = NewUnsafeValue

func NewUnsafeValue(entity interface{}, meta *model.Model) Value {
	return unsafeValue{
		address: unsafe.Pointer(reflect.ValueOf(entity).Pointer()),
		meta:    meta,
	}
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
		ptr := unsafe.Pointer(uintptr(u.address) + cm.Offset)
		val := reflect.NewAt(cm.Type, ptr)
		colValues[i] = val.Interface()
	}
	return rows.Scan(colValues...)
}
