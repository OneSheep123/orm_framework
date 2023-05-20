// Package orm create by chencanhua in 2023/5/14
package orm

type model struct {
	tableName string
	fieldMap  map[string]*field
}

type field struct {
	colName string
}

// 我们支持的全部标签上的 key 都放在这里
// 方便用户查找，和我们后期维护
const (
	tagKeyColumn = "column"
)

type TableName interface {
	TableName() string
}
