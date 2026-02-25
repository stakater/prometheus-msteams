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

// Package service provides the Service interface and related
// implementations for handling webhook messages.
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/prometheus/alertmanager/notify/webhook"
	"github.com/stakater/prometheus-msteams/pkg/card"
	"go.opencensus.io/trace"
)

// WebhookType is the type of Microsoft Teams webhook.
type WebhookType string

const (
	// O365 is the original Microsoft Teams webhook type,
	// which uses Office365ConnectorCard as the message format.
	O365 WebhookType = "o365"
	// Workflow is the Microsoft Teams webhook type that uses the
	// Microsoft Workflow card format.
	Workflow WebhookType = "microsoft-workflow"
)

// PostResponse is the prometheus msteams service response.
type PostResponse struct {
	WebhookURL string `json:"webhook_url"`
	Status     int    `json:"status"`
	Message    string `json:"message"`
}

// Service is the Alertmanager to Microsoft Teams webhook service.
type Service interface {
	Post(context.Context, webhook.Message) (resp []PostResponse, err error)
}

type simpleService struct {
	converter   card.Converter
	client      *http.Client
	webhookURL  string
	webhookType WebhookType
}

// NewSimpleService creates a simpleService.
func NewSimpleService(converter card.Converter, client *http.Client, webhookURL string, webhookType WebhookType) Service {
	return simpleService{converter, client, webhookURL, webhookType}
}

func (s simpleService) Post(ctx context.Context, wm webhook.Message) ([]PostResponse, error) {
	ctx, span := trace.StartSpan(ctx, "simpleService.Post")
	defer span.End()

	prs := []PostResponse{}

	c, err := s.converter.Convert(ctx, wm)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook message: %w", err)
	}

	_, err = s.post(ctx, c, s.webhookURL)
	if err != nil {
		return prs, err
	}

	return prs, nil
}

func (s simpleService) post(ctx context.Context, c interface{}, url string) (PostResponse, error) {
	ctx, span := trace.StartSpan(ctx, "simpleService.post")
	defer span.End()

	pr := PostResponse{WebhookURL: url}

	b, err := json.Marshal(c)
	if err != nil {
		err = fmt.Errorf("failed to decoding JSON card: %w", err)
		return pr, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(b))
	if err != nil {
		err = fmt.Errorf("failed to creating a request: %w", err)
		return pr, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req) // nolint:gosec
	if err != nil {
		err = fmt.Errorf("http client failed: %w", err)
		return pr, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			span.AddAttributes(trace.StringAttribute("close_body_error", fmt.Sprintf("failed to close response body: %v", err)))
		}
	}()

	pr.Status = resp.StatusCode

	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed reading http response body: %w", err)
		pr.Message = err.Error()
		return pr, err
	}
	pr.Message = string(rb)

	return pr, nil
}
