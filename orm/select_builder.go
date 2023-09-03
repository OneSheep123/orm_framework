// create by chencanhua in 2023/9/3
package orm

type selectorBuilderAttribute struct {
	where   []Predicate
	having  []Predicate
	columns []Selectable
	groupBy []Column
	offset  int
	limit   int
}

type selectorBuilder struct {
	builder
	selectorBuilderAttribute
}
