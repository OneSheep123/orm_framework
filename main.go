// create by chencanhua in 2023/5/7
package main

import (
	"database/sql"
	"fmt"
	"reflect"
)

func main() {
	var u User
	t := reflect.TypeOf(u)
	fmt.Println(t.Name())
}

type User struct {
	Id        int
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
