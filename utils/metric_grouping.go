package utils

import "github.hpe.com/UNCLE/monasca-aggregation/models"

func CreateNewAggregatedMetric(metric models.Metric, aggregation models.AggregationSpecification) models.Metric {
	currentMetric := models.Metric{}
	currentMetric.Name = aggregation.AggregatedMetricName
	currentMetric.Dimensions = aggregation.FilteredDimensions
	if currentMetric.Dimensions == nil {
		currentMetric.Dimensions = map[string]string{}
	}
	for _, key := range aggregation.GroupedDimensions {
		currentMetric.Dimensions[key] = metric.Dimensions[key]
	}
	currentMetric.Value = metric.Value
	return currentMetric
}
