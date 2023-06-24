// create by chencanhua in 2023/6/12
package orm

type Aggregate struct {
	fn    string
	arg   string
	alias string
	op    string
	val   any
}

// selectable 实现这个接口可以在Select后进行插入
func (a Aggregate) selectable() {}

func (a Aggregate) expr() {}

func Avg(c string) Aggregate {
	return Aggregate{
		fn:  "AVG",
		arg: c,
	}
}

func Max(c string) Aggregate {
	return Aggregate{
		fn:  "MAX",
		arg: c,
	}
}

func Min(c string) Aggregate {
	return Aggregate{
		fn:  "MIN",
		arg: c,
	}
}

func Count(c string) Aggregate {
	return Aggregate{
		fn:  "COUNT",
		arg: c,
	}
}

func Sum(c string) Aggregate {
	return Aggregate{
		fn:  "SUM",
		arg: c,
	}
}

func (a Aggregate) LT(val any) Aggregate {
	return Aggregate{
		fn:  a.fn,
		arg: a.arg,
		op:  opLT,
		val: val,
	}
}

func (a Aggregate) As(alias string) Aggregate {
	return Aggregate{
		fn:    "AVG",
		arg:   a.arg,
		alias: alias,
	}
}

func (a Aggregate) AsPredicate() Predicate {
	return Predicate{
		left: a,
	}
}
