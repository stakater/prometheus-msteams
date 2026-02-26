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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/alertmanager/notify/webhook"
	"github.com/prometheus/alertmanager/template"
	"github.com/stakater/prometheus-msteams/pkg/adaptivecards"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockConverter implements card.Converter for testing
type mockConverter struct {
	workflowCard adaptivecards.WorkflowConnectorCard
	err          error
}

func (m mockConverter) Convert(_ context.Context, _ webhook.Message) (adaptivecards.WorkflowConnectorCard, error) {
	return m.workflowCard, m.err
}

func TestNewSimpleService(t *testing.T) {
	converter := mockConverter{}
	client := &http.Client{}
	webhookURL := "https://test.webhook.com"
	webhookType := Workflow

	svc := NewSimpleService(converter, client, webhookURL, webhookType)
	assert.NotNil(t, svc)
}

func TestSimpleService_Post_Workflow(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	workflowCard := adaptivecards.WorkflowConnectorCard{
		Type: "message",
		Attachments: []adaptivecards.AdaptiveCardItem{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
			},
		},
	}

	converter := mockConverter{
		workflowCard: workflowCard,
	}

	svc := NewSimpleService(converter, server.Client(), server.URL, Workflow)

	msg := webhook.Message{
		Data: &template.Data{
			Receiver: "test",
			Status:   "firing",
		},
	}

	responses, err := svc.Post(context.Background(), msg)

	require.NoError(t, err)
	assert.Len(t, responses, 0) // Workflow doesn't append to prs in current implementation
}

func TestSimpleService_Post_ConverterError(t *testing.T) {
	converter := mockConverter{
		err: fmt.Errorf("conversion failed"),
	}

	svc := NewSimpleService(converter, &http.Client{}, "https://test.com", Workflow)

	msg := webhook.Message{
		Data: &template.Data{
			Receiver: "test",
			Status:   "firing",
		},
	}

	responses, err := svc.Post(context.Background(), msg)

	require.Error(t, err)
	assert.Nil(t, responses)
	assert.Contains(t, err.Error(), "failed to parse webhook message")
}

func TestSimpleService_Post_HTTPError(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	workflowCard := adaptivecards.WorkflowConnectorCard{
		Type: "message",
		Attachments: []adaptivecards.AdaptiveCardItem{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
			},
		},
	}

	converter := mockConverter{
		workflowCard: workflowCard,
	}

	svc := NewSimpleService(converter, server.Client(), server.URL, Workflow)

	msg := webhook.Message{
		Data: &template.Data{
			Receiver: "test",
			Status:   "firing",
		},
	}

	responses, err := svc.Post(context.Background(), msg)

	require.NoError(t, err)
	assert.Len(t, responses, 0) // Workflow implementation doesn't append HTTP errors to responses
}

func TestSimpleService_PostWorkflowWebhook_ConverterError(t *testing.T) {
	converter := mockConverter{
		err: fmt.Errorf("workflow conversion failed"),
	}

	svc := NewSimpleService(converter, &http.Client{}, "https://test.com", Workflow)

	msg := webhook.Message{
		Data: &template.Data{
			Receiver: "test",
			Status:   "firing",
		},
	}

	responses, err := svc.Post(context.Background(), msg)

	require.Error(t, err)
	assert.Nil(t, responses)
	assert.Contains(t, err.Error(), "failed to parse webhook message")
}
