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
		//{
		//	name:    "from",
		// 不支持自定义了
		//	builder: NewSelector[TestModel](db).From("`test`.`test_model`"),
		//	wantQuery: &Query{
		//		SQL:  "SELECT * FROM `test`.`test_model`;",
		//		Args: nil,
		//	},
		//	wantErr: nil,
		//},
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
				SQL: "SELECT * FROM `test_model` WHERE ((`first_name` = ?) OR (`last_name` = ?)) AND (`age` = ?);",
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

		{
			name:    "invalid column",
			builder: NewSelector[TestModel](db).Select(Avg("Invalid")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			name:    "partial columns",
			builder: NewSelector[TestModel](db).Select(C("Id"), C("FirstName")),
			wantQuery: &Query{
				SQL: "SELECT `id`,`first_name` FROM `test_model`;",
			},
		},
		{
			name:    "avg",
			builder: NewSelector[TestModel](db).Select(Avg("Age")),
			wantQuery: &Query{
				SQL: "SELECT AVG(`age`) FROM `test_model`;",
			},
		},
		{
			name:    "raw expression",
			builder: NewSelector[TestModel](db).Select(Raw("COUNT(DISTINCT `first_name`)")),
			wantQuery: &Query{
				SQL: "SELECT COUNT(DISTINCT `first_name`) FROM `test_model`;",
			},
		},
		{
			// 使用 RawExpr
			name: "raw expression",
			builder: NewSelector[TestModel](db).
				Where(Raw("`age` < ?", 18).AsPredicate()),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `age` < ?;",
				Args: []any{18},
			},
		},
		// 别名
		{
			name: "alias",
			builder: NewSelector[TestModel](db).
				Select(C("Id").As("my_id"),
					Avg("Age").As("avg_age")),
			wantQuery: &Query{
				SQL: "SELECT `id` AS `my_id`,AVG(`age`) AS `avg_age` FROM `test_model`;",
			},
		},
		// WHERE 忽略别名
		{
			name: "where ignore alias",
			builder: NewSelector[TestModel](db).
				Where(C("Id").As("my_id").LT(100)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` < ?;",
				Args: []any{100},
			},
		},
		{
			name:    "offset only",
			builder: NewSelector[TestModel](db).Offset(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` OFFSET ?;",
				Args: []any{10},
			},
		},
		{
			name:    "limit only",
			builder: NewSelector[TestModel](db).Limit(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` LIMIT ?;",
				Args: []any{10},
			},
		},
		{
			name:    "limit offset",
			builder: NewSelector[TestModel](db).Limit(20).Offset(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` LIMIT ? OFFSET ?;",
				Args: []any{20, 10},
			},
		},
		{
			// 调用了，但是啥也没传
			name:    "none",
			builder: NewSelector[TestModel](db).GroupBy(C("Age")).Having(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` GROUP BY `age`;",
			},
		},
		{
			// 单个条件
			name: "single",
			builder: NewSelector[TestModel](db).GroupBy(C("Age")).
				Having(C("FirstName").Eq("Deng")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` GROUP BY `age` HAVING `first_name` = ?;",
				Args: []any{"Deng"},
			},
		},
		{
			// 多个条件
			name: "multiple",
			builder: NewSelector[TestModel](db).GroupBy(C("Age")).
				Having(C("FirstName").Eq("Deng"), C("LastName").Eq("Ming")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` GROUP BY `age` HAVING (`first_name` = ?) AND (`last_name` = ?);",
				Args: []any{"Deng", "Ming"},
			},
		},
		{
			// 聚合函数
			name: "avg",
			builder: NewSelector[TestModel](db).GroupBy(C("Age")).
				Having(Avg("Age").Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` GROUP BY `age` HAVING AVG(`age`) = ?;",
				Args: []any{18},
			},
		},
		{
			// 调用了，但是啥也没传
			name:    "none group",
			builder: NewSelector[TestModel](db).GroupBy(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			// 单个
			name:    "single",
			builder: NewSelector[TestModel](db).GroupBy(C("Age")),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` GROUP BY `age`;",
			},
		},
		{
			// 多个
			name:    "multiple",
			builder: NewSelector[TestModel](db).GroupBy(C("Age"), C("FirstName")),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` GROUP BY `age`,`first_name`;",
			},
		},
		{
			// 不存在
			name:    "invalid column",
			builder: NewSelector[TestModel](db).GroupBy(C("Invalid")),
			wantErr: errs.NewErrUnknownField("Invalid"),
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

func TestSelector_Transaction(t *testing.T) {
	d := mysqlDB()
	db, _ := OpenDB(d)
	testCases := []struct {
		name    string
		builder QueryBuilder

		wantQuery *Query
		wantErr   error
	}{
		{
			name: "one",
			builder: func() QueryBuilder {
				tx, _ := db.BeginTx(context.Background(), &sql.TxOptions{})
				return NewSelector[TestModel](tx)
			}(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
			wantErr: nil,
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

func TestSelector_Join(t *testing.T) {
	sqlDB := mysqlDB()
	defer sqlDB.Close()
	db, err := OpenDB(sqlDB)
	require.NoError(t, err)
	type Order struct {
		Id        int
		UsingCol1 string
		UsingCol2 string
	}

	type OrderDetail struct {
		OrderId int
		ItemId  int

		UsingCol1 string
		UsingCol2 string
	}

	type Item struct {
		Id int
	}

	testCase := []struct {
		name      string
		s         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "specify table",
			s:    NewSelector[Order](db).From(TableOf(&OrderDetail{})),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `order_detail`;",
				Args: nil,
			},
		},
		{
			name: "using join",
			s: func() QueryBuilder {
				t1 := TableOf(&Order{})
				t2 := TableOf(&OrderDetail{})
				t3 := t1.Join(t2).Using("UsingCol1", "UsingCol2")
				return NewSelector[Order](db).From(t3)
			}(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM (`order` JOIN `order_detail` USING (`using_col1`,`using_col2`));",
				Args: nil,
			},
		},
		{
			name: "on join",
			s: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.Join(t2).On(t1.C("Id").Eq(t2.C("OrderId")))
				return NewSelector[Order](db).From(t3)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (`order` AS `t1` JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`);",
			},
		},
		{
			name: "join table",
			s: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.Join(t2).On(t1.C("Id").Eq(t2.C("OrderId")))
				t4 := TableOf(&Item{}).As("t4")
				t5 := t3.Join(t4).On(t4.C("Id").Eq(t2.C("ItemId")))
				return NewSelector[Order](db).From(t5)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM ((`order` AS `t1` JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`) " +
					"JOIN `item` AS `t4` ON `t4`.`id` = `t2`.`item_id`);",
			},
		},
		{
			name: "table join",
			s: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.Join(t2).On(t1.C("Id").Eq(t2.C("OrderId")))
				t4 := TableOf(&Item{}).As("t4")
				t5 := t4.Join(t3).On(t4.C("Id").Eq(t2.C("ItemId")))
				return NewSelector[Order](db).From(t5)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (`item` AS `t4` JOIN (`order` AS `t1` JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`) " +
					"ON `t4`.`id` = `t2`.`item_id`);",
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.s.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
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
	return open
}
