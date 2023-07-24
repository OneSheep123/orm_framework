// create by chencanhua in 2023/7/16
package orm

type TableReference interface {
	table()
}

type Table struct {
	entity any
	alias  string
}

func (Table) table() {}

func TableOf(entity any) Table {
	return Table{
		entity: entity,
	}
}

func (t Table) As(alias string) Table {
	return Table{
		entity: t.entity,
		alias:  alias,
	}
}

func (t Table) C(goColumn string) Column {
	return Column{
		column: goColumn,
		alias:  t.alias,
		table:  t,
	}
}

// t1 := TableOf("t1").Join(TableOf("t2"))
func (t Table) Join(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: right,
		typ:   "JOIN",
	}
}

func (t Table) LeftJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: right,
		typ:   "LEFT JOIN",
	}
}

func (t Table) RightJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: right,
		typ:   "RIGHT JOIN",
	}
}

type Join struct {
	left  TableReference
	right TableReference
	typ   string
	on    []Predicate
	using []string
}

// t3 := t1.Join(t2).On(C("Id").EQ("RefId"))
// t4 := t3.LeftJoin(t2)
func (j Join) Join(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: right,
		typ:   "JOIN",
	}
}

func (j Join) LeftJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: right,
		typ:   "LEFT JOIN",
	}
}

func (j Join) RightJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: right,
		typ:   "RIGHT JOIN",
	}
}

func (Join) table() {}

type JoinBuilder struct {
	left  TableReference
	right TableReference
	typ   string
}

// t3 := t1.Join(t2).On(C("Id").EQ("RefId"))
func (j *JoinBuilder) On(pd ...Predicate) Join {
	return Join{
		left:  j.left,
		right: j.right,
		typ:   j.typ,
		on:    pd,
	}
}

// t3 := t1.Join(t2).Using(col)
func (j *JoinBuilder) Using(cols ...string) Join {
	return Join{
		left:  j.left,
		right: j.right,
		typ:   j.typ,
		using: cols,
	}
}
