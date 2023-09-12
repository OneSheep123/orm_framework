// create by chencanhua in 2023/9/12
package orm

type inserterBuilderAttribute struct {
	columns []string
}

type insertBuilder struct {
	inserterBuilderAttribute
	builder
	// 使用一个 OnDuplicate 结构体，从而允许将来扩展更加复杂的行为
	onDuplicate *Upsert
}

// UpsertBuilder
// Inserter变为UpsertBuilder最后变为Inserter
// 其中UpsertBuilder去构建冲突部分
type UpsertBuilder[T any] struct {
	i *Inserter[T]
	// conflictColumns 这里只是作为临时变量存储，后续会赋值给Upsert内的conflictColumns
	conflictColumns []string
}

type Upsert struct {
	assigns         []Assignable
	conflictColumns []string
}

func (o *UpsertBuilder[T]) ConflictColumns(conflictColumns ...string) *UpsertBuilder[T] {
	o.conflictColumns = conflictColumns
	return o
}

// Update 也可以看做是一个终结方法，重新回到 Inserter 里面
func (o *UpsertBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicate = &Upsert{
		assigns:         assigns,
		conflictColumns: o.conflictColumns,
	}
	return o.i
}
