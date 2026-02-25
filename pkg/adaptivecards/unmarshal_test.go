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
package adaptivecards

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSmartUnmarshalJSON_SimpleCard tests basic unmarshaling with SmartUnmarshalJSON
func TestSmartUnmarshalJSON_SimpleCard(t *testing.T) {
	cardJSON := `{
		"type": "AdaptiveCard",
		"version": "1.5",
		"body": [
			{
				"type": "TextBlock",
				"text": "Hello World",
				"size": "large"
			},
			{
				"type": "Image",
				"url": "https://example.com/image.png"
			}
		],
		"actions": [
			{
				"type": "Action.Submit",
				"title": "Submit"
			}
		]
	}`
	var card AdaptiveCard
	err := SmartUnmarshalJSON([]byte(cardJSON), &card)
	require.NoError(t, err)

	assert.Equal(t, "1.5", card.Version)
	assert.Len(t, card.Body, 2)
	assert.Len(t, card.Actions, 1)

	// Verify body elements
	textBlock, ok := card.Body[0].(*TextBlock)

	require.True(t, ok, "first body element should be *TextBlock")
	assert.Equal(t, "Hello World", *textBlock.Text)

	image, ok := card.Body[1].(*Image)
	require.True(t, ok, "second body element should be *Image")
	assert.Equal(t, "https://example.com/image.png", image.URL)

	// Verify actions
	submitAction, ok := card.Actions[0].(*ActionSubmit)
	require.True(t, ok, "action should be *ActionSubmit")
	assert.Equal(t, "Submit", submitAction.Title)
}

// TestSmartUnmarshalJSON_NestedContainers tests unmarshaling with nested containers
func TestSmartUnmarshalJSON_NestedContainers(t *testing.T) {
	cardJSON := `{
		"type": "AdaptiveCard",
		"version": "1.5",
		"body": [
			{
				"type": "Container",
				"items": [
					{
						"type": "TextBlock",
						"text": "Inside container"
					},
					{
						"type": "ColumnSet",
						"columns": [
							{
								"type": "Column",
								"items": [
									{
										"type": "TextBlock",
										"text": "Column 1"
									}
								]
							}
						]
					}
				]
			}
		]
	}`

	var card AdaptiveCard
	err := SmartUnmarshalJSON([]byte(cardJSON), &card)
	require.NoError(t, err)

	assert.Len(t, card.Body, 1)

	container, ok := card.Body[0].(*Container)
	require.True(t, ok, "body element should be *Container")
	assert.Len(t, container.Items, 2)

	textBlock, ok := container.Items[0].(*TextBlock)
	require.True(t, ok, "first item should be *TextBlock")
	require.NotNil(t, textBlock.Text)
	assert.Equal(t, "Inside container", *textBlock.Text)

	columnSet, ok := container.Items[1].(*ColumnSet)
	require.True(t, ok, "second item should be *ColumnSet")
	assert.Len(t, columnSet.Columns, 1)
}

// TestSmartUnmarshalJSON_AllActionTypes tests all action types
func TestSmartUnmarshalJSON_AllActionTypes(t *testing.T) {
	cardJSON := `{
		"type": "AdaptiveCard",
		"version": "1.5",
		"actions": [
			{
				"type": "Action.Submit",
				"title": "Submit"
			},
			{
				"type": "Action.OpenUrl",
				"title": "Open",
				"url": "https://example.com"
			},
			{
				"type": "Action.ShowCard",
				"title": "Show Card",
				"card": {
					"type": "AdaptiveCard",
					"version": "1.5",
					"body": []
				}
			},
			{
				"type": "Action.ToggleVisibility",
				"title": "Toggle"
			},
			{
				"type": "Action.Execute",
				"title": "Execute"
			}
		]
	}`

	var card AdaptiveCard
	err := SmartUnmarshalJSON([]byte(cardJSON), &card)
	require.NoError(t, err)

	assert.Len(t, card.Actions, 5)

	_, ok := card.Actions[0].(*ActionSubmit)
	assert.True(t, ok, "should be *ActionSubmit")

	_, ok = card.Actions[1].(*ActionOpenURL)
	assert.True(t, ok, "should be *ActionOpenURL")

	_, ok = card.Actions[2].(*ActionShowCard)
	assert.True(t, ok, "should be *ActionShowCard")

	_, ok = card.Actions[3].(*ActionToggleVisibility)
	assert.True(t, ok, "should be *ActionToggleVisibility")

	_, ok = card.Actions[4].(*ActionExecute)
	assert.True(t, ok, "should be *ActionExecute")
}

