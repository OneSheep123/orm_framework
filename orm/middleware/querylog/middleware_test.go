package querylog

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"orm_framework/orm"
	"testing"
)

func mysqlDB() *sql.DB {
	open, _ := sql.Open("mysql", "root:123123@tcp(127.0.0.1:3306)/test?charset=utf8mb4")
	// 设置最大连接数相关操作
	open.SetMaxOpenConns(100)
	open.SetMaxIdleConns(2)
	return open
}

func TestNewMiddlewareBuilder(t *testing.T) {
	var query string
	var args []any
	m := (&MiddlewareBuilder{}).LogFunc(func(q string, as []any) {
		query = q
		args = as
	})

	sqlDB := mysqlDB()
	defer sqlDB.Close()
	db, err := orm.OpenDB(sqlDB, orm.WithMiddleWare(m.Build()))
	require.NoError(t, err)
	_, _ = orm.NewSelector[TestModel](db).Where(orm.C("Id").Eq(10)).Get(context.Background())
	assert.Equal(t, "SELECT * FROM `test_model` WHERE `id` = ?;", query)
	assert.Equal(t, []any{10}, args)

	orm.NewInserter[TestModel](db).Values(&TestModel{Id: 18}).Exec(context.Background())
	assert.Equal(t, "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?);", query)
	assert.Equal(t, []any{int64(18), "", int8(0), (*sql.NullString)(nil)}, args)
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
