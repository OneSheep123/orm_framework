// create by chencanhua in 2023/6/1
package unsafe

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewUnsafeAccessor(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}
	u := &User{Name: "Tom", Age: 18}
	accessor := NewUnsafeAccessor(u)
	val, err := accessor.Field("Age")
	require.NoError(t, err)
	assert.Equal(t, 18, val)

	val, err = accessor.Field("Name")
	assert.Equal(t, "Tom", val)

	err = accessor.SetField("Age", 19)
	require.NoError(t, err)
}
