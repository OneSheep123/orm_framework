// Package unsafe create by chencanhua in 2023/5/31
package unsafe

import (
	"errors"
	"reflect"
	"unsafe"
)

type UnsafeAccessor struct {
	fields  map[string]FieldMeta
	address unsafe.Pointer
}

type FieldMeta struct {
	tOf reflect.Type
	// 字段偏移量
	offset uintptr
}

// NewUnsafeAccessor 根据entity初始化一个UnsafeAccessor操作
func NewUnsafeAccessor(entity any) *UnsafeAccessor {
	tOf := reflect.TypeOf(entity)
	// 使用一层指针
	tOf = tOf.Elem()
	numField := tOf.NumField()
	fields := make(map[string]FieldMeta, numField)
	for index := 0; index < numField; index++ {
		field := tOf.Field(index)
		fields[field.Name] = FieldMeta{
			tOf:    field.Type,
			offset: field.Offset,
		}
	}

	// 获取地址时，使用valueOf
	value := reflect.ValueOf(entity)
	return &UnsafeAccessor{
		// 值对应的起始地址
		address: value.UnsafePointer(),
		fields:  fields,
	}
}

func (a *UnsafeAccessor) Field(field string) (any, error) {
	meta, ok := a.fields[field]
	if !ok {
		return nil, errors.New("非法字段")
	}
	fdAddress := unsafe.Pointer(uintptr(a.address) + meta.offset)

	return reflect.NewAt(meta.tOf, fdAddress).Elem().Interface(), nil
}

func (a *UnsafeAccessor) SetField(field string, val any) error {
	meta, ok := a.fields[field]
	if !ok {
		return errors.New("非法字段")
	}
	fdAddress := unsafe.Pointer(uintptr(a.address) + meta.offset)
	reflect.NewAt(meta.tOf, fdAddress).Elem().Set(reflect.ValueOf(val))
	return nil
}
