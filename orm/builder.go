// create by chencanhua in 2023/6/23
package orm

import (
	"orm_framework/orm/internal/errs"
	"strings"
)

type builder struct {
	core

	sb     strings.Builder
	args   []any
	quoter byte
}

func (b *builder) buildColumn(c *Column) error {
	switch table := c.table.(type) {
	case nil:
		field, ok := b.model.FieldMap[c.column]
		if !ok {
			return errs.NewErrUnknownField(c.column)
		}
		b.quote(field.ColName)
	case Table:
		m, err := b.r.Get(table.entity)
		if err != nil {
			return err
		}
		field, ok := m.FieldMap[c.column]
		if !ok {
			return errs.NewErrUnknownField(c.column)
		}
		if table.alias != "" {
			b.quote(table.alias)
			b.sb.WriteByte('.')
		}
		b.quote(field.ColName)
	default:
		return errs.NewErrUnsupportedTable(table)
	}

	return nil
}

func (b *builder) quote(column string) {
	b.sb.WriteByte(b.quoter)
	b.sb.WriteString(column)
	b.sb.WriteByte(b.quoter)
}

func (b *builder) addArgs(args ...any) {
	if b.args == nil {
		// 很少有查询能够超过八个参数
		// INSERT 除外
		b.args = make([]any, 0, 8)
	}
	b.args = append(b.args, args...)
}
