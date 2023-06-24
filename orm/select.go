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
	db      *DB
	table   string
	where   []Predicate
	group   []Expression
	having  []Predicate
	orderby []Column
	columns []Selectable
	limit   int64
	offset  int64
	builder
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
		},
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

	if len(s.group) > 0 {
		s.sb.WriteString(" GROUP BY ")
		if err = s.buildGroupBy(); err != nil {
			return nil, err
		}
	}

	if len(s.having) > 0 {
		s.sb.WriteString(" HAVING ")
		pre := s.having[0]
		for index := 1; index < len(s.having); index++ {
			pre = pre.And(s.having[index])
		}
		if err = s.buildExpression(pre); err != nil {
			return nil, err
		}
	}

	if len(s.orderby) > 0 {
		s.sb.WriteString(" ORDER BY ")
		if err = s.buildSortBy(); err != nil {
			return nil, err
		}
	}

	if s.offset < 0 {
		return nil, errs.ErrOffset
	} else if s.limit > 0 {
		s.sb.WriteString(" OFFSET ")
		s.sb.WriteByte('?')
		s.addArgs(s.offset)
		s.sb.WriteByte(',')
		s.sb.WriteString(" LIMIT ")
		s.sb.WriteByte('?')
		s.addArgs(s.limit)
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
		err := s.buildColumn(expr.column)
		if err != nil {
			return err
		}
	case Value:
		s.sb.WriteByte('?')
		s.addArgs(expr.val)
	case RawExpr:
		s.sb.WriteString(expr.raw)
		s.addArgs(expr.args...)
	case Aggregate:
		s.sb.WriteString(expr.fn)
		s.sb.WriteByte('(')
		err := s.buildColumn(expr.arg)
		if err != nil {
			return err
		}
		s.sb.WriteByte(')')
		s.sb.WriteString(expr.op)
		s.sb.WriteByte('?')
		s.addArgs(expr.val)
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
			s.sb.WriteString(val.fn)
			s.sb.WriteString("(")
			err := s.buildColumn(val.arg)
			if err != nil {
				return err
			}
			s.sb.WriteString(")")
			s.buildAs(val.alias)
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

// buildGroupBy 构建groupBy
func (s *Selector[T]) buildGroupBy() error {
	for index, c := range s.group {
		if index > 0 {
			s.sb.WriteByte(',')
		}
		switch expr := c.(type) {
		case Column:
			field, ok := s.model.FieldMap[expr.column]
			if !ok {
				return errs.NewErrUnknownField(expr.column)
			}
			s.sb.WriteByte('`')
			s.sb.WriteString(field.ColName)
			s.sb.WriteByte('`')
		default:
			return errs.NewErrUnsupportedExpressionType(c)
		}
	}
	return nil
}

// buildSortBy 构建排序
func (s *Selector[T]) buildSortBy() error {
	for index, sby := range s.orderby {
		if index > 0 {
			s.sb.WriteByte(',')
		}
		if err := s.buildColumn(sby.column); err != nil {
			return err
		}
		s.sb.WriteByte(' ')
		if sby.sort == "" {
			s.sb.WriteString("ASC")
		} else {
			s.sb.WriteString(string(sby.sort))
		}
	}
	return nil
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

func (s *Selector[T]) GroupBy(groups ...Expression) *Selector[T] {
	s.group = groups
	return s
}

func (s *Selector[T]) Having(havings ...Predicate) *Selector[T] {
	s.having = havings
	return s
}

func (s *Selector[T]) OrderBy(orderbys ...Column) *Selector[T] {
	s.orderby = orderbys
	return s
}

func (s *Selector[T]) Size(size int64) *Selector[T] {
	s.limit = size
	return s
}

func (s *Selector[T]) Page(page int64) *Selector[T] {
	s.offset = (page - 1) * s.limit
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
