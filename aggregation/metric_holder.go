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

import "github.hpe.com/UNCLE/monasca-aggregation/models"

type MetricHolder interface {
 	InitEnvelope(models.MetricEnvelope)
	InitValue(float64)
	UpdateValue(float64)
	GetMetric() models.MetricEnvelope
	SetTimestamp(float64)
}

type baseHolder struct {
	envelope models.MetricEnvelope
}

func (b *baseHolder) InitEnvelope(m models.MetricEnvelope) {
	b.envelope = m
}

func (b *baseHolder) GetMetric() models.MetricEnvelope {
	return b.envelope
}

func (b *baseHolder) SetTimestamp(t float64) {
	b.envelope.Metric.Timestamp = t
}


func CreateMetricType(aggSpec models.AggregationSpecification, metricEnv models.MetricEnvelope) MetricHolder {
	newMetricEnvelope := models.MetricEnvelope{}

	newMetricEnvelope.Metric.Name = aggSpec.AggregatedMetricName
	newMetricEnvelope.Metric.Dimensions = aggSpec.FilteredDimensions

	if newMetricEnvelope.Metric.Dimensions == nil {
		newMetricEnvelope.Metric.Dimensions = map[string]string{}
	}
	// get grouped dimension values
	for _, key := range aggSpec.GroupedDimensions {
		newMetricEnvelope.Metric.Dimensions[key] = metricEnv.Metric.Dimensions[key]
	}

	var metric MetricHolder
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
