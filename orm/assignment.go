// create by chencanhua in 2023/6/23
package orm

type Assignable interface {
	assign()
}

type Assignment struct {
	column string
	val    any
}

func Assign(column string, val any) Assignable {
	return Assignment{
		column: column,
		val:    val,
	}
}

func (Assignment) assign() {}
