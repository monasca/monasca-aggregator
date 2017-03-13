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

type minMetric struct {
	baseHolder
}

func (min *minMetric) InitValue(v float64) {
	min.envelope.Metric.Value = v
}

func (min *minMetric) UpdateValue(v float64) {
	if min.envelope.Metric.Value > v {
		min.envelope.Metric.Value = v
	}
}