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

func TestVersionParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected Version
		hasError bool
	}{
		{"1.0", Version{1, 0}, false},
		{"1.2", Version{1, 2}, false},
		{"1.5", Version{1, 5}, false},
		{"2.0", Version{2, 0}, false},
		{"invalid", Version{}, true},
		{"1", Version{}, true},
		{"1.2.3", Version{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseVersion(tt.input)
			if tt.hasError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMustParseVersion(t *testing.T) {
	// Valid cases should not panic
	v := MustParseVersion("1.5")
	assert.Equal(t, 1, v.Major)
	assert.Equal(t, 5, v.Minor)

	// Invalid cases should panic
	assert.Panics(t, func() {
		MustParseVersion("invalid")
	})
}

func TestVersionComparison(t *testing.T) {
	v10, _ := ParseVersion("1.0")
	v12, _ := ParseVersion("1.2")
	v15, _ := ParseVersion("1.5")
	v20, _ := ParseVersion("2.0")

	tests := []struct {
		name     string
		v1       Version
		v2       Version
		expected int
	}{
		{"equal", v12, v12, 0},
		{"less minor", v10, v12, -1},
		{"greater minor", v15, v12, 1},
		{"less major", v15, v20, -1},
		{"greater major", v20, v10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.Compare(tt.v2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionSupport(t *testing.T) {
	v12, _ := ParseVersion("1.2")
	v15, _ := ParseVersion("1.5")

	assert.True(t, v15.SupportsVersion(v12), "version 1.5 should support 1.2")
	assert.False(t, v12.SupportsVersion(v15), "version 1.2 should not support 1.5")
}

func TestCardVersionValidation(t *testing.T) {
	tests := []struct {
		name        string
		card        AdaptiveCard
		expectError bool
		errorFields []string
	}{
		{
			name: "valid 1.0 card",
			card: AdaptiveCard{
				Version: "1.0",
				Body: []Element{
					TextBlock{
						Text: AsPtr("Hello World"),
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid 1.2 card with fallback",
			card: AdaptiveCard{
				Version: "1.2",
				Body: []Element{
					TextBlock{
						Text: AsPtr("Hello"),
						//Fallback: FallbackOptionDrop,
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid - 1.0 card with 1.2 features",
			card: AdaptiveCard{
				Version:   "1.0",
				MinHeight: "100px", // Requires 1.2
			},
			expectError: true,
			errorFields: []string{"MinHeight"},
		},
		{
			name: "invalid - 1.2 card with 1.5 features",
			card: AdaptiveCard{
				Version: "1.2",
				RTL:     AsPtr(true), // Requires 1.5
			},
			expectError: true,
			errorFields: []string{"RTL"},
		},
		{
			name: "valid 1.5 card with all features",
			card: AdaptiveCard{
				Version:   "1.5",
				MinHeight: "100px",     // 1.2 feature
				RTL:       AsPtr(true), // 1.5 feature
				Body: []Element{
					TextBlock{
						Text:  AsPtr("Hello"),
						Style: AsPtr(TextBlockStyleHeading), // 1.5 feature
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateVersion(tt.card, tt.card.Version)

			if tt.expectError {
				assert.NotEmpty(t, errors)

				var actualFields []string
				for _, err := range errors {
					actualFields = append(actualFields, err.FieldName)
				}

				assert.Subset(t, actualFields, tt.errorFields)
			} else {
				assert.Empty(t, errors)
			}
		})
	}
}

func TestCardSerialization(t *testing.T) {
	card := AdaptiveCard{
		Version: "1.5",
		Body: []Element{
			TextBlock{
				Text:   AsPtr("Hello World"),
				Weight: AsPtr(FontWeightBolder),
				Size:   AsPtr(FontSizeLarge),
			},
		},
		Actions: []Action{
			ActionOpenURL{
				CommonActionProperties: &CommonActionProperties{
					Title: "Learn More",
				},
				URL: "https://adaptivecards.io",
			},
		},
	}

	// Validate before marshaling
	errors := ValidateVersion(card, card.Version)
	assert.Empty(t, errors)

	// Marshal to JSON
	data, err := json.MarshalIndent(card, "", "  ")
	assert.NoError(t, err)

	// Verify the JSON contains expected fields
	var raw map[string]any
	err = json.Unmarshal(data, &raw)
	assert.NoError(t, err)

	assert.Equal(t, "1.5", raw["version"])
	assert.Equal(t, "AdaptiveCard", raw["type"])
}

func TestContainerWithVersionedFeatures(t *testing.T) {
	container := Container{
		Items: []Element{
			TextBlock{
				Text: AsPtr("Test"),
			},
		},
		Bleed:                    true,                           // Requires 1.2
		VerticalContentAlignment: AsPtr(VerticalAlignmentCenter), // Requires 1.1
		RTL:                      AsPtr(true),                    // Requires 1.5
	}

	tests := []struct {
		version     string
		expectError bool
	}{
		{"1.0", true},  // Should fail - has 1.1, 1.2, 1.5 features
		{"1.1", true},  // Should fail - has 1.2, 1.5 features
		{"1.2", true},  // Should fail - has 1.5 features
		{"1.5", false}, // Should pass - supports all features
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			errors := ValidateVersion(container, tt.version)

			if tt.expectError {
				assert.NotEmpty(t, errors)
			} else {
				assert.Empty(t, errors)
			}
		})
	}
}
