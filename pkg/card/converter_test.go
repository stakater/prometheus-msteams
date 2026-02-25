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
package card

import (
	"context"
	"errors"
	"testing"

	"github.com/prometheus/alertmanager/notify/webhook"
	"github.com/prometheus/alertmanager/template"
	"github.com/stakater/prometheus-msteams/pkg/adaptivecards"
	"github.com/stakater/prometheus-msteams/pkg/utility"
	"github.com/stretchr/testify/assert"
)

// mockConverter is a mock implementation of Converter
type mockConverter struct {
	convertResponse adaptivecards.WorkflowConnectorCard
	convertErr      error
}

func (m mockConverter) Convert(_ context.Context, _ webhook.Message) (adaptivecards.WorkflowConnectorCard, error) {
	return m.convertResponse, m.convertErr
}

func TestLoggingMiddleware_Convert_Success(t *testing.T) {
	logger := utility.NewLogger(utility.LogFormatFmt, false)

	mockConv := mockConverter{
		convertResponse: adaptivecards.WorkflowConnectorCard{
			Type: "message",
			Attachments: []adaptivecards.AdaptiveCardItem{
				{
					ContentType: "application/vnd.microsoft.card.adaptive",
					Content: adaptivecards.AdaptiveCard{
						Version: "1.4",
						Body: []adaptivecards.Element{
							&adaptivecards.TextBlock{
								Text: adaptivecards.AsPtr("Test"),
							},
						},

						MsTeams: &adaptivecards.TeamsCardProperties{
							Width: adaptivecards.AsPtr(adaptivecards.TeamsCardWidthFull),
						},
					},
				},
			},
		},
		convertErr: nil,
	}

	middleware := loggingMiddleware{logger, mockConv}

	wm := webhook.Message{
		Version:  "4",
		GroupKey: "test-group",
		Data:     &template.Data{},
	}

	card, err := middleware.ConvertWorkflow(context.Background(), wm)

	assert.NoError(t, err)
	assert.Equal(t, "message", card.Type)
	assert.Equal(t, 1, len(card.Attachments))
}

func TestLoggingMiddleware_ConvertWorkflow_WithManyActions(t *testing.T) {
	logger := utility.NewLogger(utility.LogFormatFmt, false)

	// Create a workflow card with more than 5 actions (should trigger warning)
	mockConv := mockConverter{
		convertResponse: adaptivecards.WorkflowConnectorCard{
			Type: "message",
			Attachments: []adaptivecards.AdaptiveCardItem{
				{
					ContentType: "application/vnd.microsoft.card.adaptive",
					Content: adaptivecards.AdaptiveCard{
						Version: "1.4",
						Body:    []adaptivecards.Element{},
						MsTeams: &adaptivecards.TeamsCardProperties{Width: adaptivecards.AsPtr(adaptivecards.TeamsCardWidthFull)},
						Actions: []adaptivecards.Action{
							&adaptivecards.ActionExecute{},
							&adaptivecards.ActionExecute{},
							&adaptivecards.ActionExecute{},
							&adaptivecards.ActionExecute{},
							&adaptivecards.ActionExecute{},
							&adaptivecards.ActionExecute{}, // 6th action, should trigger warning
						},
					},
				},
			},
		},
		convertErr: nil,
	}

	middleware := loggingMiddleware{logger, mockConv}

	wm := webhook.Message{
		Version:  "4",
		GroupKey: "test-group",
		Data:     &template.Data{},
	}

	card, err := middleware.ConvertWorkflow(context.Background(), wm)

	assert.NoError(t, err)
	assert.Equal(t, 6, len(card.Attachments[0].Content.Actions))
}

func TestLoggingMiddleware_ConvertWorkflow_Error(t *testing.T) {
	logger := utility.NewLogger(utility.LogFormatFmt, false)

	mockConv := mockConverter{
		convertErr: errors.New("workflow conversion error"),
	}

	middleware := loggingMiddleware{logger, mockConv}

	wm := webhook.Message{
		Version:  "4",
		GroupKey: "test-group",
		Data:     &template.Data{},
	}

	_, err := middleware.ConvertWorkflow(context.Background(), wm)

	assert.Error(t, err)
	assert.Equal(t, "workflow conversion error", err.Error())
}

func TestLoggingMiddleware_ConvertWorkflow_MultipleAttachments(t *testing.T) {
	logger := utility.NewLogger(utility.LogFormatFmt, false)

	mockConv := mockConverter{
		convertResponse: adaptivecards.WorkflowConnectorCard{
			Type: "message",
			Attachments: []adaptivecards.AdaptiveCardItem{
				{
					ContentType: "application/vnd.microsoft.card.adaptive",
					Content: adaptivecards.AdaptiveCard{
						Version: "1.4",
						Body: []adaptivecards.Element{
							adaptivecards.TextBlock{
								Text: adaptivecards.AsPtr("First"),
							},
						},
						MsTeams: &adaptivecards.TeamsCardProperties{Width: adaptivecards.AsPtr(adaptivecards.TeamsCardWidthFull)},
					},
				},
				{
					ContentType: "application/vnd.microsoft.card.adaptive",
					Content: adaptivecards.AdaptiveCard{
						Version: "1.4",
						Body: []adaptivecards.Element{
							adaptivecards.TextBlock{
								Text: adaptivecards.AsPtr("Second"),
							},
						},
						MsTeams: &adaptivecards.TeamsCardProperties{
							Width: adaptivecards.AsPtr(adaptivecards.TeamsCardWidthFull),
						},
					},
				},
			},
		},
		convertErr: nil,
	}

	middleware := loggingMiddleware{logger, mockConv}

	wm := webhook.Message{
		Version:  "4",
		GroupKey: "test-group",
		Data:     &template.Data{},
	}

	card, err := middleware.ConvertWorkflow(context.Background(), wm)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(card.Attachments))
}
