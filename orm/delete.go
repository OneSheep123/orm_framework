// Package orm create by chencanhua in 2023/5/14
package orm

var _ QueryBuilder = &Deleter[any]{}

type Deleter[T any] struct {
	model *model
	table string
	builder
	where []Predicate
}

func (d *Deleter[T]) From(name string) *Deleter[T] {
	d.table = name
	return d
}

func (d *Deleter[T]) Where(pre ...Predicate) *Deleter[T] {
	d.where = pre
	return d
}

func (d *Deleter[T]) Build() (*Query, error) {
	m, err := parseModel(new(T))
	if err != nil {
		return nil, err
	}
	d.model = m
	d.sb.WriteString("DELETE FROM ")

	if d.table != "" {
		d.sb.WriteString(d.table)
	} else {
		d.sb.WriteByte('`')
		d.sb.WriteString(d.model.tableName)
		d.sb.WriteByte('`')
	}

	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")
		if err := d.buildPredicate(d.where, d.model); err != nil {
			return nil, err
		}
	}
	d.sb.WriteByte(';')
	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, nil
}
