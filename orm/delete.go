// Package orm create by chencanhua in 2023/5/14
package orm

import (
	"orm_framework/orm/internal/errs"
	"strings"
)

var _ QueryBuilder = &Deleter[any]{}

type Deleter[T any] struct {
	m     *model
	db    *DB
	table string
	args  []any
	sb    strings.Builder
	where []Predicate
}

func NewDeleter[T any](db *DB) *Deleter[T] {
	return &Deleter[T]{
		db: db,
	}
}

func (d *Deleter[T]) From(name string) *Deleter[T] {
	d.table = name
	return d
}

func (d *Deleter[T]) Where(pre ...Predicate) *Deleter[T] {
	d.where = pre
	return d
}

func (d *Deleter[T]) Build() (*Query, error) {
	m, err := d.db.r.get(new(T))
	if err != nil {
		return nil, err
	}
	d.m = m
	d.sb.WriteString("DELETE FROM ")

	if d.table != "" {
		d.sb.WriteString(d.table)
	} else {
		d.sb.WriteByte('`')
		d.sb.WriteString(d.m.tableName)
		d.sb.WriteByte('`')
	}

	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")
		pre := d.where[0]
		for index := 1; index < len(d.where); index++ {
			pre.And(d.where[index])
		}
		if err := d.buildExpression(pre); err != nil {
			return nil, err
		}
	}
	d.sb.WriteByte(';')
	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, nil
}

func (d *Deleter[T]) buildExpression(expression Expression) error {
	if expression == nil {
		return nil
	}
	switch expr := expression.(type) {
	case Column:
		c, ok := d.m.fieldMap[expr.column]
		if !ok {
			return errs.NewErrUnknownField(expr.column)
		}
		d.sb.WriteByte('`')
		d.sb.WriteString(c.colName)
		d.sb.WriteByte('`')
	case Value:
		d.sb.WriteByte('?')
		d.args = append(d.args, expr.val)
	case Predicate:
		_, lp := expr.left.(Predicate)
		if lp {
			d.sb.WriteByte('(')
		}
		if err := d.buildExpression(expr.left); err != nil {
			return err
		}
		if lp {
			d.sb.WriteByte(')')
		}
		d.sb.WriteByte(' ')
		d.sb.WriteString(string(expr.op))
		d.sb.WriteByte(' ')
		_, rp := expr.right.(Predicate)
		if rp {
			d.sb.WriteByte('(')
		}
		if err := d.buildExpression(expr.right); err != nil {
			return err
		}
		if rp {
			d.sb.WriteByte(')')
		}
	}
	return nil
}
