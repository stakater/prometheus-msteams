/*
Copyright 2026 Stakater.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package adaptivecards

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGridColumnValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected GridColumnWidth
		hasError bool
	}{
		{"string 50px", "50px", GridColumnWidth{"50px"}, false},
		{"number 1", 1, GridColumnWidth{1}, false},
		{"number 1.5", 1.5, GridColumnWidth{}, true},
		{"invalid type", []int{1}, GridColumnWidth{""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var col GridColumnWidth
			// Properly marshal the input to JSON first
			jsonData, err := json.Marshal(tt.input)
			assert.NoError(t, err, "failed to marshal input")

			err = json.Unmarshal(jsonData, &col)
			if tt.hasError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, col)
		})
	}
}
