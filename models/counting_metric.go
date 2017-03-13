package models

type countMetric struct {
	baseMetric
}

func (c *countMetric) InitValue(float64) {
	c.envelope.Metric.Value = 1
}

func (c *countMetric) UpdateValue(float64) {
	c.envelope.Metric.Value += 1
}

