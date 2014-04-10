package inject

import "reflect"

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
