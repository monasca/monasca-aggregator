package models

type sumMetric struct {
	baseMetric
}

func (s *sumMetric) InitValue(v float64) {
	s.envelope.Metric.Value = v
}

func (s *sumMetric) UpdateValue(v float64) {
	s.envelope.Metric.Value += v
}
