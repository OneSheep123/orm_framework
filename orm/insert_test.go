// create by chencanhua in 2023/6/20
package orm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestUpsert_SQLite3_Upsert(t *testing.T) {
	// todo: 临时使用mysql的db进行验证sqlite语句的组装情况
	d := mysqlDB()
	db, _ := OpenDB(d, WithDialect(SQLLiteDialect))
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// upsert
			name: "upsert",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).OnDuplicateKey().ConflictColumns("Id").
				Update(Assign("FirstName", "Da")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?) " +
					"ON CONFLICT(`id`) DO UPDATE SET `first_name`=?;",
				Args: []any{1, "Deng", int8(18), &sql.NullString{String: "Ming", Valid: true}, "Da"},
			},
		},
		{
			// upsert invalid column
			name: "upsert invalid column",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).OnDuplicateKey().ConflictColumns("Id").
				Update(Assign("Invalid", "Da")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			// conflict invalid column
			name: "conflict invalid column",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).OnDuplicateKey().ConflictColumns("Invalid").
				Update(Assign("FirstName", "Da")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			// 使用原本插入的值
			name: "upsert use insert value",
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
				}).OnDuplicateKey().ConflictColumns("Id").
				Update(C("FirstName"), C("LastName")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?),(?,?,?,?) " +
					"ON CONFLICT(`id`) DO UPDATE SET `first_name`=excluded.`first_name`,`last_name`=excluded.`last_name`;",
				Args: []any{1, "Deng", int8(18), &sql.NullString{String: "Ming", Valid: true},
					2, "Da", int8(19), &sql.NullString{String: "Ming", Valid: true}},
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

func TestInserter_Exec(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
	require.NoError(t, err)
	testCases := []struct {
		name     string
		i        *Inserter[TestModel]
		wantErr  error
		affected int64
	}{
		{
			name: "query error",
			i: func() *Inserter[TestModel] {
				return NewInserter[TestModel](db).Values(&TestModel{}).
					Columns("Invalid")
			}(),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			name: "db error",
			i: func() *Inserter[TestModel] {
				mock.ExpectExec("INSERT INTO .*").
					WillReturnError(errors.New("db error"))
				return NewInserter[TestModel](db).Values(&TestModel{})
			}(),
			wantErr: errors.New("db error"),
		},
		{
			name: "exec",
			i: func() *Inserter[TestModel] {
				res := driver.RowsAffected(1)
				mock.ExpectExec("INSERT INTO .*").WillReturnResult(res)
				return NewInserter[TestModel](db).Values(&TestModel{})
			}(),
			affected: 1,
		},
	}
	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			res := ts.i.Exec(context.Background())
			affected, err := res.RowsAffected()
			assert.Equal(t, ts.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, ts.affected, affected)
		})
	}
}
