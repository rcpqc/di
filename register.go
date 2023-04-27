package di

import (
	"reflect"
)

var container = map[string]reflect.Type{}

// Register 注册
func Register(kind string, object any) { container[kind] = reflect.TypeOf(object) }
