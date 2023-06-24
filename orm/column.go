// create by chencanhua in 2023/6/18
package orm

type sortType string

const (
	ASC  sortType = "ASC"
	DESC sortType = "DESC"
)

type Column struct {
	column string
	alias  string
	sort   sortType
}

func (Column) expr() {}

func (Column) selectable() {}

func (Column) assign() {}

// sub.C("name").Eq(12)
func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEQ,
		right: exprOf(arg),
	}
}

func (c Column) LT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLT,
		right: exprOf(arg),
	}
}

func (c Column) GT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opGT,
		right: exprOf(arg),
	}
}

func exprOf(arg any) Expression {
	switch exp := arg.(type) {
	case Expression:
		return exp
	default:
		return Value{val: arg}
	}
}

func C(name string) Column {
	return Column{column: name}
}

func (c Column) ASC() Column {
	return Column{
		column: c.column,
		alias:  c.alias,
		sort:   ASC,
	}
}

func (c Column) DESC() Column {
	return Column{
		column: c.column,
		alias:  c.alias,
		sort:   DESC,
	}
}

func (c Column) As(alias string) Column {
	return Column{
		column: c.column,
		alias:  alias,
	}
}

type Value struct {
	val any
}

func (Value) expr() {}
