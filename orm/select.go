// Package orm create by chencanhua in 2023/5/7
package orm

import (
	"context"
	"orm_framework/orm/internal/errs"
	"reflect"
	"strings"
)

var _ QueryBuilder = &Selector[any]{}

// Selector 用于构建Select语句
type Selector[T any] struct {
	model *Model
	db    *DB
	table string
	where []Predicate
	args  []any
	sb    strings.Builder
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
	s.sb.WriteString("SELECT * FROM ")

	if s.table == "" {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.tableName)
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
	sql, err := s.Build()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.db.QueryContext(ctx, sql.SQL, sql.Args...)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		// 这里调用的是error下的ErrNoRows
		return nil, ErrNoRows
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	vals := make([]any, 0, len(columns))
	valsElems := make([]reflect.Value, 0, len(columns))
	for _, c := range columns {
		fieldInfo, ok := s.model.columnMap[c]
		if !ok {
			return nil, errs.NewErrUnknownColumn(c)
		}
		// 例如: fieldInfo.tOf = int, 那么这里value 是 *int
		value := reflect.New(fieldInfo.tOf)
		vals = append(vals, value.Interface())
		// 记得调用Elem，因为fieldInfo.tOf = int, 那么这里value 是 *int
		valsElems = append(valsElems, value.Elem())
	}

	// 注意这里是vals...
	rows.Scan(vals...)
	res := new(T)
	tRes := reflect.ValueOf(res)
	for index, c := range columns {
		fieldInfo, ok := s.model.columnMap[c]
		if !ok {
			return nil, errs.NewErrUnknownColumn(c)
		}
		tRes.Elem().FieldByName(fieldInfo.fieldName).
			Set(valsElems[index])
	}

	return res, nil
}

func (s *Selector[T]) GetMulti(ctx context.Context) (*[]T, error) {
	panic("implement me")
}
