// Package orm create by chencanhua in 2023/6/19
package orm

import (
	"orm_framework/orm/internal/errs"
	"orm_framework/orm/model"
	"reflect"
	"strings"
)

type OnDuplicateKeyBuilder[T any] struct {
	i *Inserter[T]
}

type OnDuplicateKey struct {
	assigns []Assignable
}

func (o *OnDuplicateKeyBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicate = &OnDuplicateKey{
		assigns: assigns,
	}
	return o.i
}

type Inserter[T any] struct {
	values  []*T
	db      *DB
	sb      strings.Builder
	columns []string
	// 使用一个 OnDuplicate 结构体，从而允许将来扩展更加复杂的行为
	onDuplicate *OnDuplicateKey
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		db: db,
	}
}

// Columns 指定列，注意这里是结构体的元素
func (i *Inserter[T]) Columns(columns ...string) *Inserter[T] {
	i.columns = columns
	return i
}

func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

func (i *Inserter[T]) OnDuplicateKey() *OnDuplicateKeyBuilder[T] {
	return &OnDuplicateKeyBuilder[T]{
		i: i,
	}
}

func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}
	m, err := i.db.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}
	i.sb.WriteString("INSERT INTO `")
	i.sb.WriteString(m.TableName)
	i.sb.WriteString("`(")
	fields := m.Fields
	if len(i.columns) != 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, goColumn := range i.columns {
			field, ok := m.FieldMap[goColumn]
			if !ok {
				return nil, errs.NewErrUnknownField(goColumn)
			}
			fields = append(fields, field)
		}
	}
	for index, field := range fields {
		if index > 0 {
			i.sb.WriteByte(',')
		}
		i.sb.WriteByte('`')
		i.sb.WriteString(field.ColName)
		i.sb.WriteByte('`')
	}
	i.sb.WriteString(") VALUES")
	args := make([]any, 0, len(i.values)*len(fields))
	for vIndex, val := range i.values {
		if vIndex > 0 {
			i.sb.WriteByte(',')
		}
		i.sb.WriteByte('(')
		for fIndex, field := range fields {
			if fIndex > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')
			v := reflect.ValueOf(val).Elem().FieldByName(field.GoName).Interface()
			args = append(args, v)
		}
		i.sb.WriteByte(')')
	}

	if i.onDuplicate != nil {
		i.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
		for index, a := range i.onDuplicate.assigns {
			if index > 0 {
				i.sb.WriteByte(',')
			}
			switch assign := a.(type) {
			case Assignment:
				i.sb.WriteByte('`')
				field, ok := m.FieldMap[assign.column]
				if !ok {
					return nil, errs.NewErrUnknownField(assign.column)
				}
				i.sb.WriteString(field.ColName)
				i.sb.WriteByte('`')
				i.sb.WriteString(`=?`)
				args = append(args, assign.val)
			case Column:
				i.sb.WriteByte('`')
				fd, ok := m.FieldMap[assign.column]
				if !ok {
					return nil, errs.NewErrUnknownField(assign.column)
				}
				i.sb.WriteString(fd.ColName)
				i.sb.WriteString("`=VALUES(`")
				i.sb.WriteString(fd.ColName)
				i.sb.WriteString("`)")
			}
		}
	}

	i.sb.WriteByte(';')
	return &Query{
		SQL:  i.sb.String(),
		Args: args,
	}, nil
}
