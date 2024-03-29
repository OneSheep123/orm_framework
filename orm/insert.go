// Package orm create by chencanhua in 2023/6/19
package orm

import (
	"context"
	"database/sql"
	"github.com/valyala/bytebufferpool"
	"orm_framework/orm/internal/errs"
	"orm_framework/orm/model"
)

var _ QueryBuilder = &Inserter[any]{}

type Inserter[T any] struct {
	values []*T
	insertBuilder
	sess Session
}

func NewInserter[T any](sess Session) *Inserter[T] {
	c := sess.getCore()
	return &Inserter[T]{
		insertBuilder: insertBuilder{
			builder: builder{
				core:   c,
				quoter: c.dialect.quoter(),
				buffer: bytebufferpool.Get(),
			},
		},
		sess: sess,
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

func (i *Inserter[T]) OnDuplicateKey() *UpsertBuilder[T] {
	return &UpsertBuilder[T]{
		i: i,
	}
}

func (i *Inserter[T]) Build() (*Query, error) {
	defer bytebufferpool.Put(i.buffer)
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}
	m, err := i.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}
	i.model = m
	i.writeString("INSERT INTO ")
	i.quote(m.TableName)
	i.writeString("(")
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
			i.writeByte(',')
		}
		i.quote(field.ColName)
	}
	i.writeString(") VALUES ")
	for vIndex, val := range i.values {
		c := i.Creator(val, i.model)
		if vIndex > 0 {
			i.writeByte(',')
		}
		i.writeByte('(')
		for fIndex, field := range fields {
			if fIndex > 0 {
				i.writeByte(',')
			}
			i.writeByte('?')
			v, err := c.Field(field.GoName)
			if err != nil {
				return nil, err
			}
			i.addArgs(v)
		}
		i.writeByte(')')
	}

	if i.onDuplicate != nil {
		err = i.dialect.buildOnUpsert(&i.builder, i.onDuplicate)
		if err != nil {
			return nil, err
		}
	}

	i.writeByte(';')
	return &Query{
		SQL:  i.buffer.String(),
		Args: i.args,
	}, nil
}

func (i *Inserter[T]) Exec(ctx context.Context) sql.Result {
	qc := &QueryContext{
		Type:    "INSERT",
		Builder: i,
	}
	result := exec(ctx, i.sess, i.core, qc)
	if result.Result != nil {
		return &Result{
			res: result.Result.(sql.Result),
			err: nil,
		}
	}
	return &Result{
		err: result.Err,
	}
}
