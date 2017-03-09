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

package tests

import (
	"testing"
	"github.hpe.com/UNCLE/monasca-aggregation/models"
	"github.hpe.com/UNCLE/monasca-aggregation/utils"
)

func TestMatchMetricSimpleName(t *testing.T) {
	metric := models.Metric{
		Name: "test_metric",
		Dimensions: map[string]string{
			"hostname": "test_host",
			"service": "test_service",
		},
		Value: 3,
		ValueMeta: map[string]string{
			"status": "0",
		},
	}

	//test simple filter name
	agg_spec := models.AggregationSpecification{
		Name: "test_metric_01",
		FilteredMetricName: "test_metric",
	}

	result := utils.MatchMetric(agg_spec, metric)
	if result != true {
		t.Fail()
	}
}

func TestMatchMetricNameAndDimensions(t *testing.T) {
	metric := models.Metric{
		Name: "test_metric",
		Dimensions: map[string]string{
			"hostname": "test_host",
			"service": "test_service",
		},
		Value: 3,
		ValueMeta: map[string]string{
			"status": "0",
		},
	}

	//test simple filter name
	agg_spec := models.AggregationSpecification{
		Name: "test_metric_01",
		FilteredMetricName: "test_metric",
		FilteredDimensions: map[string]string{
			"hostname": "test_host",
		},
	}

	result := utils.MatchMetric(agg_spec, metric)
	if result != true {
		t.Fail()
	}
}

func TestMatchMetricNameAndMultipleDimensions(t *testing.T) {
	metric := models.Metric{
		Name: "test_metric",
		Dimensions: map[string]string{
			"hostname": "test_host",
			"service": "test_service",
		},
		Value: 3,
		ValueMeta: map[string]string{
			"status": "0",
		},
	}

	//test simple filter name
	agg_spec := models.AggregationSpecification{
		Name: "test_metric_01",
		FilteredMetricName: "test_metric",
		FilteredDimensions: map[string]string{
			"hostname": "test_host",
			"service": "test_service",
		},
	}

	result := utils.MatchMetric(agg_spec, metric)
	if result != true {
		t.Fail()
	}
}


func TestMatchMetricSimpleNameInvalidName(t *testing.T) {
	metric := models.Metric{
		Name: "not_test_metric",
		Dimensions: map[string]string{
			"hostname": "test_host",
			"service": "test_service",
		},
		Value: 3,
		ValueMeta: map[string]string{
			"status": "0",
		},
	}

	//test simple filter name
	agg_spec := models.AggregationSpecification{
		Name: "test_metric_01",
		FilteredMetricName: "test_metric",
	}

	result := utils.MatchMetric(agg_spec, metric)
	if result != false {
		t.Fail()
	}
}

func TestMatchMetricNameAndDimensionsInvalidDimension(t *testing.T) {
	metric := models.Metric{
		Name: "test_metric",
		Dimensions: map[string]string{
			"hostname": "not_test_host",
			"service": "test_service",
		},
		Value: 3,
		ValueMeta: map[string]string{
			"status": "0",
		},
	}

	//test simple filter name
	agg_spec := models.AggregationSpecification{
		Name: "test_metric_01",
		FilteredMetricName: "test_metric",
		FilteredDimensions: map[string]string{
			"hostname": "test_host",
		},
	}

	result := utils.MatchMetric(agg_spec, metric)
	if result != false {
		t.Fail()
	}
}

func TestMatchMetricNameAndMultipleDimensionsInvalidDimension(t *testing.T) {
	metric := models.Metric{
		Name: "test_metric",
		Dimensions: map[string]string{
			"hostname": "test_host",
			"service": "not_test_service",
		},
		Value: 3,
		ValueMeta: map[string]string{
			"status": "0",
		},
	}

	//test simple filter name
	agg_spec := models.AggregationSpecification{
		Name: "test_metric_01",
		FilteredMetricName: "test_metric",
		FilteredDimensions: map[string]string{
			"hostname": "test_host",
			"service": "test_service",
		},
	}

	result := utils.MatchMetric(agg_spec, metric)
	if result != false {
		t.Fail()
	}
}