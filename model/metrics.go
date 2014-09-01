package model

import "sort"

type Metrics struct {
	Metric []string
}

type MetricDataset struct {
	Points MetricDatapoints
}

type MetricDatapoints []MetricDatapoint

type MetricDatapoint struct {
	T int64
	V float32
}

func (self MetricDatapoints) Len() int {
	return len(self)
}

func (self MetricDatapoints) Less(i, j int) bool {
	return self[i].T < self[j].T
}

func (self MetricDatapoints) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self *MetricDataset) SortPointsByTime() {
	sort.Sort(self.Points)
}
