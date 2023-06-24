// Package orm create by chencanhua in 2023/6/18
package orm

type RawExpr struct {
	raw  string
	args []interface{}
}

// selectable 实现这个接口可以在Select后进行插入
func (r RawExpr) selectable() {}

// expr 实现这个接口可以在Where后进行插入
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
