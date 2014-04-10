package inject

type Binding interface {
	Get() (interface{}, error)
}
