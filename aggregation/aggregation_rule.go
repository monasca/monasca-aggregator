// Copyright 2017 Hewlett Packard Enterprise Development LP
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package aggregation

import (
	log "github.com/Sirupsen/logrus"

	"github.hpe.com/UNCLE/monasca-aggregation/models"
	"time"
)

type AggregationRule struct {
	models.AggregationSpecification
	MetricCache
}


func NewAggregationRule(aggSpec models.AggregationSpecification) AggregationRule {
	return AggregationRule{
		AggregationSpecification: aggSpec,
		MetricCache: NewMetricCache(),
	}
}

func (a * AggregationRule) AddMetric(metricEnvelope models.MetricEnvelope, windowSize time.Duration) {
	eventTime := int64(metricEnvelope.Metric.Timestamp / float64(1000 * int64(windowSize.Seconds())))

	_, exists := a.Windows[eventTime]
	if !exists {
		a.Windows[eventTime] = NewWindow()
	}

	aggregationKey := metricEnvelope.Meta["tenantId"]

	// make the key unique for the supplied groupings
	if a.GroupedDimensions != nil {
		for _, key := range a.GroupedDimensions {
			aggregationKey += "," + key + ":" + metricEnvelope.Metric.Dimensions[key]
		}
	}
	log.Infof("Storing key %s at %d", aggregationKey, eventTime)

	currentMetric, exists := a.Windows[eventTime][aggregationKey]
	// create a new metric if one did not exist
	if !exists {
		//TODO change create metric to handle new aggregation rule type
		currentMetric = CreateMetricType(a.AggregationSpecification, metricEnvelope)
		currentMetric.SetTimestamp(float64(eventTime * 1000 * int64(windowSize.Seconds())))
	} else {
		currentMetric.UpdateValue(metricEnvelope.Metric.Value)
	}
	a.Windows[eventTime][aggregationKey] = currentMetric
}

func (a *AggregationRule) GetMetrics(eventTime int64) []models.MetricEnvelope {
	var metricsList []models.MetricEnvelope

	//TODO check for metrics existing before doing extra work
	window := a.Windows[eventTime]

	if a.Rollup.Function == "" {
		metricsList = make([]models.MetricEnvelope, len(window))
		i := 0
		for _, metricHolder := range window {
			metricsList[i] = metricHolder.GetMetric()
			i++
		}
	} else {
		log.Debugf("Executing rollup: %v", a.Rollup.Function)

		rollupMetricMap := make(map[string]MetricHolder)

		for _, metricHolder := range window {
			metricEnv := metricHolder.GetMetric()

			rollupKey := metricEnv.Meta["tenantId"]
			if a.Rollup.GroupedDimensions != nil {
				for _, key := range a.Rollup.GroupedDimensions {
					rollupKey += "," + key + ":" + metricEnv.Metric.Dimensions[key]
				}
			}

			rollupMetric, exists := rollupMetricMap[rollupKey]
			if !exists {
				tempAgg := a.AggregationSpecification
				tempAgg.Function = a.Rollup.Function
				tempAgg.GroupedDimensions = a.Rollup.GroupedDimensions

				newRollup := CreateMetricType(tempAgg, metricEnv)
				log.Debugf("Created new rollup: %v", newRollup)
				newRollup.SetTimestamp(metricEnv.Metric.Timestamp)

				rollupMetric = newRollup
				rollupMetricMap[rollupKey] = rollupMetric
			} else {
				rollupMetric.UpdateValue(metricEnv.Metric.Value)
			}
		}

		//convert to list
		metricsList = make([]models.MetricEnvelope, len(rollupMetricMap))
		i := 0
		for _, metric := range rollupMetricMap {
			metricsList[i] = metric.GetMetric()
			i++
		}
	}

	log.Debugf("Emitting Metrics: %v", metricsList)

	return metricsList
}

func (a *AggregationRule) MatchesMetric(newMetric models.MetricEnvelope) bool {
	result := true
	if a.FilteredMetricName != "" && a.FilteredMetricName != newMetric.Metric.Name {
		log.Debugf("Missing name %s", a.FilteredMetricName)
		result = false
	}

	if a.FilteredDimensions != nil {
		if newMetric.Metric.Dimensions == nil {
			result = false
		}
		if !matchDimensions(a.FilteredDimensions, newMetric.Metric.Dimensions) {
			result = false
		}
	}

	if a.GroupedDimensions != nil {
		if newMetric.Metric.Dimensions == nil {
			result = false
		}
		if !matchDimensionKeys(a.GroupedDimensions, newMetric.Metric.Dimensions) {
			result = false
		}
	}

	return result
}

func matchDimensions(dimensionsSpec map[string]string, actual map[string]string) bool {
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

func matchDimensionKeys(keySpec []string, actual map[string]string) bool {
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
