// create by chencanhua in 2023/6/8
package valuer

import (
	"database/sql"
	"orm_framework/orm/internal/errs"
	"orm_framework/orm/model"
	"reflect"
)

type reflectValue struct {
	val  reflect.Value
	meta *model.Model
}

var _ Creator = NewReflectValue

// NewReflectValue 返回一个封装好的，基于反射实现的 Value
// 输入 val 必须是一个指向结构体实例的指针，而不能是任何其它类型
func NewReflectValue(val interface{}, meta *model.Model) Value {
	return reflectValue{
		val:  reflect.ValueOf(val).Elem(),
		meta: meta,
	}
}

func (r reflectValue) Field(name string) (any, error) {
	return r.val.FieldByName(name).Interface(), nil
}

func (r reflectValue) SetColumns(rows *sql.Rows) error {
	cs, err := rows.Columns()
	if err != nil {
		return err
	}
	if len(cs) > len(r.meta.FieldMap) {
		return errs.ErrTooManyReturnedColumns
	}

	vals := make([]any, 0, len(cs))
	valsElems := make([]reflect.Value, 0, len(cs))
	for _, c := range cs {
		fieldInfo, ok := r.meta.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		// 例如: fieldInfo.tOf = int, 那么这里value 是 *int
		value := reflect.New(fieldInfo.Type)
		vals = append(vals, value.Interface())
		// 记得调用Elem，因为fieldInfo.tOf = int, 那么这里value 是 *int
		valsElems = append(valsElems, value.Elem())
	}

	// 注意这里是vals...
	rows.Scan(vals...)

	for index, c := range cs {
		fieldInfo, ok := r.meta.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		r.val.FieldByName(fieldInfo.GoName).Set(valsElems[index])
	}
	return nil
}
