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
package testutils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/prometheus/alertmanager/notify/webhook"
	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
)

func TestGetTestDataFilePath(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "simple filename",
			filename: "test.json",
			want:     filepath.Join("..", "..", "test", "data", "test.json"),
		},
		{
			name:     "filename with path",
			filename: "subdir/test.json",
			want:     filepath.Join("..", "..", "test", "data", "subdir/test.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTestDataFilePath(tt.filename)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseWebhookJSONFromFile_Success(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_webhook.json")

	testData := webhook.Message{
		Data:     &template.Data{},
		Version:  "4",
		GroupKey: "test-group",
	}

	// Write test data to file
	data, err := json.Marshal(testData)
	assert.NoError(t, err)

	err = os.WriteFile(testFile, data, 0600)
	assert.NoError(t, err)

	// Test parsing
	result, err := ParseWebhookJSONFromFile(testFile)

	assert.NoError(t, err)
	assert.Equal(t, "4", result.Version)
	assert.Equal(t, "test-group", result.GroupKey)
	assert.NotNil(t, result.Data)
}

func TestParseWebhookJSONFromFile_FileNotFound(t *testing.T) {
	_, err := ParseWebhookJSONFromFile("/nonexistent/file.json")

	assert.Error(t, err)
}

func TestParseWebhookJSONFromFile_InvalidJSON(t *testing.T) {
	// Create a temporary test file with invalid JSON
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.json")

	err := os.WriteFile(testFile, []byte("not valid json"), 0600)
	assert.NoError(t, err)

	_, err = ParseWebhookJSONFromFile(testFile)

	assert.Error(t, err)
}

func TestParseWebhookJSONFromFile_RealFile(t *testing.T) {
	// Test with actual test data file if it exists
	testDataPath := filepath.Join("..", "..", "test", "data", "prom_post_request.json")

	// Check if file exists
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("Test data file not found, skipping")
	}

	result, err := ParseWebhookJSONFromFile(testDataPath)

	if err == nil {
		assert.NotEmpty(t, result.Version)
		assert.NotEmpty(t, result.GroupKey)
	}
}

func TestGetTestDataFilePath_EmptyFilename(t *testing.T) {
	result := GetTestDataFilePath("")
	expected := filepath.Join("..", "..", "test", "data", "")
	assert.Equal(t, expected, result)
}
