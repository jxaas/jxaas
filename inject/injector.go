package inject

import (
	"reflect"
)

type Injector interface {
	Get(t reflect.Type) (interface{}, error)
	Inject(p interface{}) error
}

type BinderInjector struct {
	binder *Binder
}

func (self *BinderInjector) Get(t reflect.Type) (interface{}, error) {
	return self.binder.Get(t)
}

func (self *BinderInjector) Inject(p interface{}) error {
	pType := reflect.TypeOf(p)
	t := pType.Elem()
	v, err := self.binder.Get(t)
	if err != nil {
		return err
	}
	val := reflect.ValueOf(p)
	val.Elem().Set(reflect.ValueOf(v))
	return nil
}
