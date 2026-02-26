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

// Package adaptivecards provides Go structs and helper functions for working
// with Microsoft Adaptive Cards.
package adaptivecards

func init() {
	// RegisterType("AdaptiveCardItem", AdaptiveCardItem{})
	RegisterType("message", WorkflowConnectorCard{})
}

// region AdaptiveCardItem

// AdaptiveCardItem represents a card for workflow.
type AdaptiveCardItem struct {
	ContentType string       `json:"contentType"` // Always "application/vnd.microsoft.card.adaptive"
	ContentURL  *string      `json:"contentUrl"`  // Use a pointer to handle null values
	Content     AdaptiveCard `json:"content"`
}

/*
// MarshalJSON ensures that the "type" field is included when marshaling a
// AdaptiveCardItem to JSON.
func (a AdaptiveCardItem) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "Action.OpenUrl")
}

// UnmarshalJSON ensures we only unmarshal if the type is correct
func (a *AdaptiveCardItem) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}
*/

// endregion AdaptiveCardItem

// region WorkflowConnectorCard

// WorkflowConnectorCard represents a card for workflow.
type WorkflowConnectorCard struct {
	Type        string             `json:"type"`
	Attachments []AdaptiveCardItem `json:"attachments"`
}

// MarshalJSON ensures that the "type" field is included when marshaling a
// WorkflowConnectorCard to JSON.
func (a WorkflowConnectorCard) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "Action.OpenUrl")
}

// UnmarshalJSON ensures we only unmarshal if the type is correct
func (a *WorkflowConnectorCard) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion WorkflowConnectorCard
