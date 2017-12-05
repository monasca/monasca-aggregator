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
	"testing"
	"github.com/monasca/monasca-aggregator/models"
	"github.com/magiconair/properties/assert"
)

func TestAggregationRuleWithBadRollUpDim(t *testing.T) {
	rollUpGroupDim := []string{"service1"}
	groupDim := []string{"hostname", "service"}
	rollUpSpec := models.Rollup{"sum", rollUpGroupDim}
	spec := models.AggregationSpecification{"Aggregation", "count", "metric",
		map[string]string{"cluster": "test-cluster"},
		map[string]string{"hostname": "inactive-host"},
		groupDim, "aggregated-metric",
		rollUpSpec}
	_, err := NewAggregationRule(spec)
	assert.Equal(t, err.Error(), "Rule Aggregation must have rollup.groupedDimensions as a sub array of groupedDimensions")
}

func TestAggregationRuleWithBadDim(t *testing.T) {
	rollUpGroupDim := []string{"service"}
	groupDim := []string{}
	rollUpSpec := models.Rollup{"sum", rollUpGroupDim}
	spec := models.AggregationSpecification{"Aggregation", "count", "metric",
		map[string]string{"cluster": "test-cluster"},
		map[string]string{"hostname": "inactive-host"},
		groupDim, "aggregated-metric",
		rollUpSpec}
	_, err := NewAggregationRule(spec)
	assert.Equal(t, err.Error(), "Rule Aggregation must have rollup.groupedDimensions as a sub array of groupedDimensions")
}

func TestAggregationRuleWithGoodDim(t *testing.T) {
	rollUpGroupDim := []string{"service"}
	groupDim := []string{"hostname", "service"}
	rollUpSpec := models.Rollup{"sum", rollUpGroupDim}
	spec := models.AggregationSpecification{"Aggregation", "count", "metric",
		map[string]string{"cluster": "test-cluster"},
		map[string]string{"hostname": "inactive-host"},
		groupDim, "aggregated-metric",
		rollUpSpec}
	rule, err := NewAggregationRule(spec)
	assert.Equal(t, err, nil)
	assert.Equal(t, rule.AggregationSpecification, spec)
}

func TestAggregationRuleWithNoMetricName(t *testing.T) {
	metricName := ""
	rollUpSpec := models.Rollup{"sum", []string{"service"}}
	spec := models.AggregationSpecification{"Aggregation", "count", "metric",
		map[string]string{"cluster": "test-cluster"},
		map[string]string{"hostname": "inactive-host"},
		[]string{"hostname", "service"}, metricName,
		rollUpSpec}
	_, err := NewAggregationRule(spec)
	assert.Equal(t, err.Error(), "Rule Aggregation must have an aggregated metric name")
}

func TestAggregationRuleWithNoFilteredMetricName(t *testing.T) {
	filteredMetricName := ""
	rollUpSpec := models.Rollup{"sum", []string{"service"}}
	spec := models.AggregationSpecification{"Aggregation", "count", filteredMetricName,
		map[string]string{"cluster": "test-cluster"},
		map[string]string{"hostname": "inactive-host"},
		[]string{"hostname", "service"}, "aggregated-metric",
		rollUpSpec}
	_, err := NewAggregationRule(spec)
	assert.Equal(t, err.Error(), "Rule Aggregation must have a filtered metric name")
}

func TestAggregationRuleWithNoFunction(t *testing.T) {
	function := ""
	rollUpSpec := models.Rollup{"sum", []string{"service"}}
	spec := models.AggregationSpecification{"Aggregation", function, "metric",
		map[string]string{"cluster": "test-cluster"},
		map[string]string{"hostname": "inactive-host"},
		[]string{"hostname", "service"}, "aggregated-metric",
		rollUpSpec}
	_, err := NewAggregationRule(spec)
	assert.Equal(t, err.Error(), "Rule Aggregation must have a function")
}