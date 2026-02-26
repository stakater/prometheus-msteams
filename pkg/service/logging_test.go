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

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/prometheus/alertmanager/notify/webhook"
	"github.com/prometheus/alertmanager/template"
	"github.com/stakater/prometheus-msteams/pkg/utility"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockService is a mock implementation of service.Service
type mockService struct {
	response []PostResponse
	err      error
}

func (m mockService) Post(_ context.Context, _ webhook.Message) ([]PostResponse, error) {
	return m.response, m.err
}

func TestNewLoggingService(t *testing.T) {
	logger := utility.NewLogger(utility.LogFormatFmt, false)
	assert.NotNil(t, logger)

	mockSvc := mockService{
		response: []PostResponse{{WebhookURL: "https://test.com", Status: 200}},
		err:      nil,
	}

	logging := NewLoggingService(logger, mockSvc)
	assert.NotNil(t, logging)
}

func TestLoggingService_Post_Success(t *testing.T) {
	logger := utility.NewLogger(utility.LogFormatFmt, false)

	mockSvc := mockService{
		response: []PostResponse{
			{WebhookURL: "https://test.webhook.com", Status: 200, Message: "success"},
		},
		err: nil,
	}

	logging := NewLoggingService(logger, mockSvc)

	msg := webhook.Message{
		Data: &template.Data{
			Receiver: "test-receiver",
			Status:   "firing",
		},
	}

	responses, err := logging.Post(context.Background(), msg)

	require.NoError(t, err)
	assert.Len(t, responses, 1)
	assert.Equal(t, 200, responses[0].Status)
	assert.Equal(t, "success", responses[0].Message)
	assert.Equal(t, "https://test.webhook.com", responses[0].WebhookURL)
}

func TestLoggingService_Post_Error(t *testing.T) {
	logger := utility.NewLogger(utility.LogFormatFmt, false)

	expectedErr := errors.New("webhook post failed")
	mockSvc := mockService{
		response: []PostResponse{
			{WebhookURL: "https://test.webhook.com", Status: 500, Message: "internal error"},
		},
		err: expectedErr,
	}

	logging := NewLoggingService(logger, mockSvc)

	msg := webhook.Message{
		Data: &template.Data{
			Receiver: "test-receiver",
			Status:   "firing",
		},
	}

	responses, err := logging.Post(context.Background(), msg)

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Len(t, responses, 1)
	assert.Equal(t, 500, responses[0].Status)
}

func TestLoggingService_Post_MultipleResponses(t *testing.T) {
	logger := utility.NewLogger(utility.LogFormatFmt, false)

	mockSvc := mockService{
		response: []PostResponse{
			{WebhookURL: "https://webhook1.com", Status: 200, Message: "success 1"},
			{WebhookURL: "https://webhook2.com", Status: 200, Message: "success 2"},
			{WebhookURL: "https://webhook3.com", Status: 200, Message: "success 3"},
		},
		err: nil,
	}

	logging := NewLoggingService(logger, mockSvc)

	msg := webhook.Message{
		Data: &template.Data{
			Receiver: "test-receiver",
			Status:   "firing",
		},
	}

	responses, err := logging.Post(context.Background(), msg)

	require.NoError(t, err)
	assert.Len(t, responses, 3)
	for i, resp := range responses {
		assert.Equal(t, 200, resp.Status)
		assert.Contains(t, resp.Message, "success")
		assert.Contains(t, resp.WebhookURL, "webhook")
		t.Logf("Response %d: %+v", i+1, resp)
	}
}

func TestLoggingService_Post_EmptyResponses(t *testing.T) {
	logger := utility.NewLogger(utility.LogFormatJSON, true)

	mockSvc := mockService{
		response: []PostResponse{},
		err:      nil,
	}

	logging := NewLoggingService(logger, mockSvc)

	msg := webhook.Message{
		Data: &template.Data{
			Receiver: "test-receiver",
			Status:   "resolved",
		},
	}

	responses, err := logging.Post(context.Background(), msg)

	require.NoError(t, err)
	assert.Empty(t, responses)
}
