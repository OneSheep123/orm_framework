// create by chencanhua in 2023/6/20
package orm

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"orm_framework/orm/internal/errs"
	"testing"
)

func TestInserter_Build(t *testing.T) {
	d := mysqlDB()
	db, _ := OpenDB(d)
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// 一个都不插入
			name:    "no value",
			q:       NewInserter[TestModel](db).Values(),
			wantErr: errs.ErrInsertZeroRow,
		},
		{
			name: "single values",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?);",
				Args: []any{1, "Deng", int8(18), &sql.NullString{String: "Ming", Valid: true}},
			},
		},
		{
			name: "multiple values",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				},
				&TestModel{
					Id:        2,
					FirstName: "Da",
					Age:       19,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?),(?,?,?,?);",
				Args: []any{1, "Deng", int8(18), &sql.NullString{String: "Ming", Valid: true},
					2, "Da", int8(19), &sql.NullString{String: "Ming", Valid: true}},
			},
		},
		{
			// 指定列
			name: "invalid columns",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).Columns("FirstName", "Invalid"),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			// 指定列
			name: "columns",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).Columns("FirstName", "Age"),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`first_name`,`age`) VALUES(?,?);",
				Args: []any{"Deng", int8(18)},
			},
		},
		{
			name: "upsert invalid column",
			q: NewInserter[TestModel](db).Values(&TestModel{
				Id:        1,
				FirstName: "Deng",
				Age:       18,
				LastName:  &sql.NullString{String: "Ming", Valid: true},
			}).OnDuplicateKey().Update(Assign("FirstName", "zhangsan"),
				Assign("xxx", 19)),
			wantErr: errs.NewErrUnknownField("xxx"),
		},
		{
			name: "upsert",
			q: NewInserter[TestModel](db).Values(&TestModel{
				Id:        1,
				FirstName: "Deng",
				Age:       18,
				LastName:  &sql.NullString{String: "Ming", Valid: true},
			}).OnDuplicateKey().Update(Assign("FirstName", "zhangsan"),
				Assign("LastName", 19)),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?) " +
					"ON DUPLICATE KEY UPDATE `first_name`=?,`last_name`=?;",
				Args: []any{1, "Deng", int8(18), &sql.NullString{String: "Ming", Valid: true}, "zhangsan", 19},
			},
		},
		{
			name: "upsert column",
			q: NewInserter[TestModel](db).Values(&TestModel{
				Id:        1,
				FirstName: "Deng",
				Age:       18,
				LastName:  &sql.NullString{String: "Ming", Valid: true},
			}).OnDuplicateKey().Update(C("FirstName")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?) " +
					"ON DUPLICATE KEY UPDATE `first_name`=VALUES(`first_name`);",
				Args: []any{1, "Deng", int8(18), &sql.NullString{String: "Ming", Valid: true}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}