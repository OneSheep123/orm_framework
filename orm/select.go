// Package orm create by chencanhua in 2023/5/7
package orm

import (
	"context"
)

var _ QueryBuilder = &Selector[any]{}

// Selector 用于构建Select语句
type Selector[T any] struct {
	model *model
	table string
	where []Predicate
	builder
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
		if err := s.buildPredicate(s.where, s.model); err != nil {
			return nil, err
		}
	}
	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
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
