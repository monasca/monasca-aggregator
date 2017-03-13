package models

type StorageMetric interface {
 	InitEnvelope(MetricEnvelope)
	InitValue(float64)
	UpdateValue(float64)
	GetMetric() MetricEnvelope
	SetTimestamp(float64)
}

type baseMetric struct {
	envelope MetricEnvelope
}

func (b *baseMetric) InitEnvelope(m MetricEnvelope) {
	b.envelope = m
}

func (b *baseMetric) GetMetric() MetricEnvelope {
	return b.envelope
}

func (b *baseMetric) SetTimestamp(t float64) {
	b.envelope.Metric.Timestamp = t
}


func CreateMetricType(aggSpec AggregationSpecification, metricEnv MetricEnvelope) StorageMetric {
	newMetricEnvelope := MetricEnvelope{}

	newMetricEnvelope.Metric.Name = aggSpec.AggregatedMetricName
	newMetricEnvelope.Metric.Dimensions = aggSpec.FilteredDimensions

	if newMetricEnvelope.Metric.Dimensions == nil {
		newMetricEnvelope.Metric.Dimensions = map[string]string{}
	}
	// get grouped dimension values
	for _, key := range aggSpec.GroupedDimensions {
		newMetricEnvelope.Metric.Dimensions[key] = metricEnv.Metric.Dimensions[key]
	}

	var metric StorageMetric
	switch aggSpec.Function {
	case "count":
		metric = new(countMetric)
	case "sum":
		metric = new(sumMetric)
	case "max":
		metric = new(maxMetric)
	case "min":
		metric = new(minMetric)
	case "avg":
		metric = new(avgMetric)
	}
	metric.InitEnvelope(newMetricEnvelope)
	metric.InitValue(metricEnv.Metric.Value)
	return metric
}
