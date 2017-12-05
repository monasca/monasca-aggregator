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
	"github.com/monasca/monasca-aggregator/models"
	"time"
	"fmt"
)

type Rule struct {
	models.AggregationSpecification
	MetricCache
}

func NewAggregationRule(aggSpec models.AggregationSpecification) (Rule, error) {
	if aggSpec.AggregatedMetricName == "" {
		return Rule{}, fmt.Errorf("Rule %s must have an aggregated metric name", aggSpec.Name)
	}
	if aggSpec.FilteredMetricName == "" {
		return Rule{}, fmt.Errorf("Rule %s must have a filtered metric name", aggSpec.Name)
	}
	if aggSpec.Function == "" {
		return Rule{}, fmt.Errorf("Rule %s must have a function", aggSpec.Name)
	}
	//check when rollup.groupedDimensions is not a subset of groupedDimensions
	if !CheckSubArray(aggSpec.Rollup.GroupedDimensions, aggSpec.GroupedDimensions) {
		return Rule{}, fmt.Errorf("Rule %s must have rollup.groupedDimensions as a sub array of groupedDimensions", aggSpec.Name)
	}
	return Rule{
		AggregationSpecification: aggSpec,
		MetricCache:              NewMetricCache(),
	}, nil
}

func (a *Rule) AddMetric(metricEnvelope models.MetricEnvelope, windowSize time.Duration) {
	log.Debugf("Adding metric to %s", a.Name)
	eventTime := int64(metricEnvelope.Metric.Timestamp / float64(1000*int64(windowSize.Seconds())))

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
	log.Debugf("Storing key %s at %d", aggregationKey, eventTime)

	currentMetric, exists := a.Windows[eventTime][aggregationKey]
	// create a new metric if one did not exist
	if !exists {
		//TODO change create metric to handle new aggregation rule type
		currentMetric = CreateMetricType(a.AggregationSpecification, metricEnvelope)
		currentMetric.SetTimestamp(float64(eventTime * 1000 * int64(windowSize.Seconds())))
	} else {
		currentMetric.UpdateValue(metricEnvelope)
	}
	a.Windows[eventTime][aggregationKey] = currentMetric
}

func (a *Rule) GetMetrics(eventTime int64) []models.MetricEnvelope {
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

			log.Debugf("RollupKey: %s", rollupKey)
			rollupMetric, exists := rollupMetricMap[rollupKey]
			if !exists {
				tempAgg := a.AggregationSpecification
				tempAgg.Function = a.Rollup.Function
				tempAgg.GroupedDimensions = a.Rollup.GroupedDimensions

				newRollup := CreateMetricType(tempAgg, metricEnv)
				log.Debugf("Created new rollup: %v", newRollup)
				newRollup.SetTimestamp(metricEnv.Metric.Timestamp)

				rollupMetric = newRollup
			} else {
				log.Debugf("Updating rollup: %s", rollupKey)
				rollupMetric.UpdateValue(metricEnv)
			}
			rollupMetricMap[rollupKey] = rollupMetric
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

func (a *Rule) MatchesMetric(newMetric models.MetricEnvelope) bool {
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

	if a.RejectedDimensions != nil {
		if newMetric.Metric.Dimensions == nil {
			result = false
		}
		if rejectDimensions(a.RejectedDimensions, newMetric.Metric.Dimensions) {
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
	for sKey, sValue := range dimensionsSpec {
		for aKey, aValue := range actual {
			if sKey == aKey && sValue == aValue {
				continue outer
			}
		}
		log.Debugf("Missing dimension %s:%s", sKey, sValue)
		return false
	}
	return true
}

func rejectDimensions(dimensionsSpec map[string]string, actual map[string]string) bool {
	for sKey, sValue := range dimensionsSpec {
		for aKey, aValue := range actual {
			if sKey == aKey && (sValue == "" || sValue == aValue) {
				return true
			}
		}
	}
	return false
}

func matchDimensionKeys(keySpec []string, actual map[string]string) bool {
outer:
	for _, sKey := range keySpec {
		for aKey := range actual {
			if sKey == aKey {
				continue outer
			}
		}
		log.Debugf("Missing key %s", sKey)
		return false
	}
	return true
}
