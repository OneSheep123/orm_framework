// create by chencanhua in 2023/5/7
package orm

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"orm_framework/orm/internal/errs"
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
				SQL:  "SELECT * FROM `user`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "from",
			builder: (&Selector[User]{}).From("`test`.`user`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test`.`user`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "where empty",
			builder: (&Selector[User]{}).Where(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `user`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name: "where",
			builder: (&Selector[User]{}).Where(
				C("FirstName").Eq("zhangsan").Or(C("LastName").Eq("list")),
				C("Age").Eq(12),
			),
			wantQuery: &Query{
				SQL: "SELECT * FROM `user` WHERE ((`first_name` = ?) OR (`last_name` = ?)) And (`age` = ?);",
				Args: []any{
					"zhangsan", "list", 12,
				},
			},
			wantErr: nil,
		},
		{
			name: "where err",
			builder: (&Selector[User]{}).Where(
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

type User struct {
	Id        int
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
