package inject

type SingletonBinding struct {
	obj interface{}
}

func (self *SingletonBinding) Get() (interface{}, error) {
	return self.obj, nil
}
