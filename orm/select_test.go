// create by chencanhua in 2023/5/7
package orm

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelector_Build(t *testing.T) {
	testCases := []struct {
		name    string
		builder QueryBuilder

		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "no from",
			builder: &Selector[User]{},
			wantQuery: &Query{
				SQL:  "SELECT * FROM `User`;",
				args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "from",
			builder: (&Selector[User]{}).From("`test`.`user`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test`.`user`;",
				args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "where empty",
			builder: (&Selector[User]{}).Where(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `User`;",
				args: nil,
			},
			wantErr: nil,
		},
		{
			name: "where",
			builder: (&Selector[User]{}).Where(
				C("firstName").Eq("zhangsan").Or(C("lastName").Eq("list")),
				C("age").Eq(12),
			),
			wantQuery: &Query{
				SQL: "SELECT * FROM `User` WHERE ((`firstName` = ?) OR (`lastName` = ?)) And (`age` = ?);",
				args: []any{
					"zhangsan", "list", 12,
				},
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

type User struct {
	Id        int
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
