package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_geb(t *testing.T) {
	buffer := &bytes.Buffer{}
	err := gen(buffer, "testdata/user.go")
	require.NoError(t, err)
	assert.Equal(t, `package testdata

import (
    "gitee.com/geektime-geekbang/geektime-go/orm"
    
    sqlx "database/sql"
    
)`, buffer.String())
}

func Test_genFile(t *testing.T) {
	f, err := os.Create("testdata/user.gen.go")
	require.NoError(t, err)
	err = gen(f, "testdata/user.go")
	require.NoError(t, err)
}
