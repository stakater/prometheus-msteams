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

// Package testutils provides utility functions for testing, such as comparing
// JSON output to golden files and parsing test data from JSON files.
package testutils

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/prometheus/alertmanager/notify/webhook"
)

// GetTestDataFilePath returns the path to a test data file given its filename.
func GetTestDataFilePath(filename string) string {
	return filepath.Join("..", "..", "test", "data", filename)
}

// ParseWebhookJSONFromFile is a helper for parsing webhook data from JSON files.
func ParseWebhookJSONFromFile(f string) (webhook.Message, error) {
	filePath := filepath.Clean(f)
	b, err := os.ReadFile(filePath)
	if err != nil {
		return webhook.Message{}, err
	}
	var w webhook.Message
	if err := json.Unmarshal(b, &w); err != nil {
		return webhook.Message{}, err
	}
	return w, nil
}
