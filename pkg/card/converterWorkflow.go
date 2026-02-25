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

// Package card provides the logic for converting an alert manager webhook
// message to an Office365ConnectorCard.
package card

import (
	"context"
	"time"

	"github.com/prometheus/alertmanager/notify/webhook"
	"github.com/stakater/prometheus-msteams/pkg/adaptivecards"
)

func (l loggingMiddleware) ConvertWorkflow(
	ctx context.Context,
	a webhook.Message) (c adaptivecards.WorkflowConnectorCard, err error) {
	defer func(begin time.Time) {
		for _, attachment := range c.Attachments {
			if len(attachment.Content.Actions) > 5 {
				l.logger.Warn(
					"There can only be a maximum of 5 actions in a action collection",
					"actions", attachment.Content.Actions,
				)
			}
		}

		l.logger.Info(
			"alert", a,
			"card", c,
			"took", time.Since(begin),
		)
	}(time.Now())
	return l.next.Convert(ctx, a)
}
