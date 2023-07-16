// Package orm create by chencanhua in 2023/6/19
package orm

import (
	"context"
	"database/sql"
	"orm_framework/orm/internal/errs"
	"orm_framework/orm/model"
)

// UpsertBuilder Inserter变为UpsertBuilder最后变为Inserter
// 其中UpsertBuilder去构建冲突部分
type UpsertBuilder[T any] struct {
	i *Inserter[T]
	// conflictColumns 这里只是作为临时变量存储，后续会赋值给Upsert内的conflictColumns
	conflictColumns []string
}

type Upsert struct {
	assigns         []Assignable
	conflictColumns []string
}

func (o *UpsertBuilder[T]) ConflictColumns(conflictColumns ...string) *UpsertBuilder[T] {
	o.conflictColumns = conflictColumns
	return o
}

// Update 也可以看做是一个终结方法，重新回到 Inserter 里面
func (o *UpsertBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicate = &Upsert{
		assigns:         assigns,
		conflictColumns: o.conflictColumns,
	}
	return o.i
}

type Inserter[T any] struct {
	values []*T
	builder
	columns []string
	// 使用一个 OnDuplicate 结构体，从而允许将来扩展更加复杂的行为
	onDuplicate *Upsert

	sess Session
}

func NewInserter[T any](sess Session) *Inserter[T] {
	c := sess.getCore()
	return &Inserter[T]{
		builder: builder{
			core:   c,
			quoter: c.dialect.quoter(),
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
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}
	m, err := i.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}
	i.model = m
	i.sb.WriteString("INSERT INTO ")
	i.quote(m.TableName)
	i.sb.WriteString("(")
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
		i.quote(field.ColName)
	}
	i.sb.WriteString(") VALUES ")
	for vIndex, val := range i.values {
		c := i.Creator(val, i.model)
		if vIndex > 0 {
			i.sb.WriteByte(',')
		}
		i.sb.WriteByte('(')
		for fIndex, field := range fields {
			if fIndex > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')
			v, err := c.Field(field.GoName)
			if err != nil {
				return nil, err
			}
			i.addArgs(v)
		}
		i.sb.WriteByte(')')
	}

	if i.onDuplicate != nil {
		err = i.dialect.buildOnUpsert(&i.builder, i.onDuplicate)
		if err != nil {
			return nil, err
		}
	}

	i.sb.WriteByte(';')
	return &Query{
		SQL:  i.sb.String(),
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
