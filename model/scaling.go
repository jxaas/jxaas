package model

type ScalingPolicy struct {
}

type Scaling struct {
	Policy        ScalingPolicy
	CurrentMetric float32
}
