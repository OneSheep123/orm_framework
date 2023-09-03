// Package orm create by chencanhua in 2023/5/7
package orm

import (
	"context"
	"database/sql"
)

// Querier 不同语句的各自实现
// 使用泛型做类型的约束: 例如 SELECT 语句和 INSERT 语句
type Querier[T any] interface {
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) (*[]T, error)
}

// Executor 执行角色，返回执行结果(insert、update、delete)
type Executor interface {
	Exec(ctx context.Context) sql.Result
}

// QueryBuilder sql语句构建
type QueryBuilder interface {
	Build() (*Query, error)
}

// Session 代表一个抽象的概念，即会话
type Session interface {
	getCore() core
	queryContext(context context.Context, query string, args ...any) (*sql.Rows, error)
	execContext(context context.Context, query string, args ...any) (sql.Result, error)
}

type Query struct {
	SQL  string
	Args []any
}
