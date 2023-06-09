// Package orm create by chencanhua in 2023/5/7
package orm

import (
	"context"
)

type Selectable interface {
	selectable()
}

// Selector 用于构建Select语句
type Selector[T any] struct {
	builder
	table   string
	where   []Predicate
	having  []Predicate
	columns []Selectable
	groupBy []Column
	offset  int
	limit   int

	sess Session
}

func NewSelector[T any](sess Session) *Selector[T] {
	c := sess.getCore()
	return &Selector[T]{
		builder: builder{
			core:   c,
			quoter: c.dialect.quoter(),
		},
		sess: sess,
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
	if s.table == "" {
		s.quote(s.model.TableName)
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
			if err = s.buildColumn(c.column); err != nil {
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

// buildExpression 构建Where后面部分
// 这里case都实现了expr方法
func (s *Selector[T]) buildExpression(expression Expression) error {
	if expression == nil {
		return nil
	}
	switch expr := expression.(type) {
	case Column:
		return s.buildColumn(expr.column)
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
	if err := s.buildColumn(a.arg); err != nil {
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
			err := s.buildColumn(val.column)
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

func (s *Selector[T]) From(tableName string) *Selector[T] {
	s.table = tableName
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
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return getHandler[T](ctx, s.sess, s.core, qc)
	}
	for index := len(s.mdls) - 1; index >= 0; index-- {
		root = s.mdls[index](root)
	}
	res := root(ctx, qc)
	if res.Result != nil {
		return res.Result.(*T), nil
	}
	return nil, res.Err
}

func getHandler[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	sql, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Result: nil,
			Err:    err,
		}
	}
	rows, err := sess.queryContext(ctx, sql.SQL, sql.Args...)
	// 注意这里查询完后要进行关闭，否则连接会无法释放
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return &QueryResult{
			Result: nil,
			Err:    err,
		}
	}
	if !rows.Next() {
		// 这里调用的是error下的ErrNoRows
		return &QueryResult{
			Result: nil,
			Err:    ErrNoRows,
		}
	}

	tp := new(T)
	meta, err := c.r.Get(tp)
	if err != nil {
		return &QueryResult{
			Result: nil,
			Err:    ErrNoRows,
		}
	}
	val := c.Creator(tp, meta)
	err = val.SetColumns(rows)
	return &QueryResult{
		Result: tp,
		Err:    err,
	}
}

func (s *Selector[T]) GetMulti(ctx context.Context) (*[]T, error) {
	panic("implement me")
}
