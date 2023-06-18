// Package orm create by chencanhua in 2023/6/18
package orm

type RawExpr struct {
	raw  string
	args []interface{}
}

func (r RawExpr) selectable() {}

func (r RawExpr) expr() {}

func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}

func Raw(raw string, args ...any) RawExpr {
	return RawExpr{
		raw:  raw,
		args: args,
	}
}
