package models

type avgMetric struct {
	baseMetric
	count      int64
}

func (a *avgMetric) InitValue(v float64) {
	a.envelope.Metric.Value = v
	a.count = 1
}

func (a *avgMetric) UpdateValue(v float64) {
	a.envelope.Metric.Value += v
	a.count++
}

func (a *avgMetric) GetMetric() MetricEnvelope {
	sum := a.envelope.Metric.Value
	a.envelope.Metric.Value = sum / float64(a.count)
	return a.baseMetric.GetMetric()
}