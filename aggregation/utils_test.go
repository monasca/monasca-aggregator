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
	"github.com/stretchr/testify/assert"
	"testing"
)
func TestCheckSubArray(t *testing.T) {
	array := []string{"hostname", "service", "podname", "path"}
	subArray1 := []string{"path"}
	assert.Equal(t, true, CheckSubArray(subArray1, array))
	subArray2 := []string{"path", "Test"}
	assert.Equal(t, false, CheckSubArray(subArray2, array))
	emptyArray := []string{}
	assert.Equal(t, true, CheckSubArray(emptyArray, array))
	assert.Equal(t, false, CheckSubArray(subArray1, emptyArray))
}
