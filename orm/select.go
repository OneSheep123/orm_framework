// Package orm create by chencanhua in 2023/5/7
package orm

import (
	"context"
	"orm_framework/orm/internal/errs"
	"orm_framework/orm/model"
	"strings"
)

var _ QueryBuilder = &Selector[any]{}

type Selectable interface {
	selectable()
}

// Selector 用于构建Select语句
type Selector[T any] struct {
	model   *model.Model
	db      *DB
	table   string
	where   []Predicate
	args    []any
	sb      strings.Builder
	columns []Selectable
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	m, err := s.db.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	s.model = m
	s.sb.WriteString("SELECT ")
	err = s.buildColumns()
	if err != nil {
		return nil, err
	}
	s.sb.WriteString(" FROM ")
	if s.table == "" {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.TableName)
		s.sb.WriteByte('`')
	} else {
		s.sb.WriteString(s.table)
	}

	if len(s.where) > 0 {
		// 类似这种可有可无的部分，都要在前面加一个空格
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
		c, ok := s.model.FieldMap[expr.column]
		if !ok {
			return errs.NewErrUnknownField(expr.column)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(c.ColName)
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

func (s *Selector[T]) buildColumns() error {
	if len(s.columns) == 0 {
		s.sb.WriteByte('*')
		return nil
	}
	for i, c := range s.columns {
		if i > 0 {
			s.sb.WriteByte(',')
		}
		switch val := c.(type) {
		case Column:
			s.sb.WriteByte('`')
			fd, ok := s.model.FieldMap[val.column]
			if !ok {
				return errs.NewErrUnknownField(val.column)
			}
			s.sb.WriteString(fd.ColName)
			s.sb.WriteByte('`')
		case Aggregate:
			s.sb.WriteString(val.fn)
			s.sb.WriteString("(`")
			fd, ok := s.model.FieldMap[val.arg]
			if !ok {
				return errs.NewErrUnknownField(val.arg)
			}
			s.sb.WriteString(fd.ColName)
			s.sb.WriteString("`)")
		}
	}
	return nil
}

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
	return s
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
	sql, err := s.Build()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.db.QueryContext(ctx, sql.SQL, sql.Args...)
	// 注意这里查询完后要进行关闭，否则连接会无法释放
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		// 这里调用的是error下的ErrNoRows
		return nil, ErrNoRows
	}

	tp := new(T)
	meta, err := s.db.r.Get(tp)
	if err != nil {
		return nil, err
	}
	val := s.db.Creator(tp, meta)
	err = val.SetColumns(rows)
	return tp, err
}

func (s *Selector[T]) GetMulti(ctx context.Context) (*[]T, error) {
	panic("implement me")
}
