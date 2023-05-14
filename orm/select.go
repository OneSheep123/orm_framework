// Package orm create by chencanhua in 2023/5/7
package orm

import (
	"context"
	"orm_framework/orm/internal/errs"
	"strings"
)

var _ QueryBuilder = &Selector[any]{}

// Selector 用于构建Select语句
type Selector[T any] struct {
	model *model
	table string
	where []Predicate
	sb    strings.Builder
	args  []any
}

func (s *Selector[T]) Build() (*Query, error) {
	m, err := parseModel(new(T))
	if err != nil {
		return nil, err
	}
	s.model = m
	s.sb.WriteString("SELECT * FROM ")

	if s.table == "" {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.tableName)
		s.sb.WriteByte('`')
	} else {
		s.sb.WriteString(s.table)
	}

	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		pre := s.where[0]
		for index := 1; index < len(s.where); index++ {
			pre = pre.And(s.where[index])
		}
		if err := s.buildExpression(pre); err != nil {
			return nil, err
		}
	}
	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) buildExpression(expression Expression) error {
	if expression == nil {
		return nil
	}
	switch expr := expression.(type) {
	case Column:
		c, ok := s.model.fieldMap[expr.column]
		if !ok {
			return errs.NewErrUnknownField(expr.column)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(c.colName)
		s.sb.WriteByte('`')
	case Value:
		s.sb.WriteByte('?')
		s.args = append(s.args, expr.val)
	case Predicate:
		_, lp := expr.left.(Predicate)
		if lp {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(expr.left); err != nil {
			return err
		}
		if lp {
			s.sb.WriteByte(')')
		}
		s.sb.WriteByte(' ')
		s.sb.WriteString(string(expr.op))
		s.sb.WriteByte(' ')
		_, lp = expr.right.(Predicate)
		if lp {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(expr.right); err != nil {
			return err
		}
		if lp {
			s.sb.WriteByte(')')
		}
	}
	return nil
}

func (s *Selector[T]) Where(pre ...Predicate) *Selector[T] {
	s.where = pre
	return s
}

func (s *Selector[T]) From(tableName string) *Selector[T] {
	s.table = tableName
	return s
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Selector[T]) GetMulti(ctx context.Context) (*[]T, error) {
	//TODO implement me
	panic("implement me")
}
