package utils

import (
	"github.hpe.com/UNCLE/monasca-aggregation/models"
	log "github.com/Sirupsen/logrus"
)

func MatchMetric(metricSpec models.AggregationSpecification, actual models.Metric) bool {
	result := true
	if metricSpec.FilteredMetricName != "" && metricSpec.FilteredMetricName != actual.Name {
		log.Debugf("Missing name %s", metricSpec.Name)
		result = false
	}

	if metricSpec.FilteredDimensions != nil {
		if actual.Dimensions == nil {
			result = false
		}
		if !matchMap(metricSpec.FilteredDimensions, actual.Dimensions) {
			result = false
		}
	}

	if metricSpec.GroupedDimensions != nil {
		if actual.Dimensions == nil {
			result = false
		}
		if !matchMapKeys(metricSpec.GroupedDimensions, actual.Dimensions) {
			result = false
		}
	}

	return result
}

func matchMap(dimensionsSpec map[string]string, actual map[string]string) bool {
	outer:
	for s_key, s_value := range dimensionsSpec {
		for a_key, a_value := range actual {
			if s_key == a_key && s_value == a_value {
				continue outer
			}
		}
		log.Debugf("Missing dimension %s:%s", s_key, s_value)
		return false
	}
	return true
}

func matchMapKeys(keySpec []string, actual map[string]string) bool {
	outer:
	for _, s_key := range keySpec {
		for a_key := range actual {
			if s_key == a_key {
				continue outer
			}
		}
		log.Debugf("Missing key %s", s_key)
		return false
	}
	return true
}