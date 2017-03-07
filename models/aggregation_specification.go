package models

type AggregationSpecification struct {
	Name                 string
	FilteredMetricName   string
	FilteredDimensions   map[string]string
	GroupedDimensions    []string
	AggregatedMetricName string
}
