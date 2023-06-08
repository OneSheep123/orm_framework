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
	"orm_framework/orm/internal/valuer"
	"testing"
	"time"
)

func TestSelector_Build(t *testing.T) {
	d := mysqlDB()
	db, _ := OpenDB(d)
	testCases := []struct {
		name    string
		builder QueryBuilder

		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "no from",
			builder: NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "from",
			builder: NewSelector[TestModel](db).From("`test`.`test_model`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test`.`test_model`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "where empty",
			builder: NewSelector[TestModel](db).Where(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name: "where",
			builder: NewSelector[TestModel](db).Where(
				C("FirstName").Eq("zhangsan").Or(C("LastName").Eq("list")),
				C("Age").Eq(12),
			),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` WHERE ((`first_name` = ?) OR (`last_name` = ?)) And (`age` = ?);",
				Args: []any{
					"zhangsan", "list", 12,
				},
			},
			wantErr: nil,
		},
		{
			name: "where err",
			builder: NewSelector[TestModel](db).Where(
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
		s    *Selector[TestModel]

		wantQuery *TestModel
		wantErr   error
	}{
		{
			name:    "invalid error",
			s:       NewSelector[TestModel](db).Where(C("XXX").Eq("12")),
			wantErr: errs.NewErrUnknownField("XXX"),
		},
		{
			name:    "query error",
			s:       NewSelector[TestModel](db).Where(C("Id").Eq("1")),
			wantErr: errors.New("query error"),
		},
		{
			name:    "no rows",
			s:       NewSelector[TestModel](db).Where(C("Id").Eq("1")),
			wantErr: ErrNoRows,
		},
		{
			name: "one rows",
			s:    NewSelector[TestModel](db).Where(C("Id").Eq("1")),
			wantQuery: &TestModel{
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

// 在 orm 目录下执行
// go test -bench=BenchmarkQuerier_Get -benchmem -benchtime=10000x
func BenchmarkQuerier_Get(b *testing.B) {
	sqlDB := mysqlDB()
	defer sqlDB.Close()
	db, err := OpenDB(sqlDB)
	if err != nil {
		b.Fatal(err)
	}
	_, err = db.db.Exec(TestModel{}.CreateSQL())
	if err != nil {
		b.Fatal(err)
	}

	res, err := db.db.Exec("INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`)"+
		"VALUES (?,?,?,?)", 12, "Deng", 18, "Ming")

	if err != nil {
		b.Fatal(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		b.Fatal(err)
	}
	if affected == 0 {
		b.Fatal()
	}

	b.Run("unsafe", func(b *testing.B) {
		db.Creator = valuer.NewUnsafeValue
		for i := 0; i < b.N; i++ {
			_, err = NewSelector[TestModel](db).Get(context.Background())
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("reflect", func(b *testing.B) {
		db.Creator = valuer.NewReflectValue
		for i := 0; i < b.N; i++ {
			_, err = NewSelector[TestModel](db).Get(context.Background())
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

type TestModel struct {
	Id        int
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

func (TestModel) CreateSQL() string {
	return `
CREATE TABLE IF NOT EXISTS test_model(
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    age INTEGER,
    last_name TEXT NOT NULL
)
`
}

func mysqlDB() *sql.DB {
	open, _ := sql.Open("mysql", "root:123123@tcp(127.0.0.1:3306)/test?charset=utf8mb4")
	// 设置最大连接数相关操作
	open.SetMaxOpenConns(100)
	open.SetMaxIdleConns(2)
	open.SetConnMaxIdleTime(1 * time.Millisecond)
	open.SetConnMaxLifetime(1 * time.Millisecond)
	return open
}
