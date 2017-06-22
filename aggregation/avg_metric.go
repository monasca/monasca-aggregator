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

type avgMetric struct {
	baseHolder
	count int64
}

func (a *avgMetric) InitValue(v models.MetricEnvelope) {
	a.envelope.Metric.Value = v.Metric.Value
	a.count = 1
}

func (a *avgMetric) UpdateValue(v models.MetricEnvelope) {
	a.envelope.Metric.Value += v.Metric.Value
	a.count++
}

func (a *avgMetric) GetMetric() models.MetricEnvelope {
	sum := a.envelope.Metric.Value
	a.envelope.Metric.Value = sum / float64(a.count)
	return a.baseHolder.GetMetric()
}
