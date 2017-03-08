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

package models

import (
	log "github.com/Sirupsen/logrus"
)

type Metric struct {
	Name       string            `json:"name"`
	Dimensions map[string]string `json:"dimensions"`
	Timestamp  float64           `json:"timestamp"`
	Value      float64           `json:"value"`
	ValueMeta  map[string]string `json:"value_meta"`
}


func MatchMetric(metricSpec AggregationSpecification, actual Metric) bool {
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