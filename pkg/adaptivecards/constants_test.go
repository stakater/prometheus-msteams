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
)

func TestConstantsMarshalingTextBlock(t *testing.T) {
	tb := TextBlock{
		Text:   AsPtr("Hello World"),
		Size:   AsPtr(FontSizeLarge),
		Weight: AsPtr(FontWeightBolder),
		Color:  AsPtr(ColorAccent),
		Common: &Common{
			Spacing: AsPtr(SpacingSmall),
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(tb)
	assert.NoError(t, err)

	// Unmarshal back
	var tb2 TextBlock
	err = json.Unmarshal(data, &tb2)
	assert.NoError(t, err)

	// Verify values
	assert.NotNil(t, tb2.Text)
	assert.Equal(t, "Hello World", *tb2.Text)
	assert.NotNil(t, tb2.Size)
	assert.Equal(t, FontSizeLarge, *tb2.Size)
	assert.NotNil(t, tb2.Weight)
	assert.Equal(t, FontWeightBolder, *tb2.Weight)
	assert.NotNil(t, tb2.Color)
	assert.Equal(t, ColorAccent, *tb2.Color)
	assert.NotNil(t, tb2.Spacing)
	assert.Equal(t, SpacingSmall, *tb2.Spacing)
}

func TestConstantsMarshalingImage(t *testing.T) {
	img := Image{
		URL:   "https://example.com/image.png",
		Size:  AsPtr(ImageSizeMedium),
		Style: AsPtr(ImageStylePerson),
		Common: &Common{
			Spacing: AsPtr(SpacingLarge),
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(img)
	assert.NoError(t, err)

	// Unmarshal back
	var img2 Image
	err = json.Unmarshal(data, &img2)
	assert.NoError(t, err)

	// Verify values
	assert.Equal(t, "https://example.com/image.png", img2.URL)
	assert.NotNil(t, img2.Size)
	assert.Equal(t, ImageSizeMedium, *img2.Size)
	assert.NotNil(t, img2.Style)
	assert.Equal(t, ImageStylePerson, *img2.Style)
	assert.NotNil(t, img2.Spacing)
	assert.Equal(t, SpacingLarge, *img2.Spacing)
}

func TestConstantsMarshalingActionExecute(t *testing.T) {
	action := ActionExecute{
		CommonActionProperties: &CommonActionProperties{
			Title: "Submit",
			Style: AsPtr(ActionStylePositive),
			Mode:  AsPtr(ActionModePrimary),
		},
		AssociatedInputs: AsPtr(AssociatedInputsAuto),
	}

	// Marshal to JSON
	data, err := json.Marshal(action)
	assert.NoError(t, err)

	// Unmarshal back
	var action2 ActionExecute
	err = json.Unmarshal(data, &action2)
	assert.NoError(t, err)

	// Verify values
	assert.Equal(t, "Submit", action2.Title)
	assert.NotNil(t, action2.Style)
	assert.Equal(t, ActionStylePositive, *action2.Style)
	assert.NotNil(t, action2.Mode)
	assert.Equal(t, ActionModePrimary, *action2.Mode)
	assert.NotNil(t, action2.AssociatedInputs)
	assert.Equal(t, AssociatedInputsAuto, *action2.AssociatedInputs)
}

func TestConstantsMarshalingContainer(t *testing.T) {
	container := Container{
		Style:                    AsPtr(ContainerStyleEmphasis),
		VerticalContentAlignment: AsPtr(VerticalAlignmentCenter),
		Common: &Common{
			Height:  AsPtr(BlockElementHeightStretch),
			Spacing: AsPtr(SpacingMedium),
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(container)
	assert.NoError(t, err)

	// Unmarshal back
	var container2 Container
	err = json.Unmarshal(data, &container2)
	assert.NoError(t, err)

	// Verify values
	assert.NotNil(t, container2.Style)
	assert.Equal(t, ContainerStyleEmphasis, *container2.Style)
	assert.NotNil(t, container2.VerticalContentAlignment)
	assert.Equal(t, VerticalAlignmentCenter, *container2.VerticalContentAlignment)
	assert.NotNil(t, container2.Height)
	assert.Equal(t, BlockElementHeightStretch, *container2.Height)
	assert.NotNil(t, container2.Spacing)
	assert.Equal(t, SpacingMedium, *container2.Spacing)
}

func TestConstantsMarshalingInputText(t *testing.T) {
	inputText := InputText{
		Common: &Common{
			ID:      "name",
			Spacing: AsPtr(SpacingNone),
		},
		Label:         "Name",
		Style:         AsPtr(TextInputStyleText),
		LabelPosition: AsPtr(LabelPositionInline),
	}

	// Marshal to JSON
	data, err := json.Marshal(inputText)
	assert.NoError(t, err)

	// Unmarshal back
	var inputText2 InputText
	err = json.Unmarshal(data, &inputText2)
	assert.NoError(t, err)

	// Verify values
	assert.Equal(t, "name", inputText2.ID)
	assert.Equal(t, "Name", inputText2.Label)
	assert.NotNil(t, inputText2.Style)
	assert.Equal(t, TextInputStyleText, *inputText2.Style)
	assert.NotNil(t, inputText2.LabelPosition)
	assert.Equal(t, LabelPositionInline, *inputText2.LabelPosition)
	assert.NotNil(t, inputText2.Spacing)
	assert.Equal(t, SpacingNone, *inputText2.Spacing)
}

func TestConstantsMarshalingBackgroundImage(t *testing.T) {
	bg := BackgroundImage{
		URL:                 "https://example.com/bg.png",
		FillMode:            AsPtr(ImageFillModeCover),
		HorizontalAlignment: AsPtr(HorizontalAlignmentCenter),
		VerticalAlignment:   AsPtr(VerticalAlignmentTop),
	}

	// Marshal to JSON
	data, err := json.Marshal(bg)
	assert.NoError(t, err)

	// Unmarshal back
	var bg2 BackgroundImage
	err = json.Unmarshal(data, &bg2)
	assert.NoError(t, err)

	// Verify values
	assert.Equal(t, "https://example.com/bg.png", bg2.URL)
	assert.NotNil(t, bg2.FillMode)
	assert.Equal(t, ImageFillModeCover, *bg2.FillMode)
	assert.NotNil(t, bg2.HorizontalAlignment)
	assert.Equal(t, HorizontalAlignmentCenter, *bg2.HorizontalAlignment)
	assert.NotNil(t, bg2.VerticalAlignment)
	assert.Equal(t, VerticalAlignmentTop, *bg2.VerticalAlignment)
}
