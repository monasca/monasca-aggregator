package models

type minMetric struct {
	baseMetric
}

func (min *minMetric) InitValue(v float64) {
	min.envelope.Metric.Value = v
}

func (min *minMetric) UpdateValue(v float64) {
	if min.envelope.Metric.Value > v {
		min.envelope.Metric.Value = v
	}
}