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

import "github.com/monasca/monasca-aggregator/models"

type rateMetric struct {
	baseHolder
	startTime  float64
	finalValue float64
	finalTime  float64
}

func (r *rateMetric) InitValue(v models.MetricEnvelope) {
	r.envelope.Metric.Value = v.Metric.Value
	r.startTime = v.Metric.Timestamp
}

func (r *rateMetric) UpdateValue(v models.MetricEnvelope) {
	r.finalValue = v.Metric.Value
	r.finalTime = v.Metric.Timestamp
}

func (r *rateMetric) GetMetric() models.MetricEnvelope {
	deltaValue := r.finalValue - r.envelope.Metric.Value
	deltaTime := (r.finalTime - r.startTime) / 1000.0
	ratePerSec := deltaValue / deltaTime

	r.envelope.Metric.Value = ratePerSec
	return r.baseHolder.GetMetric()
}