// TestSmartUnmarshalJSON_PointerFields tests unmarshaling of pointer fields
func TestSmartUnmarshalJSON_PointerFields(t *testing.T) {
	cardJSON := `{
		"type": "AdaptiveCard",
		"version": "1.5",
		"rtl": true,
		"style": "emphasis",
		"backgroundImage": {
			"url": "https://example.com/bg.png"
		}
	}`

	var card AdaptiveCard
	err := SmartUnmarshalJSON([]byte(cardJSON), &card)
	require.NoError(t, err)

	require.NotNil(t, card.RTL)
	assert.True(t, *card.RTL)

	require.NotNil(t, card.Style)
	assert.Equal(t, ContainerStyleEmphasis, *card.Style)

	require.NotNil(t, card.BackgroundImage)
	assert.Equal(t, "https://example.com/bg.png", card.BackgroundImage.URL)
}

// TestSmartUnmarshalJSON_FallbackField tests fallback field handling
func TestSmartUnmarshalJSON_FallbackField(t *testing.T) {
	t.Run("Fallback as string", func(t *testing.T) {
		cardJSON := `{
			"type": "AdaptiveCard",
			"version": "1.2",
			"body": [
				{
					"type": "TextBlock",
					"text": "Test",
					"fallback": "drop"
				}
			]
		}`

		var card AdaptiveCard
		err := SmartUnmarshalJSON([]byte(cardJSON), &card)
		require.NoError(t, err)

		require.Len(t, card.Body, 1)
		textBlock, ok := card.Body[0].(*TextBlock)
		require.True(t, ok)

		require.NotNil(t, textBlock.Fallback)
		fallbackOption, ok := textBlock.Fallback.(FallbackOption)
		require.True(t, ok, "fallback should be FallbackOption")
		assert.Equal(t, FallbackOptionDrop, fallbackOption)
	})

	t.Run("Fallback as Element", func(t *testing.T) {
		cardJSON := `{
			"type": "AdaptiveCard",
			"version": "1.2",
			"body": [
				{
					"type": "TextBlock",
					"text": "Main content",
					"fallback": {
						"type": "TextBlock",
						"text": "Fallback content"
					}
				}
			]
		}`

		var card AdaptiveCard
		err := SmartUnmarshalJSON([]byte(cardJSON), &card)
		require.NoError(t, err)

		require.Len(t, card.Body, 1)
		textBlock, ok := card.Body[0].(*TextBlock)
		require.True(t, ok)

		require.NotNil(t, textBlock.Fallback)
		fallbackElement, ok := textBlock.Fallback.(*TextBlock)
		require.True(t, ok, "fallback should be *TextBlock")
		require.NotNil(t, fallbackElement.Text)
		assert.Equal(t, "Fallback content", *fallbackElement.Text)
	})

	t.Run("Fallback as Action", func(t *testing.T) {
		cardJSON := `{
			"type": "AdaptiveCard",
			"version": "1.2",
			"actions": [
				{
					"type": "Action.Submit",
					"title": "Submit",
					"fallback": {
						"type": "Action.OpenUrl",
						"title": "Fallback Action",
						"url": "https://example.com"
					}
				}
			]
		}`

		var card AdaptiveCard
		err := SmartUnmarshalJSON([]byte(cardJSON), &card)
		require.NoError(t, err)

		require.Len(t, card.Actions, 1)
		submitAction, ok := card.Actions[0].(*ActionSubmit)
		require.True(t, ok)

		require.NotNil(t, submitAction.Fallback)
		fallbackAction, ok := submitAction.Fallback.(*ActionOpenURL)
		require.True(t, ok, "fallback should be *ActionOpenURL")
		assert.Equal(t, "Fallback Action", fallbackAction.Title)
		assert.Equal(t, "https://example.com", fallbackAction.URL)
	})

	t.Run("Card-level Fallback", func(t *testing.T) {
		cardJSON := `{
			"type": "AdaptiveCard",
			"version": "1.2",
			"body": [
				{
					"type": "TextBlock",
					"text": "Main card"
				}
			],
			"fallback": "drop"
		}`

		var card AdaptiveCard
		err := SmartUnmarshalJSON([]byte(cardJSON), &card)
		require.NoError(t, err)

		require.NotNil(t, card.Fallback)
		fallbackOption, ok := card.Fallback.(FallbackOption)
		require.True(t, ok, "card fallback should be FallbackOption")
		assert.Equal(t, FallbackOptionDrop, fallbackOption)
	})
}

