package model

type Metrics struct {
	Metric []string
}

type MetricDataset struct {
	Points []MetricDatapoint
}

type MetricDatapoint struct {
	T int64
	V float32
}
