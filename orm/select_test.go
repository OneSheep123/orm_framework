// create by chencanhua in 2023/5/7
package orm

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"orm_framework/orm/internal/errs"
	"testing"
)

func TestSelector_Build(t *testing.T) {
	d := mysqlDB(t)
	db, _ := OpenDB(d)
	testCases := []struct {
		name    string
		builder QueryBuilder

		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "no from",
			builder: NewSelector[User](db),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `user`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "from",
			builder: NewSelector[User](db).From("`test`.`user`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test`.`user`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "where empty",
			builder: NewSelector[User](db).Where(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `user`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name: "where",
			builder: NewSelector[User](db).Where(
				C("FirstName").Eq("zhangsan").Or(C("LastName").Eq("list")),
				C("Age").Eq(12),
			),
			wantQuery: &Query{
				SQL: "SELECT * FROM `user` WHERE ((`first_name` = ?) OR (`last_name` = ?)) And (`age` = ?);",
				Args: []any{
					"zhangsan", "list", 12,
				},
			},
			wantErr: nil,
		},
		{
			name: "where err",
			builder: NewSelector[User](db).Where(
				C("FirstName").Eq("zhangsan").Or(C("XXX").Eq("list")),
				C("Age").Eq(12),
			),
			wantErr: errs.NewErrUnknownField("XXX"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}

func TestSelector_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	// 第一句不用加这个，因为不是SQL执行的错误
	//mock.ExpectQuery("SELECT .*").WillReturnError(errs.NewErrUnknownField("XXX"))

	// 2
	mock.ExpectQuery("SELECT .*").WillReturnError(errors.New("query error"))

	// 3
	rows := mock.NewRows([]string{"id", "first_name", "age", "last_name"})
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	// 4
	rows = mock.NewRows([]string{"id", "first_name", "age", "last_name"})
	rows.AddRow([]byte("1"), []byte("Da"), []byte("18"), []byte("Ming"))
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	db, err := OpenDB(mockDB)
	require.NoError(t, err)
	testCases := []struct {
		name string
		s    *Selector[User]

		wantQuery *User
		wantErr   error
	}{
		{
			name:    "invalid error",
			s:       NewSelector[User](db).Where(C("XXX").Eq("12")),
			wantErr: errs.NewErrUnknownField("XXX"),
		},
		{
			name:    "query error",
			s:       NewSelector[User](db).Where(C("Id").Eq("1")),
			wantErr: errors.New("query error"),
		},
		{
			name:    "no rows",
			s:       NewSelector[User](db).Where(C("Id").Eq("1")),
			wantErr: ErrNoRows,
		},
		{
			name: "one rows",
			s:    NewSelector[User](db).Where(C("Id").Eq("1")),
			wantQuery: &User{
				Id:        1,
				FirstName: "Da",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "Ming"},
			},
		},
	}

	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			res, err := ts.s.Get(context.Background())
			assert.Equal(t, ts.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, ts.wantQuery, res)
		})
	}
}

type User struct {
	Id        int
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

func mysqlDB(t *testing.T) *sql.DB {
	open, err := sql.Open("mysql", "root:123123@tcp(127.0.0.1:3306)/test?charset=utf8mb4")
	require.NoError(t, err)
	return open
}
