package inject

import (
	"fmt"
	"reflect"
)

type Binder struct {
	bindings map[reflect.Type]Binding
}

type Binding interface {
	Get() (interface{}, error)
}

type FunctionBinding struct {
	fn    interface{}
	valFn reflect.Value
}

func (self *FunctionBinding) Get() (interface{}, error) {
	// TODO: Inject arguments?
	in := []reflect.Value{}
	out := self.valFn.Call(in)

	var val interface{}
	if len(out) >= 1 {
		if !out[0].IsNil() {
			val = out[0].Interface()
		}
	}

	var err error
	if len(out) >= 2 {
		if !out[1].IsNil() {
			err = out[1].Interface().(error)
		}
	}

	return val, err
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
