package scheduler

type Runnable interface {
	Run() error
}
