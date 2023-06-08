// create by chencanhua in 2023/6/8
package valuer

import "testing"

func TestNewReflectValue(t *testing.T) {
	testSetColumns(t, NewReflectValue)
}
