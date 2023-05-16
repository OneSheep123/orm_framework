// create by chencanhua in 2023/5/14
package orm

import (
	"orm_framework/orm/internal/errs"
	"strings"
)

type builder struct {
	args []any
	sb   strings.Builder
}

func (d *builder) buildPredicate(ps []Predicate, m *model) error {
	pre := ps[0]
	for index := 1; index < len(ps); index++ {
		pre = pre.And(ps[index])
	}
	if err := d.buildExpression(pre, m); err != nil {
		return err
	}
	return nil
}

func (d *builder) buildExpression(expression Expression, m *model) error {
	if expression == nil {
		return nil
	}
	switch expr := expression.(type) {
	case Column:
		c, ok := m.fieldMap[expr.column]
		if !ok {
			return errs.NewErrUnknownField(expr.column)
		}
		d.sb.WriteByte('`')
		d.sb.WriteString(c.colName)
		d.sb.WriteByte('`')
	case Value:
		d.sb.WriteByte('?')
		d.args = append(d.args, expr.val)
	case Predicate:
		_, lp := expr.left.(Predicate)
		if lp {
			d.sb.WriteByte('(')
		}
		if err := d.buildExpression(expr.left, m); err != nil {
			return err
		}
		if lp {
			d.sb.WriteByte(')')
		}
		d.sb.WriteByte(' ')
		d.sb.WriteString(string(expr.op))
		d.sb.WriteByte(' ')
		_, rp := expr.right.(Predicate)
		if rp {
			d.sb.WriteByte('(')
		}
		if err := d.buildExpression(expr.right, m); err != nil {
			return err
		}
		if rp {
			d.sb.WriteByte(')')
		}
	}
	return nil
}
