package models

type maxMetric struct {
	baseMetric
}

func (max *maxMetric) InitValue(v float64) {
	max.envelope.Metric.Value = v
}

func (max *maxMetric) UpdateValue(v float64) {
	if max.envelope.Metric.Value < v {
		max.envelope.Metric.Value = v
	}
}
