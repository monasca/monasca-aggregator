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

type deltaMetric struct {
	baseHolder
	finalValue float64
}

func (d *deltaMetric) InitValue(v models.MetricEnvelope) {
	d.envelope.Metric.Value = v.Metric.Value
}

func (d *deltaMetric) UpdateValue(v models.MetricEnvelope) {
	d.finalValue = v.Metric.Value
}

func (d *deltaMetric) GetMetric() models.MetricEnvelope {
	deltaValue := d.finalValue - d.envelope.Metric.Value
	d.envelope.Metric.Value = deltaValue
	return d.baseHolder.GetMetric()
}
