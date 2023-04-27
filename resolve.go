package di

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// IResolvable 可解析
type IResolvable interface{ OnResolve() error }

// Resolver 解析器
type Resolver interface {
	Resolve(rv reflect.Value, node any) error
}

// Resolve 解析配置
func Resolve(bytes []byte, kindkey string) (any, error) {
	st := reflect.StructOf([]reflect.StructField{{Name: kindkey}})
	base := reflect.New(st)
	if err := yaml.Unmarshal(bytes, base.Interface()); err != nil {
		return nil, err
	}
	kind := base.Elem().Field(0).String()
	t, ok := container[kind]
	if !ok {
		return nil, fmt.Errorf("unregistered kind(%s)", kind)
	}
	rv := reflect.New(t).Elem()
	if err := resolve(rv, bytes); err != nil {
		return nil, fmt.Errorf("resolve err(%v)", err)
	}
	entity := rv.Interface()
	if ir, ok := entity.(IResolvable); ok {
		if err := ir.OnResolve(); err != nil {
			return nil, fmt.Errorf("resolve err(%v)", err)
		}
	}
	return entity, nil
}
