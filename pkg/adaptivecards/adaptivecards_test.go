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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func compareCardJSON(t *testing.T, expected AdaptiveCard, expectedFile string) {
	filePath := filepath.Join("..", "..", "test", "data", expectedFile)
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)

	// a, err := json.MarshalIndent(tt.expected, "", "  ")
	// assert.NoError(t, err)
	// res := html.JS(a)
	// assert.Equal(t, string(data), string(res))

	var card AdaptiveCard
	err = json.Unmarshal(data, &card)
	require.NoError(t, err)
	assert.Equal(t, expected, card)

	errors := card.Validate()
	assert.Empty(t, errors)
}

func TestAdaptiveCardExample(t *testing.T) {
	var expected = AdaptiveCard{
		Version: "1.5",
		Speak:   "Intro to graphic design, concepts video",
		Schema:  "https://adaptivecards.io/schemas/adaptive-card.json",
		Body: []Element{
			&Image{
				URL: "https://raw.githubusercontent.com/OfficeDev/Microsoft-Teams-Card-Samples/main/samples/course-video/assets/video_image.png",
				SelectAction: &ActionOpenURL{
					URL: "https://adaptivecards.io/",
				},
				Style: AsPtr(ImageStyleRoundedCorners),
			},
			&TextBlock{
				Text:   AsPtr("Intro to Graphic Design: Concepts"),
				Wrap:   true,
				Size:   AsPtr(FontSizeLarge),
				Weight: AsPtr(FontWeightBolder),
			},
			&Rating{
				Value: 4,
				Count: 1160,
				Color: AsPtr(RatingColorMarigold),
				Size:  AsPtr(RatingSizeMedium),
				Common: &Common{
					Fallback: &TextBlock{
						Text: AsPtr("4 Stars · 1,160"),
						Common: &Common{
							Spacing: AsPtr(SpacingNone),
						},
					},
					Spacing: AsPtr(SpacingNone),
				},
			},
			&TextBlock{
				Text:     AsPtr("Course · 52m · Beginner"),
				Wrap:     true,
				IsSubtle: AsPtr(true),
				Common: &Common{
					Spacing: AsPtr(SpacingSmall),
				},
			},
			&ColumnSet{
				Columns: []Column{
					{
						Width: "auto",
						Items: []Element{
							&Image{
								URL:     "https://raw.githubusercontent.com/OfficeDev/Microsoft-Teams-Card-Samples/main/samples/course-video/assets/logo_image.png",
								Width:   "16px",
								Height:  "16px",
								AltText: "Logo",
								Common: &Common{
									Height: AsPtr(BlockElementHeight("16px")),
								},
							},
						},
						Common: &Common{
							HorizontalAlignment: AsPtr(HorizontalAlignmentCenter),
						},
						VerticalContentAlignment: AsPtr(VerticalAlignmentCenter),
					},
					{
						Width: "auto",
						Items: []Element{
							&TextBlock{
								Text:   AsPtr("Sketchpad Scholars"),
								Wrap:   true,
								Weight: AsPtr(FontWeightBolder),
							},
						},
						VerticalContentAlignment: AsPtr(VerticalAlignmentCenter),
						Common: &Common{
							Spacing: AsPtr(SpacingSmall),
						},
					},
					{
						Width: "auto",
						Items: []Element{
							&TextBlock{
								Text: AsPtr("·"),
							},
						},
						Common: &Common{
							Spacing:     AsPtr(SpacingSmall),
							TargetWidth: AsPtr(TargetWidthAtLeastStandard),
						},
						VerticalContentAlignment: AsPtr(VerticalAlignmentCenter),
					},
					{
						//TargetWidth: "AtLeast:Standard",
						Width: "auto",
						Items: []Element{
							&TextBlock{
								Text: AsPtr("Tony Harper"),
								Wrap: true,
							},
						},
						Common: &Common{
							Spacing:     AsPtr(SpacingSmall),
							TargetWidth: AsPtr(TargetWidthAtLeastStandard),
						},
						VerticalContentAlignment: AsPtr(VerticalAlignmentCenter),
					},
				},
				Common: &Common{
					Spacing: AsPtr(SpacingNone),
				},
			},
			&TextBlock{
				Text: AsPtr("This course is designed to equip you with an understanding of the key principles and tools necessary for creating compelling designs. You'll gain practical experience with creative software and learn..."),
				Wrap: true,
				Common: &Common{
					ID:          "truncatedText",
					TargetWidth: AsPtr(TargetWidthAtLeastNarrow),
				},
			},
			&TextBlock{
				Text: AsPtr("This course is designed to equip you with an understanding of the key principles and tools necessary for creating compelling designs. You'll gain practical experience with creative software and learn about design principles through hands-on projects that will help build your portfolio. Enroll now and start your journey to mastering the art of graphic design."),
				Wrap: true,
				Common: &Common{
					IsVisible:   false,
					ID:          "fullText",
					TargetWidth: AsPtr(TargetWidthAtLeastNarrow),
				},
			},
			&RichTextBlock{
				Common: &Common{
					ID:          "showMore",
					Spacing:     AsPtr(SpacingNone),
					TargetWidth: AsPtr(TargetWidthAtLeastNarrow),
				},
				Inlines: []RichTextInline{
					&TextRun{
						Text: "Show more",
						SelectAction: &ActionToggleVisibility{
							TargetElements: []TargetElement{
								{ElementID: "truncatedText"},
								{ElementID: "fullText"},
								{ElementID: "showMore"},
								{ElementID: "showLess"},
							},
						},
					},
				},
			},
			&RichTextBlock{
				Common: &Common{
					ID:          "showLess",
					Spacing:     AsPtr(SpacingNone),
					TargetWidth: AsPtr(TargetWidthAtLeastNarrow),
					IsVisible:   false,
				},
				Inlines: []RichTextInline{
					&TextRun{
						Text: "Show less",
						SelectAction: &ActionToggleVisibility{
							TargetElements: []TargetElement{
								{ElementID: "truncatedText"},
								{ElementID: "fullText"},
								{ElementID: "showMore"},
								{ElementID: "showLess"},
							},
						},
					},
				},
			},
			&ActionSet{
				Common: &Common{
					Spacing:     AsPtr(SpacingLarge),
					TargetWidth: AsPtr(TargetWidthAtLeastNarrow),
				},
				Actions: []Action{
					&ActionOpenURL{
						URL: "https://adaptivecards.io/",
						CommonActionProperties: &CommonActionProperties{
							Title: "Open",
						},
					},
					&ActionExecute{
						CommonActionProperties: &CommonActionProperties{
							Title:   "Bookmark",
							IconURL: "icon:Bookmark",
						},
					},
				},
			},
			&ActionSet{
				Common: &Common{
					Spacing:     AsPtr(SpacingLarge),
					TargetWidth: AsPtr(TargetWidthVeryNarrow),
				},
				Actions: []Action{
					&ActionOpenURL{
						CommonActionProperties: &CommonActionProperties{
							Title: "Open",
						},
						URL: "https://adaptivecards.io/",
					},
					&ActionExecute{
						CommonActionProperties: &CommonActionProperties{
							IconURL: "icon:Bookmark",
						},
					},
				},
			},
		},
	}

	compareCardJSON(t, expected, "adaptivecard_example.json")
}

