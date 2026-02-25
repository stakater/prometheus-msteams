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
	"testing"

	"github.com/stakater/prometheus-msteams/pkg/adaptivecards"
	"github.com/stakater/prometheus-msteams/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_templatedCard_Convert(t *testing.T) {
	tests := []struct {
		name              string
		promAlertFile     string
		templateFile      string
		escapeUnderscores bool
		want              adaptivecards.WorkflowConnectorCard
		wantErr           bool
	}{
		{
			name:              "do not escape underscores",
			promAlertFile:     "prom_post_request.json",
			templateFile:      "default-message-workflow-card.tmpl",
			escapeUnderscores: false,
			want: adaptivecards.WorkflowConnectorCard{
				Type: "message",
				Attachments: []adaptivecards.AdaptiveCardItem{
					{
						ContentType: "application/vnd.microsoft.card.adaptive",
						Content: adaptivecards.AdaptiveCard{
							Schema: "http://adaptivecards.io/schemas/adaptive-card.json",
							Body: []adaptivecards.Element{
								&adaptivecards.TextBlock{
									Text:   adaptivecards.AsPtr("Prometheus Alert (Firing)"),
									Style:  adaptivecards.AsPtr(adaptivecards.TextBlockStyleHeading),
									Size:   adaptivecards.AsPtr(adaptivecards.FontSize("medium")),
									Weight: adaptivecards.AsPtr(adaptivecards.FontWeight("bolder")),
									Color:  adaptivecards.AsPtr(adaptivecards.ColorNone),
								},
								&adaptivecards.TextBlock{
									Text: adaptivecards.AsPtr("Prometheus Test"),
									Wrap: true,
								},
								&adaptivecards.TextBlock{
									Text: adaptivecards.AsPtr("[10.80.40.11 reported high memory usage with 23.28%.](http://docker.for.mac.host.internal:9093)"),
									Wrap: true,
								},
								&adaptivecards.FactSet{
									Facts: []adaptivecards.Fact{
										{Title: "", Value: ""},
										{Title: "summary", Value: "Server High Memory usage"},
										{Title: "alertname", Value: "high_memory_load"},
										{Title: "instance", Value: "instance-with-hyphen_and_underscore"},
										{Title: "job", Value: "docker_nodes"},
										{Title: "monitor", Value: "master"},
										{Title: "severity", Value: "warning"},
									},
								},
							},
							MsTeams: &adaptivecards.TeamsCardProperties{
								Width: adaptivecards.AsPtr(adaptivecards.TeamsCardWidthFull),
							},
							Version: "1.2",
							BackgroundImage: &adaptivecards.BackgroundImage{
								URL:      "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAECAIAAADAusJtAAAAEklEQVR42mL4v5QBGQMCAAD//ziMBpHr7M3UAAAAAElFTkSuQmCC",
								FillMode: adaptivecards.AsPtr(adaptivecards.ImageFillModeRepeatHorizontally),
							},
						},
					},
				},
			},
		},
		{
			name:              "escape underscores",
			promAlertFile:     "prom_post_request.json",
			templateFile:      "default-message-workflow-card.tmpl",
			escapeUnderscores: true,
			want: adaptivecards.WorkflowConnectorCard{
				Type: "message",
				Attachments: []adaptivecards.AdaptiveCardItem{
					{
						ContentType: "application/vnd.microsoft.card.adaptive",
						Content: adaptivecards.AdaptiveCard{
							Schema: "http://adaptivecards.io/schemas/adaptive-card.json",
							Body: []adaptivecards.Element{
								&adaptivecards.TextBlock{
									Text:   adaptivecards.AsPtr("Prometheus Alert (Firing)"),
									Style:  adaptivecards.AsPtr(adaptivecards.TextBlockStyleHeading),
									Size:   adaptivecards.AsPtr(adaptivecards.FontSize("medium")),
									Weight: adaptivecards.AsPtr(adaptivecards.FontWeight("bolder")),
									Color:  adaptivecards.AsPtr(adaptivecards.ColorNone),
								},
								&adaptivecards.TextBlock{
									Text: adaptivecards.AsPtr("Prometheus Test"),
									Wrap: true,
								},
								&adaptivecards.TextBlock{
									Text: adaptivecards.AsPtr("[10.80.40.11 reported high memory usage with 23.28%.](http://docker.for.mac.host.internal:9093)"),
									Wrap: true,
								},
								&adaptivecards.FactSet{
									Facts: []adaptivecards.Fact{
										{Title: "", Value: ""},
										{Title: "summary", Value: "Server High Memory usage"},
										{Title: "alertname", Value: "high\\_memory\\_load"},
										{Title: "instance", Value: "instance-with-hyphen\\_and\\_underscore"},
										{Title: "job", Value: "docker\\_nodes"},
										{Title: "monitor", Value: "master"},
										{Title: "severity", Value: "warning"},
									},
								},
							},
							MsTeams: &adaptivecards.TeamsCardProperties{
								Width: adaptivecards.AsPtr(adaptivecards.TeamsCardWidthFull),
							},
							Version: "1.2",
							BackgroundImage: &adaptivecards.BackgroundImage{
								URL:      "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAECAIAAADAusJtAAAAEklEQVR42mL4v5QBGQMCAAD//ziMBpHr7M3UAAAAAElFTkSuQmCC",
								FillMode: adaptivecards.AsPtr(adaptivecards.ImageFillModeRepeatHorizontally),
							},
						},
					},
				},
			},
		},
		{
			name:              "action card",
			promAlertFile:     "prom_post_request.json",
			templateFile:      "action-message-card.tmpl",
			escapeUnderscores: true,
			want:              adaptivecards.WorkflowConnectorCard{},
			wantErr:           true, // Old O365 template not compatible with Workflow
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			templateFile := testutils.GetTestDataFilePath(tt.templateFile)
			tmpl, err := ParseTemplateFile(templateFile)
			if err != nil {
				t.Fatal(err)
			}

			promAlertFile := testutils.GetTestDataFilePath(tt.promAlertFile)
			a, err := testutils.ParseWebhookJSONFromFile(promAlertFile)
			if err != nil {
				t.Fatal(err)
			}

			m := NewTemplatedCardCreator(tmpl, tt.escapeUnderscores)

			got, err := m.Convert(context.Background(), a)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
