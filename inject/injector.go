package inject

import "reflect"

type Injector interface {
	Get(t reflect.Type) (interface{}, error)
}

type BinderInjector struct {
	binder *Binder
}

func (self *BinderInjector) Get(t reflect.Type) (interface{}, error) {
	return self.binder.Get(t)
}
