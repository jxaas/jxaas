package inject

import (
	"fmt"
	"reflect"
)

type Binder struct {
	bindings map[reflect.Type]Binding
}

func NewBinder() *Binder {
	self := &Binder{}
	self.bindings = make(map[reflect.Type]Binding)
	return self
}

func (self *Binder) AddProvider(fn interface{}) {
	binding := &FunctionBinding{}
	binding.fn = fn
	binding.valFn = reflect.ValueOf(fn)

	if binding.valFn.Type().Kind() != reflect.Func {
		panic("Binding to invalid provider kind")
	}

	numOut := binding.valFn.Type().NumOut()
	if numOut != 1 && numOut != 2 {
		panic("Invalid number of return values from function provider")
	}

	returnType := binding.valFn.Type().Out(0)
	self.bindings[returnType] = binding
}

func (self *Binder) AddSingleton(obj interface{}) {
	binding := &SingletonBinding{}
	binding.obj = obj
	t := reflect.TypeOf(obj)
	self.bindings[t] = binding
}

func (self *Binder) Get(t reflect.Type) (interface{}, error) {
	binding, found := self.bindings[t]
	if !found {
		return nil, fmt.Errorf("No binding for type %v", t)
	}
	return binding.Get()
}

func (self *Binder) CreateInjector() Injector {
	i := &BinderInjector{}
	i.binder = self
	return i
}
