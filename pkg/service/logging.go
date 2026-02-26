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
	"context"

	"github.com/prometheus/alertmanager/notify/webhook"
	"github.com/stakater/prometheus-msteams/pkg/utility"
)

// loggingService is a logging middleware for Service.
type loggingService struct {
	logger *utility.Logger
	next   Service
}

// NewLoggingService creates a loggingService.
func NewLoggingService(logger *utility.Logger, next Service) Service {
	return loggingService{logger, next}
}

func (s loggingService) Post(ctx context.Context, wm webhook.Message) (prs []PostResponse, err error) {
	defer func() {
		for _, pr := range prs {
			s.logger.Debug(
				"response_message", pr.Message,
				"response_status", pr.Status,
				"webhook_url", pr.WebhookURL,
				"err", err,
			)
		}
	}()
	return s.next.Post(ctx, wm)
}