// TestSmartUnmarshalJSON_InputElements tests all input types
func TestSmartUnmarshalJSON_InputElements(t *testing.T) {
	cardJSON := `{
		"type": "AdaptiveCard",
		"version": "1.5",
		"body": [
			{
				"type": "Input.Text",
				"id": "text-input",
				"placeholder": "Enter text"
			},
			{
				"type": "Input.Number",
				"id": "number-input",
				"min": 0,
				"max": 100
			},
			{
				"type": "Input.Date",
				"id": "date-input"
			},
			{
				"type": "Input.Time",
				"id": "time-input"
			},
			{
				"type": "Input.Toggle",
				"id": "toggle-input",
				"title": "Accept terms"
			},
			{
				"type": "Input.ChoiceSet",
				"id": "choice-input",
				"choices": [
					{
						"title": "Option 1",
						"value": "1"
					}
				]
			}
		]
	}`

	var card AdaptiveCard
	err := SmartUnmarshalJSON([]byte(cardJSON), &card)
	require.NoError(t, err)

	assert.Len(t, card.Body, 6)

	_, ok := card.Body[0].(*InputText)
	assert.True(t, ok, "should be *InputText")

	_, ok = card.Body[1].(*InputNumber)
	assert.True(t, ok, "should be *InputNumber")

	_, ok = card.Body[2].(*InputDate)
	assert.True(t, ok, "should be *InputDate")

	_, ok = card.Body[3].(*InputTime)
	assert.True(t, ok, "should be *InputTime")

	_, ok = card.Body[4].(*InputToggle)
	assert.True(t, ok, "should be *InputToggle")

	_, ok = card.Body[5].(*InputChoiceSet)
	assert.True(t, ok, "should be *InputChoiceSet")
}

// Example of a custom struct using SmartUnmarshalJSON
type CustomContainer struct {
	Type    string    `json:"type"`
	ID      string    `json:"id"`
	Items   []Element `json:"items,omitempty"`
	Actions []Action  `json:"actions,omitempty"`
}

func (c CustomContainer) isElement() {}

func (c *CustomContainer) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

func TestSmartUnmarshalJSON_CustomStruct(t *testing.T) {
	// Register the custom type
	RegisterType("CustomContainer", CustomContainer{})

	customJSON := `{
		"type": "CustomContainer",
		"id": "my-custom",
		"items": [
			{
				"type": "TextBlock",
				"text": "Custom item"
			}
		],
		"actions": [
			{
				"type": "Action.Submit",
				"title": "Submit"
			}
		]
	}`

	var custom CustomContainer
	err := SmartUnmarshalJSON([]byte(customJSON), &custom)
	require.NoError(t, err)

	assert.Equal(t, "CustomContainer", custom.Type)
	assert.Equal(t, "my-custom", custom.ID)
	assert.Len(t, custom.Items, 1)
	assert.Len(t, custom.Actions, 1)

	textBlock, ok := custom.Items[0].(*TextBlock)
	require.True(t, ok)
	require.NotNil(t, textBlock.Text)
	assert.Equal(t, "Custom item", *textBlock.Text)
}

// Benchmark comparing SmartUnmarshalJSON vs manual UnmarshalJSON
func BenchmarkSmartUnmarshalJSON(b *testing.B) {
	cardJSON := []byte(`{
		"type": "AdaptiveCard",
		"version": "1.5",
		"body": [
			{"type": "TextBlock", "text": "Line 1"},
			{"type": "TextBlock", "text": "Line 2"},
			{"type": "Image", "url": "https://example.com/image.png"}
		],
		"actions": [
			{"type": "Action.Submit", "title": "Submit"},
			{"type": "Action.OpenUrl", "title": "Open", "url": "https://example.com"}
		]
	}`)

	b.Run("SmartUnmarshalJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var card AdaptiveCard
			if err := SmartUnmarshalJSON(cardJSON, &card); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ManualUnmarshalJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var card AdaptiveCard
			if err := json.Unmarshal(cardJSON, &card); err != nil {
				b.Fatal(err)
			}
		}
	})
}
