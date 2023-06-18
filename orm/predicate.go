// Package orm create by chencanhua in 2023/5/8
package orm

type op string

const (
	opEq  = "="
	opNot = "NOT"
	opOr  = "OR"
	opAnd = "And"
)

/**
使用Builder模式构造ORM中构造复杂SQL
*/

type Predicate struct {
	left  Expression
	op    op
	right Expression
}

func (Predicate) expr() {}

func Not(right Predicate) Predicate {
	return Predicate{
		op:    opNot,
		right: right,
	}
}

func (left Predicate) Or(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opOr,
		right: right,
	}
}

func (left Predicate) And(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opAnd,
		right: right,
	}
}

type Column struct {
	column string
}

func (Column) expr() {}

func (Column) selectable() {}

// sub.C("name").Eq(12)
func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEq,
		right: Value{val: arg},
	}
}

func C(name string) Column {
	return Column{column: name}
}

type Value struct {
	val any
}

func (Value) expr() {}

// Expression 是一个标记接口，代表表达式
type Expression interface {
	expr()
}
