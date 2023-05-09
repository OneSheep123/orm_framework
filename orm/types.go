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

// Executor 执行角色，返回执行结果
type Executor interface {
	Exec(ctx context.Context) (sql.Result, error)
}

// QueryBuilder sql语句构建
type QueryBuilder interface {
	Build() (*Query, error)
}

type Query struct {
	SQL  string
	args []any
}
