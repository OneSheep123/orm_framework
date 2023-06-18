// Package orm create by chencanhua in 2023/5/8
package orm

type op string

const (
	opEQ  = "="
	opLT  = "<"
	opGT  = ">"
	opAND = "AND"
	opOR  = "OR"
	opNOT = "NOT"
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
		op:    opNOT,
		right: right,
	}
}

func (left Predicate) Or(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opOR,
		right: right,
	}
}

func (left Predicate) And(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opAND,
		right: right,
	}
}

// Expression 是一个标记接口，代表表达式
type Expression interface {
	expr()
}
