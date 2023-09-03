// Package orm create by chencanhua in 2023/5/7
package orm

import (
	"context"
	"orm_framework/orm/internal/errs"
)

type Selectable interface {
	selectable()
}

// Selector 用于构建Select语句
type Selector[T any] struct {
	// 表别名
	table TableReference
	// select语句构建元素
	selectorBuilder
	sess Session
}

func NewSelector[T any](sess Session) *Selector[T] {
	c := sess.getCore()
	return &Selector[T]{
		sess: sess,
		selectorBuilder: selectorBuilder{
			builder: builder{
				core:   c,
				quoter: c.dialect.quoter(),
			},
		},
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	m, err := s.r.Get(new(T))
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

	if err = s.buildTable(s.table); err != nil {
		return nil, err
	}

	if len(s.where) > 0 {
		// 类似这种可有可无的部分，都要在前面加一个空格
		s.sb.WriteString(" WHERE ")
		pre := s.where[0]
		for index := 1; index < len(s.where); index++ {
			pre = pre.And(s.where[index])
		}
		if err = s.buildExpression(pre); err != nil {
			return nil, err
		}
	}

	if len(s.groupBy) > 0 {
		s.sb.WriteString(" GROUP BY ")
		for i, c := range s.groupBy {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			if err = s.buildColumn(&Column{column: c.column}); err != nil {
				return nil, err
			}
		}
	}

	if len(s.having) > 0 {
		s.sb.WriteString(" HAVING ")
		// HAVING 是可以用别名的
		if err = s.buildPredicates(s.having); err != nil {
			return nil, err
		}
	}

	if s.limit > 0 {
		s.sb.WriteString(" LIMIT ?")
		s.addArgs(s.limit)
	}

	if s.offset > 0 {
		s.sb.WriteString(" OFFSET ?")
		s.addArgs(s.offset)
	}

	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) buildTable(table TableReference) error {
	switch t := table.(type) {
	case nil:
		s.quote(s.model.TableName)
	case Table:
		model, err := s.r.Get(t.entity)
		if err != nil {
			return err
		}
		s.quote(model.TableName)
		if t.alias != "" {
			s.sb.WriteString(" AS ")
			s.quote(t.alias)
		}
	case Join:
		s.sb.WriteByte('(')

		err := s.buildTable(t.left)
		if err != nil {
			return err
		}

		s.sb.WriteByte(' ')
		s.sb.WriteString(t.typ)
		s.sb.WriteByte(' ')

		err = s.buildTable(t.right)
		if err != nil {
			return err
		}

		if len(t.using) > 0 {
			s.sb.WriteString(" USING (")
			for i, col := range t.using {
				if i > 0 {
					s.sb.WriteByte(',')
				}
				err = s.buildColumn(&Column{column: col})
				if err != nil {
					return err
				}
			}
			s.sb.WriteByte(')')
		} else if len(t.on) > 0 {
			s.sb.WriteString(" ON ")
			p := t.on[0]
			for i := 1; i < len(t.on); i++ {
				p = p.And(t.on[i])
			}
			if err = s.buildExpression(p); err != nil {
				return err
			}
		}
		s.sb.WriteByte(')')
	default:
		return errs.NewErrUnsupportedTable(table)
	}
	return nil
}

// buildExpression 构建Where后面部分
// 这里case都实现了expr方法
func (s *Selector[T]) buildExpression(expression Expression) error {
	if expression == nil {
		return nil
	}
	switch expr := expression.(type) {
	case Column:
		return s.buildColumn(&expr)
	case Aggregate:
		return s.buildAggregate(expr, false)
	case Value:
		s.sb.WriteByte('?')
		s.addArgs(expr.val)
	case RawExpr:
		s.sb.WriteString(expr.raw)
		s.addArgs(expr.args...)
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

		// 可能只有左边
		if expr.op == "" {
			return nil
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

func (s *Selector[T]) buildPredicates(ps []Predicate) error {
	p := ps[0]
	for i := 1; i < len(ps); i++ {
		p = p.And(ps[i])
	}
	return s.buildExpression(p)
}

func (s *Selector[T]) buildAggregate(a Aggregate, useAlias bool) error {
	s.sb.WriteString(a.fn)
	s.sb.WriteByte('(')
	if err := s.buildColumn(&Column{column: a.arg}); err != nil {
		return err
	}
	s.sb.WriteByte(')')
	if useAlias {
		s.buildAs(a.alias)
	}
	return nil
}

// buildColumns 构建select后面部分
// 这里的case都有实现了selectable接口
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
			err := s.buildColumn(&Column{column: val.column})
			if err != nil {
				return err
			}
			s.buildAs(val.alias)
		case Aggregate:
			return s.buildAggregate(val, true)
		case RawExpr:
			s.sb.WriteString(val.raw)
			s.addArgs(val.args...)
		}
	}
	return nil
}

// buildAs 构建as
func (s *Selector[T]) buildAs(alias string) {
	if alias != "" {
		s.sb.WriteString(" AS ")
		s.quote(alias)
	}
}

func (s *Selector[T]) addArgs(args ...any) {
	if len(args) == 0 {
		return
	}
	if s.args == nil {
		s.args = make([]any, 0, 8)
	}
	s.args = append(s.args, args...)
}

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
	return s
}

func (s *Selector[T]) Where(pre ...Predicate) *Selector[T] {
	s.where = pre
	return s
}

func (s *Selector[T]) From(table TableReference) *Selector[T] {
	s.table = table
	return s
}

// GroupBy 设置 group by 子句
func (s *Selector[T]) GroupBy(cols ...Column) *Selector[T] {
	s.groupBy = cols
	return s
}

func (s *Selector[T]) Having(ps ...Predicate) *Selector[T] {
	s.having = ps
	return s
}

func (s *Selector[T]) Offset(offset int) *Selector[T] {
	s.offset = offset
	return s
}

func (s *Selector[T]) Limit(limit int) *Selector[T] {
	s.limit = limit
	return s
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	qc := &QueryContext{
		Type:    "SELECT",
		Builder: s,
	}
	res := get[T](ctx, s.sess, s.core, qc)
	if res.Result != nil {
		return res.Result.(*T), nil
	}
	return nil, res.Err
}

func (s *Selector[T]) GetMulti(ctx context.Context) (*[]T, error) {
	panic("implement me")
}
