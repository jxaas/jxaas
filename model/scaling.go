package model

type ScalingPolicy struct {
	Window *int

	MetricName *string

	// TODO: This is a pretty stupid policy; better to have a 'target' value and a model
	MetricMin *float32
	MetricMax *float32

	ScaleMin *int
	ScaleMax *int
}

type Scaling struct {
	Policy        ScalingPolicy
	MetricCurrent float32
	ScaleCurrent  int
	ScaleTarget   int
}
