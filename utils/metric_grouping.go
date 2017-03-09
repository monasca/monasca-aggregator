package utils

import "github.hpe.com/UNCLE/monasca-aggregation/models"

func GetGroupedAggregationKey(metric models.Metric, aggSpec models.AggregationSpecification) string {
	newKey := aggSpec.AggregatedMetricName

	// make the key unique for the supplied groupings
	if aggSpec.GroupedDimensions != nil {
		for _, key := range aggSpec.GroupedDimensions {
			newKey += "," + key + ":" + metric.Dimensions[key]
		}
	}
	return newKey
}

func CreateNewAggregatedMetric(metric models.Metric, aggSpec models.AggregationSpecification) models.Metric {
	currentMetric := models.Metric{}

	currentMetric.Name = aggSpec.AggregatedMetricName
	currentMetric.Dimensions = aggSpec.FilteredDimensions

	if currentMetric.Dimensions == nil {
		currentMetric.Dimensions = map[string]string{}
	}
	for _, key := range aggSpec.GroupedDimensions {
		currentMetric.Dimensions[key] = metric.Dimensions[key]
	}
	currentMetric.Value = metric.Value

	return currentMetric
}