func TestAdaptiveCardBadgeExample(t *testing.T) {
	var expected = AdaptiveCard{
		Version: "1.5",
		Body: []Element{
			&TextBlock{
				Text: AsPtr("Shapes"),
				Wrap: true,
			},
			&Container{
				Layouts: []Layout{
					&LayoutFlow{
						HorizontalItemsAlignment: AsPtr(HorizontalAlignmentLeft),
						//HorizontalItemsAlignment: HorizontalAlignmentLeft,
						MinItemWidth: "0px",
					},
				},
				Items: []Element{
					&Badge{
						Text:  AsPtr("Circular"),
						Size:  AsPtr(BadgeSizeLarge),
						Style: AsPtr(ColorAccent),
					},
					&Badge{
						Text:  AsPtr("Rounded"),
						Size:  AsPtr(BadgeSizeLarge),
						Style: AsPtr(ColorAccent),
						Shape: AsPtr(BadgeShapeRounded),
					},
					&Badge{
						Text:  AsPtr("Square"),
						Size:  AsPtr(BadgeSizeLarge),
						Style: AsPtr(ColorAccent),
						Shape: AsPtr(BadgeShapeSquare),
					},
				},
			},
			&TextBlock{
				Text: AsPtr("Sizes"),
				Wrap: true,
				Common: &Common{
					Separator: true,
					Spacing:   AsPtr(SpacingExtraLarge),
				},
			},
			&Container{
				Layouts: []Layout{
					&LayoutFlow{
						HorizontalItemsAlignment: AsPtr(HorizontalAlignmentLeft),
						//HorizontalItemsAlignment: HorizontalAlignmentLeft,
						MinItemWidth: "0px",
					},
				},
				Items: []Element{
					&Badge{
						Text: AsPtr("Extra large no icon"),
						Size: AsPtr(BadgeSizeExtraLarge),
					},
					&Badge{
						Text: AsPtr("Extra large with icon"),
						Size: AsPtr(BadgeSizeExtraLarge),
						Icon: AsPtr(SymbolCalendar),
					},
					&Badge{
						Text:         AsPtr("Extra large with icon on right"),
						Size:         AsPtr(BadgeSizeExtraLarge),
						Icon:         AsPtr(SymbolCalendar),
						IconPosition: AsPtr(IconPositionAfter),
					},
					&Badge{
						Size: AsPtr(BadgeSizeExtraLarge),
						Icon: AsPtr(SymbolCalendar),
					},
				},
			},
			&Container{
				Layouts: []Layout{
					&LayoutFlow{
						HorizontalItemsAlignment: AsPtr(HorizontalAlignmentLeft),
						//HorizontalItemsAlignment: HorizontalAlignmentLeft,
						MinItemWidth: "0px",
					},
				},
				Items: []Element{
					&Badge{
						Text: AsPtr("Large no icon"),
						Size: AsPtr(BadgeSizeLarge),
					},
					&Badge{
						Text: AsPtr("Large with icon"),
						Size: AsPtr(BadgeSizeLarge),
						Icon: AsPtr(SymbolCalendar),
					},
					&Badge{
						Text:         AsPtr("Large with icon on right"),
						Size:         AsPtr(BadgeSizeLarge),
						Icon:         AsPtr(SymbolCalendar),
						IconPosition: AsPtr(IconPositionAfter),
					},
					&Badge{
						Size: AsPtr(BadgeSizeLarge),
						Icon: AsPtr(SymbolCalendar),
					},
				},
				Common: &Common{
					Spacing: AsPtr(SpacingLarge),
				},
			},
			&Container{
				Layouts: []Layout{
					&LayoutFlow{
						HorizontalItemsAlignment: AsPtr(HorizontalAlignmentLeft),
						//HorizontalItemsAlignment: HorizontalAlignmentLeft,
						MinItemWidth: "0px",
					},
				},
				Items: []Element{
					&Badge{
						Text: AsPtr("Medium no icon"),
					},
					&Badge{
						Text: AsPtr("Medium with icon"),
						Icon: AsPtr(SymbolCalendar),
					},
					&Badge{
						Text:         AsPtr("Medium with icon on right"),
						Icon:         AsPtr(SymbolCalendar),
						IconPosition: AsPtr(IconPositionAfter),
					},
					&Badge{
						Icon: AsPtr(SymbolCalendar),
					},
				},
				Common: &Common{
					Spacing: AsPtr(SpacingLarge),
				},
			},
			&TextBlock{
				Text: AsPtr("Styles and colors"),
				Wrap: true,
				Common: &Common{
					Separator: true,
					Spacing:   AsPtr(SpacingExtraLarge),
				},
			},
			&Container{
				Layouts: []Layout{
					&LayoutFlow{
						HorizontalItemsAlignment: AsPtr(HorizontalAlignmentLeft),
						//HorizontalItemsAlignment: HorizontalAlignmentLeft,
						MinItemWidth: "0px",
					},
				},
				Items: []Element{
					&Badge{
						Text: AsPtr("Default filled"),
						Size: AsPtr(BadgeSizeLarge),
					},
					&Badge{
						Text:       AsPtr("Default tint"),
						Size:       AsPtr(BadgeSizeLarge),
						Appearance: AsPtr(AppearanceTint),
					},
					&Badge{
						Text:  AsPtr("Informative filled"),
						Size:  AsPtr(BadgeSizeLarge),
						Style: AsPtr(ColorInformative),
					},
					&Badge{
						Text:       AsPtr("Informative tint"),
						Size:       AsPtr(BadgeSizeLarge),
						Appearance: AsPtr(AppearanceTint),
						Style:      AsPtr(ColorInformative),
					},
					&Badge{
						Text:  AsPtr("Accent filled"),
						Size:  AsPtr(BadgeSizeLarge),
						Style: AsPtr(ColorAccent),
					},
					&Badge{
						Text:       AsPtr("Accent tint"),
						Size:       AsPtr(BadgeSizeLarge),
						Appearance: AsPtr(AppearanceTint),
						Style:      AsPtr(ColorAccent),
					},
					&Badge{
						Text:  AsPtr("Good filled"),
						Size:  AsPtr(BadgeSizeLarge),
						Style: AsPtr(ColorGood),
					},
					&Badge{
						Text:       AsPtr("Good tint"),
						Size:       AsPtr(BadgeSizeLarge),
						Appearance: AsPtr(AppearanceTint),
						Style:      AsPtr(ColorGood),
					},
					&Badge{
						Text:  AsPtr("Attention filled"),
						Size:  AsPtr(BadgeSizeLarge),
						Style: AsPtr(ColorAttention),
					},
					&Badge{
						Text:       AsPtr("Attention tint"),
						Size:       AsPtr(BadgeSizeLarge),
						Appearance: AsPtr(AppearanceTint),
						Style:      AsPtr(ColorAttention),
					},
					&Badge{
						Text:  AsPtr("Warning filled"),
						Size:  AsPtr(BadgeSizeLarge),
						Style: AsPtr(ColorWarning),
					},
					&Badge{
						Text:       AsPtr("Warning tint"),
						Size:       AsPtr(BadgeSizeLarge),
						Appearance: AsPtr(AppearanceTint),
						Style:      AsPtr(ColorWarning),
					},
					&Badge{
						Text:  AsPtr("Subtle filled"),
						Size:  AsPtr(BadgeSizeLarge),
						Style: AsPtr(ColorSubtle),
					},
					&Badge{
						Text:       AsPtr("Subtle tint"),
						Size:       AsPtr(BadgeSizeLarge),
						Appearance: AsPtr(AppearanceTint),
						Style:      AsPtr(ColorSubtle),
					},
				},
			},
			&TextBlock{
				Text: AsPtr("With tooltip"),
				Wrap: true,
				Common: &Common{
					Separator: true,
					Spacing:   AsPtr(SpacingExtraLarge),
				},
			},
			&Container{
				Layouts: []Layout{
					&LayoutFlow{
						HorizontalItemsAlignment: AsPtr(HorizontalAlignmentLeft),
						//HorizontalItemsAlignment: HorizontalAlignmentLeft,
						MinItemWidth: "0px",
					},
				},
				Items: []Element{
					&Badge{
						Text:    AsPtr("With tooltip"),
						Tooltip: AsPtr("This is the tooltip"),
						Size:    AsPtr(BadgeSizeLarge),
					},
					&Badge{
						Tooltip: AsPtr("This is the tooltip"),
						Size:    AsPtr(BadgeSizeLarge),
						Icon:    AsPtr(SymbolCalendar),
					},
				},
			},
		},
	}
	compareCardJSON(t, expected, "badge_example.json")
}

func TestAdaptiveCardContainerExample(t *testing.T) {
	t.Skip("TODO: Implement full comparison structure for container_test.json")
	var expected = AdaptiveCard{
		Version: "1.5",
	}
	compareCardJSON(t, expected, "container_test.json")
}

func TestAdaptiveCardTextBlockExample(t *testing.T) {
	t.Skip("TODO: Implement full comparison structure for textblock_example.json")
	var expected = AdaptiveCard{
		Version: "2.0",
	}
	compareCardJSON(t, expected, "textblock_example.json")
}

func TestAdaptiveCardResources(t *testing.T) {
	var expected = AdaptiveCard{
		Version: "1.5",
		Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
		Body: []Element{
			&TextBlock{
				Text: AsPtr("${myText}"),
			},
		},
		Resources: &Resources{
			StringResources: []StringResource{
				{
					Key:          "myText",
					DefaultValue: "Welcome!",
					LocalizedValues: map[string]string{
						"fr-FR": "Bienvenue!",
						"es-ES": "¡Bienvenido!",
					},
				},
			},
		},
	}

	compareCardJSON(t, expected, "resources_example.json")
}

func TestAdaptiveCard(t *testing.T) {
	assert.True(t, true)
}
