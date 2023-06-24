// create by chencanhua in 2023/6/23
package orm

import "orm_framework/orm/internal/errs"

type Dialect interface {
	quoter() byte
	buildOnUpsert(b *builder, odk *Upsert) error
}

var (
	MySQLDialect   Dialect = &mysqlDialect{}
	SQLLiteDialect Dialect = &sqlite3Dialect{}
)

type standardSQL struct {
}

func (s *standardSQL) quoter() byte {
	//TODO implement me
	panic("implement me")
}

func (s *standardSQL) buildOnUpsert() error {
	//TODO implement me
	panic("implement me")
}

type mysqlDialect struct {
	standardSQL
}

func (s *mysqlDialect) quoter() byte {
	return '`'
}

func (s *mysqlDialect) buildOnUpsert(b *builder, odk *Upsert) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	var err error
	for index, a := range odk.assigns {
		if index > 0 {
			b.sb.WriteByte(',')
		}
		switch assign := a.(type) {
		case Assignment:
			err = b.buildColumn(assign.column)
			if err != nil {
				return err
			}
			b.sb.WriteString(`=?`)
			b.addArgs(assign.val)
		case Column:
			err = b.buildColumn(assign.column)
			if err != nil {
				return err
			}
			b.sb.WriteString("=VALUES(")
			b.buildColumn(assign.column)
			b.sb.WriteString(")")
		}
	}
	return nil
}

type sqlite3Dialect struct {
	standardSQL
}

func (s *sqlite3Dialect) quoter() byte {
	return '`'
}

func (s *sqlite3Dialect) buildOnUpsert(b *builder, odk *Upsert) error {
	b.sb.WriteString(" ON CONFLICT")
	if len(odk.conflictColumns) > 0 {
		b.sb.WriteByte('(')
		for i, col := range odk.conflictColumns {
			if i > 0 {
				b.sb.WriteByte(',')
			}
			err := b.buildColumn(col)
			if err != nil {
				return err
			}
		}
		b.sb.WriteByte(')')
	}
	b.sb.WriteString(" DO UPDATE SET ")

	for idx, a := range odk.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch assign := a.(type) {
		case Column:
			fd, ok := b.model.FieldMap[assign.column]
			if !ok {
				return errs.NewErrUnknownField(assign.column)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=excluded.")
			b.quote(fd.ColName)
		case Assignment:
			err := b.buildColumn(assign.column)
			if err != nil {
				return err
			}
			b.sb.WriteString("=?")
			b.addArgs(assign.val)
		default:
			return errs.NewErrUnsupportedAssignableType(a)
		}
	}
	return nil
}
