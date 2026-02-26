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

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/stakater/prometheus-msteams/pkg/utility"
)

const (
	// DefaultVersion is the default Adaptive Card version to use if not specified.
	// This can be overridden by setting the Version field on an AdaptiveCard.
	DefaultVersion = "1.6"
	// DefaultSchema is the default schema URL for Adaptive Cards.
	DefaultSchema = "https://adaptivecards.io/schemas/adaptive-card.json"
)

var logger *utility.Logger

// init registers all known Element and Action types for dynamic unmarshaling
func init() {
	typeRegistry["AdaptiveCard"] = reflect.TypeFor[AdaptiveCard]()

	// Register all Element types
	RegisterType("ActionSet", ActionSet{})
	RegisterType("Badge", Badge{})
	RegisterType("CodeBlock", CodeBlock{})
	RegisterType("Container", Container{})
	RegisterType("ColumnSet", ColumnSet{})
	RegisterType("Column", Column{})
	RegisterType("FactSet", FactSet{})
	RegisterType("Icon", Icon{})
	RegisterType("Image", Image{})
	RegisterType("ImageSet", ImageSet{})
	RegisterType("Media", Media{})
	RegisterType("Rating", Rating{})
	RegisterType("RichTextBlock", RichTextBlock{})
	RegisterType("Table", Table{})
	RegisterType("TextBlock", TextBlock{})

	// Register all Input types
	RegisterType("Input.ChoiceSet", InputChoiceSet{})
	RegisterType("Input.Date", InputDate{})
	RegisterType("Input.Number", InputNumber{})
	RegisterType("Input.Rating", InputRating{})
	RegisterType("Input.Text", InputText{})
	RegisterType("Input.Time", InputTime{})
	RegisterType("Input.Toggle", InputToggle{})

	// Register all RichTextInline types
	RegisterType("TextRun", TextRun{})
	RegisterType("CitationRun", CitationRun{})
	RegisterType("IconRun", IconRun{})
	RegisterType("ImageRun", ImageRun{})

	// Register all Action types
	RegisterType("Action.Execute", ActionExecute{})
	RegisterType("Action.InsertImage", ActionInsertImage{})
	RegisterType("Action.OpenUrl", ActionOpenURL{})
	RegisterType("Action.OpenUrlDialog", ActionOpenURLDialog{})
	RegisterType("Action.Popover", ActionPopover{})
	RegisterType("Action.ResetInputs", ActionResetInputs{})
	RegisterType("Action.RunCommands", ActionRunCommands{})
	RegisterType("Action.ShowCard", ActionShowCard{})
	RegisterType("Action.Submit", ActionSubmit{})
	RegisterType("Action.ToggleVisibility", ActionToggleVisibility{})

	RegisterType("imBack", ImBackSubmitActionData{})
	RegisterType("messageBack", MessageBackSubmitActionData{})
	RegisterType("invoke", InvokeSubmitActionData{})
	RegisterType("task/fetch", TaskFetchSubmitActionData{})
	RegisterType("signin", SigninSubmitActionData{})

	// Register all Reference types
	RegisterType("AdaptiveCardReference", AdaptiveCardReference{})
	RegisterType("DocumentReference", DocumentReference{})

	// Register all Layout types
	RegisterType("Layout.Stack", LayoutStack{})
	RegisterType("Layout.Flow", LayoutFlow{})
	RegisterType("Layout.AreaGrid", LayoutAreaGrid{})
	RegisterType("StringResource", StringResource{})

	if logger == nil {
		logger = utility.NewLogger(utility.LogFormatJSON, true)
	}
}

// AsPtr is a helper function that takes a value of any type and returns a
// pointer to that value.
func AsPtr[T any](b T) *T { return &b }

// Common properties shared by multiple elements
type Common struct {
	Fallback            any                  `json:"fallback,omitempty" version:"1.2"`
	GridArea            string               `json:"grid.area,omitempty" version:"1.5"`
	Height              *BlockElementHeight  `json:"height,omitempty" version:"1.1"`
	HorizontalAlignment *HorizontalAlignment `json:"horizontalAlignment,omitempty" version:"1.0"`
	ID                  string               `json:"id,omitempty" version:"1.0"`
	IsSortKey           bool                 `json:"isSortKey,omitempty" version:"1.5"`
	IsVisible           bool                 `json:"isVisible,omitempty" version:"1.2"`
	IsVisibleDynamic    bool                 `json:"isVisible.dynamic,omitempty" version:"1.5"`
	Key                 string               `json:"key,omitempty" version:"1.0"`
	Lang                string               `json:"lang,omitempty" version:"1.1"`
	Requires            *Requires            `json:"requires,omitempty" version:"1.2"`
	Separator           bool                 `json:"separator,omitempty" version:"1.0"`
	Spacing             *Spacing             `json:"spacing,omitempty" version:"1.0"`
	TargetWidth         *TargetWidth         `json:"targetWidth,omitempty" version:"1.0"`
}

// CommonActionProperties are properties shared by all Action types
type CommonActionProperties struct {
	Fallback         any               `json:"fallback,omitempty" version:"1.2"`
	IconURL          string            `json:"iconUrl,omitempty" version:"1.1"`
	ID               string            `json:"id,omitempty"`
	IsEnabled        *bool             `json:"isEnabled,omitempty" version:"1.5"`
	IsEnabledDynamic *bool             `json:"isEnabled.dynamic,omitempty" version:"1.5"`
	IsVisible        *bool             `json:"isVisible,omitempty" version:"1.5"`
	IsVisibleDynamic *bool             `json:"isVisible.dynamic,omitempty" version:"1.5"`
	Key              string            `json:"key,omitempty" version:"1.0"`
	MenuActions      []Action          `json:"menuActions,omitempty" version:"1.5"`
	Mode             *ActionMode       `json:"mode,omitempty" version:"1.5"`
	Requires         map[string]string `json:"requires,omitempty" version:"1.2"`
	Style            *ActionStyle      `json:"style,omitempty" version:"1.2"`
	ThemedIconURLs   []ThemedURL       `json:"themedIconUrls,omitempty" version:"1.5"`
	Title            string            `json:"title,omitempty"`
	TitleDynamic     string            `json:"title.dynamic,omitempty" version:"1.5"`
	Tooltip          string            `json:"tooltip,omitempty" version:"1.5"`
	TooltipDynamic   string            `json:"tooltip.dynamic,omitempty" version:"1.5"`
}

// region AdaptiveCard

// AdaptiveCard - https://adaptivecards.io/explorer/AdaptiveCard.html
type AdaptiveCard struct {
	Schema                   string               `json:"$schema,omitempty"`
	Actions                  []Action             `json:"actions,omitempty"`
	Authentication           *Authentication      `json:"authentication,omitempty" version:"1.4"`
	Body                     []Element            `json:"body,omitempty"`
	FallbackText             string               `json:"fallbackText,omitempty"`
	Metadata                 *Metadata            `json:"metadata,omitempty" version:"1.4"`
	MsTeams                  *TeamsCardProperties `json:"msteams,omitempty" version:"1.0"`
	References               []Reference          `json:"references,omitempty" version:"1.5"`
	Refresh                  *Refresh             `json:"refresh,omitempty" version:"1.4"`
	Resources                *Resources           `json:"resources,omitempty" version:"1.5"`
	Speak                    string               `json:"speak,omitempty"`
	Version                  string               `json:"version"`
	BackgroundImage          *BackgroundImage     `json:"backgroundImage,omitempty" version:"1.2"`
	Fallback                 any                  `json:"fallback,omitempty" version:"1.2"`
	GridArea                 *LayoutAreaGrid      `json:"grid.area,omitempty" version:"1.5"`
	ID                       string               `json:"id,omitempty"`
	IsSortKey                bool                 `json:"isSortKey,omitempty" version:"1.5"`
	Key                      string               `json:"key,omitempty"`
	Lang                     string               `json:"lang,omitempty" version:"1.1"`
	Layouts                  *Layout              `json:"layouts,omitempty" version:"1.5"`
	MinHeight                string               `json:"minHeight,omitempty" version:"1.2"`
	Requires                 *Requires            `json:"requires,omitempty" version:"1.2"`
	RTL                      *bool                `json:"rtl,omitempty" version:"1.5"`
	SelectAction             Action               `json:"selectAction,omitempty" version:"1.1"`
	Style                    *ContainerStyle      `json:"style,omitempty"`
	VerticalContentAlignment *VerticalAlignment   `json:"verticalContentAlignment,omitempty" version:"1.1"`
}

// MarshalJSON automatically injects "type": "AdaptiveCard"
func (a AdaptiveCard) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "AdaptiveCard")
}

// UnmarshalJSON ensures we only unmarshal if the type is correct
func (a *AdaptiveCard) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// Validate checks if the card's features are compatible with its declared version.
// Returns a slice of ValidationErrors if any incompatibilities are found.
//
// Example:
//
//	card := adaptivecards.AdaptiveCard{Version: "1.0", RTL: boolPtr(true)}
//	if errors := card.Validate(); len(errors) > 0 {
//	    for _, err := range errors {
//	        log.Printf("Validation error: %v", err)
//	    }
//	}
func (a *AdaptiveCard) Validate() []ValidationError {
	return ValidateVersion(*a, a.Version)
}

// endregion AdaptiveCard

// region Reference

// Reference is an interface for all adaptive card references
type Reference interface {
	isReference()
}

// region AdaptiveCardReference

// AdaptiveCardReference represents a reference to another Adaptive Card,
// allowing for modular card design and reuse of card content across multiple cards.
type AdaptiveCardReference struct {
	*DocumentReference
	Content *any `json:"content,omitempty" version:"1.5"` // Must be an AdaptiveCard payload
}

func (c AdaptiveCardReference) isReference() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// AdaptiveCardReference to JSON.
func (c AdaptiveCardReference) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "AdaptiveCardReference")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an AdaptiveCardReference struct.
func (c *AdaptiveCardReference) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// endregion AdaptiveCardReference

// region DocumentReference

// DocumentReference represents a reference to an external document,
// which can be used in Adaptive Cards to link to additional content or resources.
type DocumentReference struct {
	Abstract string   `json:"abstract,omitempty" version:"1.5"`
	Icon     *Symbol  `json:"icon,omitempty" version:"1.5"`
	Key      string   `json:"key,omitempty"`
	Keywords []string `json:"keywords,omitempty" version:"1.5"` // Max allowed is 3
	Title    string   `json:"title,omitempty" version:"1.5"`
	URL      string   `json:"url,omitempty" version:"1.5"`
}

func (c DocumentReference) isReference() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// DocumentReference to JSON.
func (c DocumentReference) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "DocumentReference")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a DocumentReference struct.
func (c *DocumentReference) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// endregion DocumentReference

// endregion Reference

// region Elements

// Element is an interface for all adaptive card elements
type Element interface {
	isElement()
}

// region ActionSet

// ActionSet - https://adaptivecards.io/explorer/ActionSet.html
type ActionSet struct {
	*Common
	Actions []Action `json:"actions"`
}

func (a ActionSet) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// ActionSet to JSON.
func (a ActionSet) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "ActionSet")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an ActionSet struct.
func (a *ActionSet) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion ActionSet

// region Badge

// Badge represents a badge element in an Adaptive Card, which can display a small piece of information or status.
type Badge struct {
	*Common
	Appearance   *Appearance   `json:"appearance,omitempty" version:"1.5"`
	Icon         *Symbol       `json:"icon,omitempty" version:"1.5"`
	IconPosition *IconPosition `json:"iconPosition,omitempty" version:"1.5"`
	Shape        *BadgeShape   `json:"shape,omitempty" version:"1.5"`
	Size         *BadgeSize    `json:"size,omitempty" version:"1.5"`
	Style        *Colors       `json:"style,omitempty" version:"1.5"`
	Text         *string       `json:"text,omitempty" version:"1.5"`
	Tooltip      *string       `json:"tooltip,omitempty" version:"1.5"`
}

func (c Badge) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling a Badge to JSON.
func (c Badge) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "Badge")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling JSON into a Badge struct.
func (c *Badge) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// endregion Badge

// region CodeBlock

// CodeBlock represents a code block element in an Adaptive Card, which can display a snippet of code with optional syntax highlighting.
type CodeBlock struct {
	*Common
	CodeSnippet     string        `json:"codeSnippet,omitempty" version:"1.5"`
	Language        *CodeLanguage `json:"language,omitempty" version:"1.5"`
	StartLineNumber int           `json:"startLineNumber,omitempty" version:"1.5"`
}

func (c CodeBlock) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling a CodeBlock to JSON.
func (c CodeBlock) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "CodeBlock")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling JSON into a CodeBlock struct.
func (c *CodeBlock) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// endregion CodeBlock

// region ColumnSet

// ColumnSet - https://adaptivecards.io/explorer/ColumnSet.html
type ColumnSet struct {
	*Common
	Columns        []Column `json:"columns,omitempty"`
	MinWidth       string   `json:"minWidth,omitempty" version:"1.5"`
	Bleed          bool     `json:"bleed,omitempty" version:"1.2"`
	MinHeight      string   `json:"minHeight,omitempty" version:"1.2"`
	RoundedCorners bool     `json:"roundedCorners,omitempty" version:"1.5"`
	ShowBorder     bool     `json:"showBorder,omitempty" version:"1.5"`
	SelectAction   Action   `json:"selectAction,omitempty" version:"1.1"`
}

func (c ColumnSet) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// ColumnSet to JSON.
func (c ColumnSet) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "ColumnSet")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a ColumnSet struct.
func (c *ColumnSet) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// region Column

// Column - https://adaptivecards.io/explorer/Column.html
type Column struct {
	*Common
	Items                    []Element          `json:"items,omitempty"`
	Width                    any                `json:"width,omitempty"` // can be string or number
	BackgroundImage          *BackgroundImage   `json:"backgroundImage,omitempty"`
	Bleed                    bool               `json:"bleed,omitempty"`
	Layout                   *Layout            `json:"layout,omitempty" version:"1.5"`
	MaxHeight                string             `json:"maxHeight,omitempty"`
	MinHeight                string             `json:"minHeight,omitempty"`
	RoundedCorners           bool               `json:"roundedCorners,omitempty" version:"1.5"`
	RTL                      *bool              `json:"rtl,omitempty"`
	SelectAction             Action             `json:"selectAction,omitempty"`
	ShowBorder               bool               `json:"showBorder,omitempty" version:"1.5"`
	Style                    *ContainerStyle    `json:"style,omitempty"`
	VerticalContentAlignment *VerticalAlignment `json:"verticalContentAlignment,omitempty"`
}

func (c Column) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// Column to JSON.
func (c Column) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "Column")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a Column struct.
func (c *Column) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// endregion Column

// endregion ColumnSet

// region CompoundButton

// CompoundButton represents a compound button element in an Adaptive Card,
// which can display a button with multiple pieces of information and an optional icon.
type CompoundButton struct {
	*Common
	Badge        *Badge  `json:"badge,omitempty" version:"1.5"`
	Description  string  `json:"description,omitempty" version:"1.5"`
	Icon         *Symbol `json:"icon,omitempty" version:"1.5"`
	SelectAction Action  `json:"selectAction,omitempty" version:"1.1"`
	Title        string  `json:"title,omitempty" version:"1.5"`
}

func (c CompoundButton) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// CompoundButton to JSON.
func (c CompoundButton) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "CompoundButton")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a CompoundButton struct.
func (c *CompoundButton) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// endregion CompoundButton

// region Container

// Container - https://adaptivecards.io/explorer/Container.html
type Container struct {
	*Common
	BackgroundImage          *BackgroundImage   `json:"backgroundImage,omitempty" version:"1.2"`
	Items                    []Element          `json:"items"`
	MaxHeight                string             `json:"maxHeight,omitempty" version:"1.5"`
	MinHeight                string             `json:"minHeight,omitempty" version:"1.5"`
	RTL                      *bool              `json:"rtl,omitempty" version:"1.5"`
	VerticalContentAlignment *VerticalAlignment `json:"verticalContentAlignment,omitempty" version:"1.1"`
	Bleed                    bool               `json:"bleed,omitempty" version:"1.2"`
	Layouts                  []Layout           `json:"layouts,omitempty" version:"1.5"`
	RoundedCorners           bool               `json:"roundedCorners,omitempty" version:"1.5"`
	SelectAction             Action             `json:"selectAction,omitempty" version:"1.1"`
	ShowBorder               bool               `json:"showBorder,omitempty" version:"1.5"`
	Style                    *ContainerStyle    `json:"style,omitempty"`
}

func (c Container) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// Container to JSON.
func (c Container) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "Container")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a Container struct.
func (c *Container) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// endregion Container

// region FactSet

// FactSet - https://adaptivecards.io/explorer/FactSet.html
type FactSet struct {
	*Common
	Facts []Fact `json:"facts"`
}

func (f FactSet) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// FactSet to JSON.
func (f FactSet) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(f)
	// return marshalWithType(f, "FactSet")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a FactSet struct.
func (f *FactSet) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, f)
}

// region Fact

// Fact describes a Fact in a FactSet as a key/value pair
type Fact struct {
	Title string `json:"title" version:"1.0"`
	Value string `json:"value" version:"1.0"`
	Key   string `json:"key,omitempty" version:"1.0"`
}

/*
// MarshalJSON ensures that the "type" field is included when marshaling a
// Fact to JSON.
func (f Fact) MarshalJSON() ([]byte, error) {
	return json.Marshal(f)
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a Fact struct.
func (f *Fact) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, f)
}
*/

// endregion Fact

// endregion FactSet

// region Icon

// Icon - https://adaptivecards.io/explorer/Icon.html
type Icon struct {
	*Common
	URL          string     `json:"url"`
	Color        Colors     `json:"color,omitempty" version:"1.5"`
	Name         *Symbol    `json:"name,omitempty" version:"1.5"`
	SelectAction Action     `json:"selectAction,omitempty" version:"1.5"`
	Size         *IconSize  `json:"size,omitempty" version:"1.5"`
	Style        *IconStyle `json:"style,omitempty" version:"1.5"`
}

func (i Icon) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// Icon to JSON.
func (i Icon) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(i)
	// return marshalWithType(i, "Icon")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an Icon struct.
func (i *Icon) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, i)
}

// endregion Icon

// region Image

// Image - https://adaptivecards.io/explorer/Image.html
type Image struct {
	*Common
	AllowExpand                bool                  `json:"allowExpand,omitempty" verson:"1.2"`
	AltText                    string                `json:"altText,omitempty" version:"1.0"`
	BackgroundColor            string                `json:"backgroundColor,omitempty" version:"1.1"`
	FitMode                    *ImageFit             `json:"fitMode,omitempty" version:"1.2"`
	Height                     any                   `json:"height,omitempty"` // can be string or BlockElementHeight
	HorizontalContentAlignment *HorizontalAlignment  `json:"horizontalContentAlignment,omitempty"`
	MSTeams                    *TeamsImageProperties `json:"msteams,omitempty" version:"1.2"`
	SelectAction               Action                `json:"selectAction,omitempty"`
	Size                       *ImageSize            `json:"size,omitempty"`
	Style                      *ImageStyle           `json:"style,omitempty"`
	URL                        string                `json:"url"`
	VerticalContentAlignment   *HorizontalAlignment  `json:"verticalContentAlignment,omitempty"`
	Width                      string                `json:"width,omitempty"`
}

func (i Image) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// Image to JSON.
func (i Image) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(i)
	// return marshalWithType(i, "Image")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an Image struct.
func (i *Image) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, i)
}

// region TeamsImageProperties

// TeamsImageProperties defines Microsoft Teams-specific properties for
// Image elements in Adaptive Cards.
type TeamsImageProperties struct {
	AllowExpand bool   `json:"allowExpand,omitempty" verson:"1.2"`
	Key         string `json:"key,omitempty" version:"1.0"`
}

// endregion TeamsImageProperties

// endregion Image

// region ImageSet

// ImageSet - https://adaptivecards.io/explorer/ImageSet.html
type ImageSet struct {
	*Common
	Images    []Image    `json:"images"`
	ImageSize *ImageSize `json:"imageSize,omitempty"`
}

func (i ImageSet) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// ImageSet to JSON.
func (i ImageSet) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(i)
	// return marshalWithType(i, "ImageSet")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an ImageSet struct.
func (i *ImageSet) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, i)
}

// endregion ImageSet

// region Media

// Media - https://adaptivecards.io/explorer/Media.html
type Media struct {
	*Common
	AltText        string          `json:"altText,omitempty"`
	CaptionSources []CaptionSource `json:"captionSources,omitempty" version:"1.6"`
	Poster         string          `json:"poster,omitempty"`
	Sources        []MediaSource   `json:"sources"`
}

func (m Media) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// Media to JSON.
func (m Media) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(m)
	// return marshalWithType(m, "Media")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a Media struct.
func (m *Media) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, m)
}

// region CaptionSource

// CaptionSource defines a source for captions (version 1.6)
type CaptionSource struct {
	MimeType string `json:"mimeType"`
	URL      string `json:"url"`
	Label    string `json:"label"`
}

// MarshalJSON ensures that the "type" field is included when marshaling a
// CaptionSource to JSON.
func (c CaptionSource) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "CaptionSource")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a CaptionSource struct.
func (c *CaptionSource) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// endregion CaptionSource

// region MediaSource

// MediaSource defines a source for a Media element
type MediaSource struct {
	MimeType string `json:"mimeType"`
	URL      string `json:"url"`
}

// MarshalJSON ensures that the "type" field is included when marshaling a
// MediaSource to JSON.
func (m MediaSource) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(m)
	// return marshalWithType(m, "MediaSource")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a MediaSource struct.
func (m *MediaSource) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, m)
}

// endregion MediaSource

// endregion Media

// region ProgressBar

// ProgressBar represents a progress bar element in an Adaptive Card,
// which can visually indicate the progress of a task or process.
type ProgressBar struct {
	*Common
	Color *Colors `json:"color,omitempty" version:"1.5"`
	Max   int     `json:"max,omitempty" version:"1.5"`
	Value int     `json:"value,omitempty" version:"1.5"`
}

func (c ProgressBar) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// ProgressBar to JSON.
func (c ProgressBar) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "ProgressBar")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a ProgressBar struct.
func (c *ProgressBar) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// endregion ProgressBar

// region ProgressRing

// ProgressRing represents a progress ring element in an Adaptive Card,
// which can visually indicate the progress of a task or process in a
// circular format.
type ProgressRing struct {
	*Common
	Label         string                     `json:"label,omitempty" version:"1.5"`
	LabelPosition *ProgressRingLabelPosition `json:"labelPosition,omitempty" version:"1.5"`
	Size          *ProgressRingSize          `json:"size,omitempty" version:"1.5"`
}

func (c ProgressRing) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// ProgressRing to JSON.
func (c ProgressRing) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "ProgressRing")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a ProgressRing struct.
func (c *ProgressRing) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// endregion ProgressRing

// region Rating

// Rating represents a rating element in an Adaptive Card, which can display a
// rating value using stars or other symbols.
type Rating struct {
	*Common
	Color *RatingColor `json:"color,omitempty" version:"1.5"`
	Count int          `json:"count,omitempty" version:"1.5"`
	Max   int          `json:"max,omitempty" version:"1.5"`
	Size  *RatingSize  `json:"size,omitempty" version:"1.5"`
	Style *RatingStyle `json:"style,omitempty" version:"1.5"`
	Value int          `json:"value,omitempty" version:"1.5"`
}

func (c Rating) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// Rating to JSON.
func (c Rating) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "Rating")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a Rating struct.
func (c *Rating) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// endregion Rating

// region RichTextBlock

// RichTextInline represents a block of rich text in an Adaptive Card,
// which can contain multiple inlines with different formatting and styles.
type RichTextInline interface {
	isRichTextInline()
}

// RichTextBlock - https://adaptivecards.io/explorer/RichTextBlock.html
type RichTextBlock struct {
	*Common
	Inlines  []RichTextInline `json:"inlines"`
	LabelFor string           `json:"labelFor,omitempty" version:"1.5"`
}

func (r RichTextBlock) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// RichTextBlock to JSON.
func (r RichTextBlock) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(r)
	// return marshalWithType(r, "RichTextBlock")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a RichTextBlock struct.
func (r *RichTextBlock) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, r)
}

// region TextRun

// TextRun - https://adaptivecards.io/explorer/TextRun.html
type TextRun struct {
	*Common
	Highlight     bool        `json:"highlight,omitempty"`
	Italic        bool        `json:"italic,omitempty"`
	Strikethrough bool        `json:"strikethrough,omitempty"`
	Underline     bool        `json:"underline,omitempty"`
	Color         *Colors     `json:"color,omitempty"`
	FontType      *FontType   `json:"fontType,omitempty"`
	IsSubtle      *bool       `json:"isSubtle,omitempty"`
	Size          *FontSize   `json:"size,omitempty"`
	Text          string      `json:"text"`
	TextDynamic   string      `json:"text.dynamic,omitempty" version:"1.5"`
	SelectAction  Action      `json:"selectAction,omitempty" version:"1.5"`
	Weight        *FontWeight `json:"weight,omitempty"`
}

func (t TextRun) isRichTextInline() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// TextRun to JSON.
func (t TextRun) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(t)
	// return marshalWithType(t, "TextRun")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a TextRun struct.
func (t *TextRun) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, t)
}

// endregion TextRun

// region CitationRun

// CitationRun - https://adaptivecards.io/explorer/CitationRun.html
type CitationRun struct {
	ReferenceIndex int    `json:"referenceIndex,omitempty" version:"1.5"` // Index starts at 1
	Text           string `json:"text"`
	TextDynamic    string `json:"text.dynamic,omitempty" version:"1.5"`

	Fallback         any    `json:"fallback,omitempty" version:"1.2"`
	GridArea         string `json:"grid.area,omitempty" version:"1.5"`
	ID               string `json:"id,omitempty" version:"1.0"`
	IsVisibleDynamic bool   `json:"isVisible.dynamic,omitempty" version:"1.5"`
	Key              string `json:"key,omitempty" version:"1.0"`
}

func (c CitationRun) isRichTextInline() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// CitationRun to JSON.
func (c CitationRun) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "CitationRun")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a CitationRun struct.
func (c *CitationRun) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// endregion CitationRun

// region IconRun

// IconRun - https://adaptivecards.io/explorer/IconRun.html
type IconRun struct {
	Color            *Colors `json:"color,omitempty" version:"1.5"`
	Name             *Symbol `json:"name,omitempty" version:"1.5"`
	SelectAction     Action  `json:"selectAction,omitempty" version:"1.5"`
	Fallback         any     `json:"fallback,omitempty" version:"1.2"`
	GridArea         string  `json:"grid.area,omitempty" version:"1.5"`
	ID               string  `json:"id,omitempty" version:"1.0"`
	IsSortKey        bool    `json:"isSortKey,omitempty" version:"1.5"`
	IsVisible        bool    `json:"isVisible,omitempty" version:"1.2"`
	IsVisibleDynamic bool    `json:"isVisible.dynamic,omitempty" version:"1.5"`
	Key              string  `json:"key,omitempty" version:"1.0"`
	Lang             string  `json:"lang,omitempty" version:"1.1"`
}

func (i IconRun) isRichTextInline() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// IconRun to JSON.
func (i IconRun) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(i)
	// return marshalWithType(i, "IconRun")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an IconRun struct.
func (i *IconRun) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, i)
}

// endregion IconRun

// region ImageRun

// ImageRun - https://adaptivecards.io/explorer/ImageRun.html
type ImageRun struct {
	SelectAction     Action      `json:"selectAction,omitempty" version:"1.5"`
	Size             *ImageSize  `json:"size,omitempty" version:"1.5"`
	Style            *ImageStyle `json:"style,omitempty" version:"1.5"`
	ThemedUrls       *ThemedURL  `json:"themedUrls,omitempty" version:"1.5"`
	URL              string      `json:"url,omitempty" version:"1.5"`
	Fallback         any         `json:"fallback,omitempty" version:"1.2"`
	GridArea         string      `json:"grid.area,omitempty" version:"1.5"`
	ID               string      `json:"id,omitempty" version:"1.0"`
	IsSortKey        bool        `json:"isSortKey,omitempty" version:"1.5"`
	IsVisible        bool        `json:"isVisible,omitempty" version:"1.2"`
	IsVisibleDynamic bool        `json:"isVisible.dynamic,omitempty" version:"1.5"`
	Key              string      `json:"key,omitempty" version:"1.0"`
	Lang             string      `json:"lang,omitempty" version:"1.1"`
}

func (i ImageRun) isRichTextInline() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// ImageRun to JSON.
func (i ImageRun) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(i)
	// return marshalWithType(i, "ImageRun")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an ImageRun struct.
func (i *ImageRun) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, i)
}

// endregion ImageRun

// endregion RichTextBlock

// region Table

// Table - https://adaptivecards.io/explorer/Table.html
// Requires version 1.5+
type Table struct {
	*Common
	Columns                        []TableColumnDefinition `json:"columns,omitempty"`
	FirstRowAsHeader               bool                    `json:"firstRowAsHeader,omitempty"`
	GridStyle                      *ContainerStyle         `json:"gridStyle,omitempty"`
	HorizontalCellContentAlignment *HorizontalAlignment    `json:"horizontalCellContentAlignment,omitempty"`
	MinWidth                       string                  `json:"minWidth,omitempty"`
	Rows                           []TableRow              `json:"rows,omitempty"`
	ShowGridLines                  bool                    `json:"showGridLines,omitempty"`
	VerticalCellContentAlignment   *VerticalAlignment      `json:"verticalCellContentAlignment,omitempty"`
	RoundedCorners                 bool                    `json:"roundedCorners,omitempty"`
	ShowBorder                     bool                    `json:"showBorder,omitempty"`
}

func (t Table) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// Table to JSON.
func (t Table) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(t)
	// return marshalWithType(t, "Table")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a Table struct.
func (t *Table) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, t)
}

// region TableColumnDefinition

// TableColumnDefinition defines the characteristics of a column in a Table element
type TableColumnDefinition struct {
	Width                          any                  `json:"width,omitempty"` // can be string or number
	HorizontalCellContentAlignment *HorizontalAlignment `json:"horizontalCellContentAlignment,omitempty"`
	VerticalCellContentAlignment   *VerticalAlignment   `json:"verticalCellContentAlignment,omitempty"`
	Key                            string               `json:"key,omitempty" version:"1.0"`
}

// MarshalJSON ensures that the "type" field is included when marshaling a
// TableColumnDefinition to JSON.
func (t TableColumnDefinition) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(t)
	// return marshalWithType(t, "TableColumnDefinition")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a TableColumnDefinition struct.
func (t *TableColumnDefinition) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, t)
}

// endregion TableColumnDefinition

// region TableRow

// TableRow represents a row of cells within a Table element
type TableRow struct {
	*Common
	Cells                          []TableCell          `json:"cells,omitempty"`
	HorizontalCellContentAlignment *HorizontalAlignment `json:"horizontalCellContentAlignment,omitempty"`
	VerticalCellContentAlignment   *VerticalAlignment   `json:"verticalCellContentAlignment,omitempty"`
	Style                          *ContainerStyle      `json:"style,omitempty"`
	RoundedCorners                 bool                 `json:"roundedCorners,omitempty"`
	ShowBorder                     bool                 `json:"showBorder,omitempty"`
}

// MarshalJSON ensures that the "type" field is included when marshaling a
// TableRow to JSON.
func (t TableRow) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(t)
	// return marshalWithType(t, "TableRow")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a TableRow struct.
func (t *TableRow) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, t)
}

// endregion TableRow

// region TableCell

// TableCell represents a cell within a row of a Table element
type TableCell struct {
	*Common
	Items                    []Element          `json:"items"`
	BackgroundImage          *BackgroundImage   `json:"backgroundImage,omitempty"`
	Bleed                    bool               `json:"bleed,omitempty"`
	Layouts                  []Layout           `json:"layouts,omitempty" version:"1.5"`
	MaxHeight                string             `json:"maxHeight,omitempty"`
	MinHeight                string             `json:"minHeight,omitempty"`
	SelectAction             Action             `json:"selectAction,omitempty"`
	VerticalContentAlignment *VerticalAlignment `json:"verticalContentAlignment,omitempty"`
	Style                    *ContainerStyle    `json:"style,omitempty"`
	RTL                      *bool              `json:"rtl,omitempty"`
}

// MarshalJSON ensures that the "type" field is included when marshaling a
// TableCell to JSON.
func (t TableCell) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(t)
	// return marshalWithType(t, "TableCell")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a TableCell struct.
func (t *TableCell) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, t)
}

// endregion TableCell

// endregion Table

// region TextBlock

// TextBlock - https://adaptivecards.io/explorer/TextBlock.html
type TextBlock struct {
	*Common
	LabelFor    string          `json:"labelFor,omitempty" version:"1.5"`
	MaxLines    int             `json:"maxLines,omitempty"`
	Style       *TextBlockStyle `json:"style,omitempty" version:"1.5"`
	Wrap        bool            `json:"wrap,omitempty"`
	Color       *Colors         `json:"color,omitempty"`
	FontType    *FontType       `json:"fontType,omitempty" version:"1.2"`
	Size        *FontSize       `json:"size,omitempty"`
	Weight      *FontWeight     `json:"weight,omitempty"`
	IsSubtle    *bool           `json:"isSubtle,omitempty"`
	Text        *string         `json:"text"`
	TextDynamic *string         `json:"text.dynamic,omitempty" version:"1.5"`
}

func (t TextBlock) isElement() {}

// MarshalJSON automatically injects "type": "TextBlock"
func (t TextBlock) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(t)
	// return marshalWithType(t, "TextBlock")
}

// UnmarshalJSON ensures we only unmarshal if the type is correct
func (t *TextBlock) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, t)
}

// endregion TextBlock

// endregion Elements

// region Actions

// Action is an interface for all adaptive card actions
type Action interface {
	isAction()
}

// region ActionExecute

// ActionExecute - https://adaptivecards.io/explorer/Action.Execute.html
// Requires version 1.4+
type ActionExecute struct {
	*CommonActionProperties
	Verb                 string            `json:"verb,omitempty"`
	AssociatedInputs     *AssociatedInputs `json:"associatedInputs,omitempty"`
	ConditionallyEnabled bool              `json:"conditionallyEnabled,omitempty" version:"1.5"`
	Data                 any               `json:"data,omitempty"`
}

func (a ActionExecute) isAction() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// ActionExecute to JSON.
func (a ActionExecute) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "Action.Execute")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an ActionExecute struct.
func (a *ActionExecute) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion ActionExecute

// region ActionInsertImage

// ActionInsertImage - https://adaptivecards.io/explorer/Action.InsertImage.html
type ActionInsertImage struct {
	*CommonActionProperties
	AltText        string          `json:"altText,omitempty"`
	InsertPosition *InsertPosition `json:"insertPosition,omitempty"`
	URL            string          `json:"url"`
}

func (a ActionInsertImage) isAction() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// ActionInsertImage to JSON.
func (a ActionInsertImage) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "Action.InsertImage")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an ActionInsertImage struct.
func (a *ActionInsertImage) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion ActionInsertImage

// region ActionOpenURL

// ActionOpenURL - https://adaptivecards.io/explorer/Action.OpenUrl.html
type ActionOpenURL struct {
	*CommonActionProperties
	URL string `json:"url"`
}

func (a ActionOpenURL) isAction() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// ActionOpenURL to JSON.
func (a ActionOpenURL) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "Action.OpenUrl")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an ActionOpenURL struct.
func (a *ActionOpenURL) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion ActionOpenURL

// region ActionOpenURLDialog

// ActionOpenURLDialog - https://adaptivecards.io/explorer/Action.OpenUrlDialog.html
type ActionOpenURLDialog struct {
	*CommonActionProperties
	URL          string `json:"url"`
	DialogHeight string `json:"dialogHeight,omitempty"`
	DialogTitle  string `json:"dialogTitle,omitempty"`
	DialogWidth  string `json:"dialogWidth,omitempty"`
}

func (a ActionOpenURLDialog) isAction() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// ActionOpenURLDialog to JSON.
func (a ActionOpenURLDialog) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "Action.OpenUrlDialog")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an ActionOpenURLDialog struct.
func (a *ActionOpenURLDialog) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion ActionOpenURLDialog

// region ActionPopover

// ActionPopover - https://adaptivecards.io/explorer/Action.Popover.html
type ActionPopover struct {
	*CommonActionProperties
	Content         *Element         `json:"content,omitempty"`
	DisplayArrow    *bool            `json:"displayArrow,omitempty"`
	MaxPopoverWidth string           `json:"maxPopoverWidth,omitempty"`
	PopoverTitle    string           `json:"popoverTitle,omitempty"`
	Position        *PopoverPosition `json:"position,omitempty"`
}

func (a ActionPopover) isAction() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// ActionPopover to JSON.
func (a ActionPopover) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "Action.Popover")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an ActionPopover struct.
func (a *ActionPopover) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion ActionPopover

// region ActionResetInputs

// ActionResetInputs - https://adaptivecards.io/explorer/Action.ResetInputs.html
type ActionResetInputs struct {
	*CommonActionProperties
	TargetInputIDs []string `json:"targetInputIds,omitempty"`
}

func (a ActionResetInputs) isAction() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// ActionResetInputs to JSON.
func (a ActionResetInputs) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "Action.ResetInputs")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an ActionResetInputs struct.
func (a *ActionResetInputs) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion ActionResetInputs

// region ActionRunCommands

// ActionRunCommands - https://adaptivecards.io/explorer/Action.RunCommands.html
type ActionRunCommands struct {
	*CommonActionProperties
	Commands  []string `json:"commands,omitempty"`
	OnFailure string   `json:"onFailure,omitempty"`
}

func (a ActionRunCommands) isAction() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// ActionRunCommands to JSON.
func (a ActionRunCommands) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "Action.RunCommands")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an ActionRunCommands struct.
func (a *ActionRunCommands) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion ActionRunCommands

// region ActionShowCard

// ActionShowCard - https://adaptivecards.io/explorer/Action.ShowCard.html
type ActionShowCard struct {
	*CommonActionProperties
	Card *AdaptiveCard `json:"card,omitempty"`
}

func (a ActionShowCard) isAction() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// ActionShowCard to JSON.
func (a ActionShowCard) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "Action.ShowCard")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an ActionShowCard struct.
func (a *ActionShowCard) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion ActionShowCard

// region ActionSubmit

// ActionSubmit - https://adaptivecards.io/explorer/Action.Submit.html
type ActionSubmit struct {
	*CommonActionProperties
	MSTeams              *TeamsSubmitActionProperties `json:"msteams,omitempty" version:"1.2"`
	AssociatedInputs     *AssociatedInputs            `json:"associatedInputs,omitempty"`
	ConditionallyEnabled bool                         `json:"conditionallyEnabled,omitempty" version:"1.5"`
	Data                 any                          `json:"data,omitempty"`
}

func (a ActionSubmit) isAction() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// ActionSubmit to JSON.
func (a ActionSubmit) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "Action.Submit")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an ActionSubmit struct.
func (a *ActionSubmit) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion ActionSubmit

// region ActionToggleVisibility

// ActionToggleVisibility - https://adaptivecards.io/explorer/Action.ToggleVisibility.html
type ActionToggleVisibility struct {
	*CommonActionProperties
	TargetElements []TargetElement `json:"targetElements"`
}

func (a ActionToggleVisibility) isAction() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// ActionToggleVisibility to JSON.
func (a ActionToggleVisibility) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "Action.ToggleVisibility")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an ActionToggleVisibility struct.
func (a *ActionToggleVisibility) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion ActionToggleVisibility

// region TargetElement

// TargetElement represents an entry for Action.ToggleVisibility's targetElements property
type TargetElement struct {
	ElementID string `json:"elementId"`
	IsVisible *bool  `json:"isVisible,omitempty"`
}

// MarshalJSON allows TargetElement to be marshaled as either a simple string
// (elementId) or an object with elementId and isVisible properties.
func (t TargetElement) MarshalJSON() ([]byte, error) {
	// If IsVisible is nil, marshal as a simple string
	if t.IsVisible == nil {
		return json.Marshal(t.ElementID)
	}

	return SmartMarshalFromJSON(t)
	// return marshalWithType(t, "TargetElement")
}

// UnmarshalJSON allows TargetElement to be unmarshaled from either a simple string
// (elementId) or an object with elementId and isVisible properties.
func (t *TargetElement) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &t.ElementID) // Try unmarshaling as a simple string first
	if err == nil {
		return nil
	}
	err = json.Unmarshal(data, &t) // If that fails, try unmarshaling as an object
	if err != nil {
		return fmt.Errorf("failed to unmarshal TargetElement: %w", err)
	}
	return SmartUnmarshalJSON(data, t)
}

// endregion TargetElement

// region ActionData

// ActionData is an interface for all adaptive card action data
type ActionData interface {
	isActionData()
}

// region ImBackSubmitActionData

// ImBackSubmitActionData represents Teams-specific data in an Action.Submit
// to send an Instant Message back to the Bot.
type ImBackSubmitActionData struct {
	// The value that will be sent to the Bot.
	Value string `json:"value,omitempty"`
	Key   string `json:"key,omitempty"`
}

// MarshalJSON ensures that the "type" field is included when marshaling an
// ImBackSubmitActionData to JSON.
func (a ImBackSubmitActionData) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "imBack")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an ImBackSubmitActionData struct.
func (a *ImBackSubmitActionData) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion ImBackSubmitActionData

// region MessageBackSubmitActionData

// MessageBackSubmitActionData represents Teams-specific data in an Action.Submit
// to send a message back to the Bot.
type MessageBackSubmitActionData struct {
	// The optional text that will be displayed as a new message in the conversation, as if the end-user sent it.
	// `displayText` is not sent to the Bot.
	DisplayText string `json:"displayText,omitempty"`
	// The text that will be sent to the Bot.
	Text string `json:"text,omitempty"`
	// Optional additional value that will be sent to the Bot.
	// For instance, `value`` can encode specific context for the action,
	// such as unique identifiers or a JSON object.
	Value any    `json:"value,omitempty"`
	Key   string `json:"key,omitempty"`
}

// MarshalJSON ensures that the "type" field is included when marshaling an
// MessageBackSubmitActionData to JSON.
func (a MessageBackSubmitActionData) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "messageBack")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an MessageBackSubmitActionData struct.
func (a *MessageBackSubmitActionData) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion MessageBackSubmitActionData

// region InvokeSubmitActionData

// InvokeSubmitActionData represents Teams-specific data in an Action.Submit
// to make an Invoke request to the Bot.
type InvokeSubmitActionData struct {
	Value any    `json:"value,omitempty"` // The object to send to the Bot with the Invoke request.
	Key   string `json:"key,omitempty"`
}

// MarshalJSON ensures that the "type" field is included when marshaling an
// InvokeSubmitActionData to JSON.
func (a InvokeSubmitActionData) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "invoke")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an InvokeSubmitActionData struct.
func (a *InvokeSubmitActionData) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion InvokeSubmitActionData

// region TaskFetchSubmitActionData

// TaskFetchSubmitActionData represents Teams-specific data in an Action.Submit
// to open a task module.
type TaskFetchSubmitActionData struct {
	Key string `json:"key,omitempty"`
}

// MarshalJSON ensures that the "type" field is included when marshaling an
// TaskFetchSubmitActionData to JSON.
func (a TaskFetchSubmitActionData) MarshalJSON() ([]byte, error) {
	return marshalWithType(a, "task/fetch")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an TaskFetchSubmitActionData struct.
func (a *TaskFetchSubmitActionData) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion TaskFetchSubmitActionData

// region SigninSubmitActionData

// SigninSubmitActionData represents Teams-specific data in an Action.Submit
// to sign in a user.
type SigninSubmitActionData struct {
	Value string `json:"value,omitempty"` // The URL to redirect the end-user for signing in.
	Key   string `json:"key,omitempty"`
}

// MarshalJSON ensures that the "type" field is included when marshaling an
// SigninSubmitActionData to JSON.
func (a SigninSubmitActionData) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "signin")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an SigninSubmitActionData struct.
func (a *SigninSubmitActionData) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion SigninSubmitActionData

// endregion ActionData

// SubmitActionData represents the data of an Action.Submit.
type SubmitActionData struct {
	// Defines the optional Teams-specific portion of the action's data
	MSTeams    ActionData     `json:"msteams,omitempty" version:"1.2"`
	Key        string         `json:"key,omitempty"`
	Properties map[string]any `json:"-"`
}

// endregion Actions

// region Inputs

// region InputChoiceSet

// InputChoiceSet - https://adaptivecards.io/explorer/Input.ChoiceSet.html
type InputChoiceSet struct {
	*Common
	Choices            []InputChoice     `json:"choices,omitempty"`
	ChoicesData        *DataQuery        `json:"choices.data,omitempty" version:"1.6"`
	ErrorMessage       string            `json:"errorMessage,omitempty" version:"1.3"`
	IsMultiSelect      bool              `json:"isMultiSelect,omitempty"`
	IsRequired         bool              `json:"isRequired,omitempty" version:"1.3"`
	Label              string            `json:"label,omitempty" version:"1.3"`
	LabelPosition      *LabelPosition    `json:"labelPosition,omitempty" version:"1.6"`
	LabelWidth         any               `json:"labelWidth,omitempty" version:"1.6"`
	MinColumnWidth     string            `json:"minColumnWidth,omitempty" version:"1.5"`
	Placeholder        string            `json:"placeholder,omitempty"`
	Style              *ChoiceInputStyle `json:"style,omitempty"`
	UseMultipleColumns any               `json:"useMultipleColumns"`
	Value              string            `json:"value,omitempty"`
	ValueChangedAction Action            `json:"valueChangedAction"`
	Wrap               bool              `json:"wrap,omitempty" version:"1.2"`
}

func (i InputChoiceSet) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// InputChoiceSet to JSON.
func (i InputChoiceSet) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(i)
	// return marshalWithType(i, "Input.ChoiceSet")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an InputChoiceSet struct.
func (i *InputChoiceSet) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, i)
}

// region InputChoice

// InputChoice describes a choice for use in a ChoiceSet
type InputChoice struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

// MarshalJSON ensures that the "type" field is included when marshaling an
// InputChoice to JSON.
func (i InputChoice) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(i)
	// return marshalWithType(i, "Input.Choice")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an InputChoice struct.
func (i *InputChoice) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, i)
}

// endregion InputChoice

// region DataQuery

// DataQuery enables dynamic choice fetching from a bot (version 1.6)
type DataQuery struct {
	Dataset string `json:"dataset"`
	Count   int    `json:"count,omitempty"`
	Skip    int    `json:"skip,omitempty"`
}

// MarshalJSON ensures that the "type" field is included when marshaling a
// DataQuery to JSON.
func (d DataQuery) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(d)
	// return marshalWithType(d, "Data.Query")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a DataQuery struct.
func (d *DataQuery) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, d)
}

// endregion DataQuery

// endregion InputChoiceSet

// region InputDate

// InputDate - https://adaptivecards.io/explorer/Input.Date.html
type InputDate struct {
	*Common
	Max                string         `json:"max,omitempty"`
	Min                string         `json:"min,omitempty"`
	ErrorMessage       string         `json:"errorMessage,omitempty" version:"1.3"`
	IsRequired         bool           `json:"isRequired,omitempty" version:"1.3"`
	Label              string         `json:"label,omitempty" version:"1.3"`
	LabelPosition      *LabelPosition `json:"labelPosition,omitempty" version:"1.6"`
	LabelWidth         any            `json:"labelWidth,omitempty" version:"1.6"`
	Placeholder        string         `json:"placeholder,omitempty"`
	Value              string         `json:"value,omitempty"`
	ValueChangedAction Action         `json:"valueChangedAction"`
}

func (i InputDate) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// InputDate to JSON.
func (i InputDate) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(i)
	// return marshalWithType(i, "Input.Date")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an InputDate struct.
func (i *InputDate) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, i)
}

// endregion InputDate

// region InputNumber

// InputNumber - https://adaptivecards.io/explorer/Input.Number.html
type InputNumber struct {
	*Common
	Max                float64        `json:"max,omitempty"`
	Min                float64        `json:"min,omitempty"`
	ErrorMessage       string         `json:"errorMessage,omitempty" version:"1.3"`
	IsRequired         bool           `json:"isRequired,omitempty" version:"1.3"`
	Label              string         `json:"label,omitempty" version:"1.3"`
	LabelPosition      *LabelPosition `json:"labelPosition,omitempty" version:"1.6"`
	LabelWidth         any            `json:"labelWidth,omitempty" version:"1.6"`
	Placeholder        string         `json:"placeholder,omitempty"`
	Value              float64        `json:"value,omitempty"`
	ValueChangedAction Action         `json:"valueChangedAction"`
}

func (i InputNumber) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// InputNumber to JSON.
func (i InputNumber) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(i)
	// return marshalWithType(i, "Input.Number")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an InputNumber struct.
func (i *InputNumber) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, i)
}

// endregion InputNumber

// region InputRating

// InputRating - https://adaptivecards.io/explorer/Input.Rating.html
type InputRating struct {
	*Common
	AllowHalfSteps     bool           `json:"allowHalfSteps,omitempty"`
	Color              *RatingColor   `json:"color,omitempty"`
	Max                int            `json:"max,omitempty"`
	Size               *RatingSize    `json:"size,omitempty"`
	ErrorMessage       string         `json:"errorMessage,omitempty" version:"1.3"`
	IsRequired         bool           `json:"isRequired,omitempty" version:"1.3"`
	Label              string         `json:"label,omitempty" version:"1.3"`
	LabelPosition      *LabelPosition `json:"labelPosition,omitempty" version:"1.6"`
	LabelWidth         any            `json:"labelWidth,omitempty" version:"1.6"`
	Placeholder        string         `json:"placeholder,omitempty"`
	Value              float64        `json:"value,omitempty"`
	ValueChangedAction Action         `json:"valueChangedAction"`
}

// MarshalJSON ensures that the "type" field is included when marshaling an
// InputRating to JSON.
func (i InputRating) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(i)
	// return marshalWithType(i, "Input.Rating")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an InputRating struct.
func (i *InputRating) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, i)
}

// endregion InputRating

// region InputText

// InputText - https://adaptivecards.io/explorer/Input.Text.html
type InputText struct {
	*Common
	InlineAction       Action          `json:"inlineAction,omitempty" version:"1.2"`
	Style              *TextInputStyle `json:"style,omitempty"`
	IsMultiline        bool            `json:"isMultiline,omitempty"`
	MaxLength          int             `json:"maxLength,omitempty"`
	Regex              string          `json:"regex,omitempty" version:"1.3"`
	ErrorMessage       string          `json:"errorMessage,omitempty" version:"1.3"`
	IsRequired         bool            `json:"isRequired,omitempty" version:"1.3"`
	Label              string          `json:"label,omitempty" version:"1.3"`
	LabelPosition      *LabelPosition  `json:"labelPosition,omitempty" version:"1.6"`
	LabelWidth         any             `json:"labelWidth,omitempty" version:"1.6"`
	Placeholder        string          `json:"placeholder,omitempty"`
	Value              string          `json:"value,omitempty"`
	ValueChangedAction Action          `json:"valueChangedAction"`
}

func (i InputText) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// InputText to JSON.
func (i InputText) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(i)
	// return marshalWithType(i, "Input.Text")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an InputText struct.
func (i *InputText) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, i)
}

// endregion InputText

// region InputTime

// InputTime - https://adaptivecards.io/explorer/Input.Time.html
type InputTime struct {
	*Common
	Max                string         `json:"max,omitempty"` // In HH:MM format
	Min                string         `json:"min,omitempty"` // In HH:MM format
	ErrorMessage       string         `json:"errorMessage,omitempty" version:"1.3"`
	IsRequired         bool           `json:"isRequired,omitempty" version:"1.3"`
	Label              string         `json:"label,omitempty" version:"1.3"`
	LabelPosition      *LabelPosition `json:"labelPosition,omitempty" version:"1.6"`
	LabelWidth         any            `json:"labelWidth,omitempty" version:"1.6"`
	Placeholder        string         `json:"placeholder,omitempty"`
	Value              string         `json:"value,omitempty"`
	ValueChangedAction Action         `json:"valueChangedAction"`
}

func (i InputTime) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// InputTime to JSON.
func (i InputTime) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(i)
	// return marshalWithType(i, "Input.Time")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an InputTime struct.
func (i *InputTime) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, i)
}

// endregion InputTime

// region InputToggle

// InputToggle - https://adaptivecards.io/explorer/Input.Toggle.html
type InputToggle struct {
	*Common
	Title              string         `json:"title"`
	ValueOff           string         `json:"valueOff,omitempty"`
	ValueOn            string         `json:"valueOn,omitempty"`
	Wrap               bool           `json:"wrap,omitempty" version:"1.2"`
	ShowTitle          bool           `json:"showTitle,omitempty" version:"1.3"`
	ErrorMessage       string         `json:"errorMessage,omitempty" version:"1.3"`
	IsRequired         bool           `json:"isRequired,omitempty" version:"1.3"`
	Label              string         `json:"label,omitempty" version:"1.3"`
	LabelPosition      *LabelPosition `json:"labelPosition,omitempty" version:"1.6"`
	LabelWidth         any            `json:"labelWidth,omitempty" version:"1.6"`
	Placeholder        string         `json:"placeholder,omitempty"`
	Value              string         `json:"value,omitempty"`
	ValueChangedAction Action         `json:"valueChangedAction"`
}

func (i InputToggle) isElement() {}

// MarshalJSON ensures that the "type" field is included when marshaling an
// InputToggle to JSON.
func (i InputToggle) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(i)
	// return marshalWithType(i, "Input.Toggle")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an InputToggle struct.
func (i *InputToggle) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, i)
}

// endregion InputToggle

// endregion Inputs

// region Types

// region BackgroundImage

// BackgroundImage specifies a background image
type BackgroundImage struct {
	URL                 string               `json:"url"`
	FillMode            *ImageFillMode       `json:"fillMode,omitempty"`
	HorizontalAlignment *HorizontalAlignment `json:"horizontalAlignment,omitempty"`
	VerticalAlignment   *VerticalAlignment   `json:"verticalAlignment,omitempty"`
}

// endregion BackgroundImage

// region RefreshDefinition

// Refresh defines how a card can be refreshed by making a request to the target Bot
type Refresh struct {
	Action  *ActionExecute `json:"action,omitempty"`
	Expires string         `json:"expires,omitempty" version:"1.6"`
	UserIDs []string       `json:"userIds,omitempty"`
}

// MarshalJSON ensures that the "type" field is included when marshaling a
// Refresh to JSON.
func (r Refresh) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(r)
	// return marshalWithType(r, "Refresh")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a Refresh struct.
func (r *Refresh) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, r)
}

// endregion Refresh

// region AuthCardButton

// AuthCardButton defines a button as displayed when prompting a user to authenticate
type AuthCardButton struct {
	Type  string `json:"type"`
	Title string `json:"title,omitempty"`
	Image string `json:"image,omitempty"`
	Value string `json:"value"`
}

// endregion AuthCardButton

// region TokenExchangeResource

// TokenExchangeResource defines information required to enable on-behalf-of
// single sign-on
type TokenExchangeResource struct {
	ID         string `json:"id"`
	URI        string `json:"uri"`
	ProviderID string `json:"providerId"`
}

// MarshalJSON ensures that the "type" field is included when marshaling a
// TokenExchangeResource to JSON.
func (t TokenExchangeResource) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(t)
	// return marshalWithType(t, "TokenExchangeResource")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a TokenExchangeResource struct.
func (t *TokenExchangeResource) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, t)
}

// endregion TokenExchangeResource

// region Authentication

// Authentication defines authentication information associated with a card
type Authentication struct {
	Text                  string                 `json:"text,omitempty"`
	ConnectionName        string                 `json:"connectionName,omitempty"`
	TokenExchangeResource *TokenExchangeResource `json:"tokenExchangeResource,omitempty"`
	Buttons               []AuthCardButton       `json:"buttons,omitempty"`
}

// MarshalJSON ensures that the "type" field is included when marshaling an
// Authentication to JSON.
func (a Authentication) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(a)
	// return marshalWithType(a, "Authentication")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into an Authentication struct.
func (a *Authentication) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, a)
}

// endregion Authentication

// region Metadata

// Metadata defines various metadata properties (version 1.6)
type Metadata struct {
	WebURL string `json:"webUrl,omitempty"`
}

// MarshalJSON ensures that the "type" field is included when marshaling
// Metadata to JSON.
func (m Metadata) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(m)
	// return marshalWithType(m, "Metadata")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a Metadata struct.
func (m *Metadata) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, m)
}

// endregion Metadata

// region StringResource

// StringResource defines a string resource with a default value and optional
// localized values. LocalizedValues is a map where the key is the locale
// (e.g., "en", "fr-FR", "es-ES") and the value is the localized string.
type StringResource struct {
	DefaultValue    string            `json:"defaultValue,omitempty" version:"1.5"`
	LocalizedValues map[string]string `json:"localizedValues,omitempty" version:"1.5"`
	Key             string            `json:"key,omitempty"`
}

// endregion StringResource

// region Resources

// Resources defines a collection of resources, such as localized strings
type Resources struct {
	StringResources []StringResource `json:"stringResources,omitempty"`
	Key             string           `json:"key,omitempty"`
}

// endregion Resources

// region MentionedEntity

// MentionedEntity defines an entity that is mentioned in a message,
// such as a user or a tag
type MentionedEntity struct {
	ID          string      `json:"id,omitempty"` // The Id of a person (typically a Microsoft Entra user Id) or tag.
	MentionType MentionType `json:"mentionType,omitempty"`
	Name        string      `json:"name,omitempty"`
	Key         string      `json:"key,omitempty"`
}

// endregion MentionedEntity

// region Mention

// Mention defines a mention of one or more entities in a message,
// along with the text of the mention
type Mention struct {
	Mentioned []MentionedEntity `json:"mentioned,omitempty"`
	Text      string            `json:"text,omitempty"`
	Key       string            `json:"key,omitempty"`
}

// MarshalJSON ensures that the "type" field is included when marshaling a
// Mention to JSON.
func (c Mention) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "mention")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a Mention struct.
func (c *Mention) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// endregion Mention

// region TeamsCardProperties

// TeamsCardProperties defines properties for a Teams card, including entities and width.
type TeamsCardProperties struct {
	Entities []Mention       `json:"entities,omitempty" version:"1.0"`
	Width    *TeamsCardWidth `json:"width,omitempty" version:"1.0"`
	Key      string          `json:"key,omitempty"`
}

// endregion TeamsCardProperties

// region TeamsSubmitActionProperties

// TeamsSubmitActionProperties defines properties for a Teams card, including entities and width.
type TeamsSubmitActionProperties struct {
	Feedback []TeamsSubmitActionFeedback `json:"feedback,omitempty" version:"1.0"`
	Key      string                      `json:"key,omitempty"`
}

// endregion TeamsSubmitActionProperties

// TeamsSubmitActionFeedback defines feedback options for a Teams submit action,
// such as hiding the original card or providing a message to display after
// submission.
type TeamsSubmitActionFeedback struct {
	Hide bool   `json:"hide,omitempty" version:"1.0"`
	Key  string `json:"key,omitempty"`
}

// endregion Types

// region Layout

// Layout represents a layout for arranging elements in a card.
// It can be one of several types, such as LayoutStack, LayoutFlow,
// or LayoutAreaGrid.
type Layout interface {
	isLayout()
}

// region Layout.Stack

// LayoutStack represents a layout that arranges elements in a vertical stack.
type LayoutStack struct {
	Key         string       `json:"key,omitempty" version:"1.0"`
	TargetWidth *TargetWidth `json:"targetWidth,omitempty" version:"1.5"`
}

func (c LayoutStack) isLayout() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// LayoutStack to JSON.
func (c LayoutStack) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "LayoutStack")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a LayoutStack struct.
func (c *LayoutStack) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// endregion Layout.Stack

// region Layout.Flow

// LayoutFlow represents a layout that arranges elements in a horizontal flow.
type LayoutFlow struct {
	ColumnSpacing            *Spacing             `json:"columnSpacing,omitempty" version:"1.5"`
	HorizontalItemsAlignment *HorizontalAlignment `json:"horizontalItemsAlignment,omitempty" version:"1.0"`
	ItemFit                  *ItemFit             `json:"itemFit,omitempty" version:"1.5"`
	ItemWidth                string               `json:"itemWidth,omitempty" version:"1.5"`    // Should not be used if MaxItemWidth or MinItemWidth are specified
	MaxItemWidth             string               `json:"maxItemWidth,omitempty" version:"1.5"` // Should not be used if ItemWidth is specified
	MinItemWidth             string               `json:"minItemWidth,omitempty" version:"1.5"` // Should not be used if ItemWidth is specified
	RowSpacing               *Spacing             `json:"rowSpacing,omitempty" version:"1.5"`
	VerticalItemsAlignment   *VerticalAlignment   `json:"verticalItemsAlignment,omitempty" version:"1.5"`
	Key                      string               `json:"key,omitempty" version:"1.0"`
	TargetWidth              *TargetWidth         `json:"targetWidth,omitempty" version:"1.0"`
}

func (c LayoutFlow) isLayout() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// LayoutFlow to JSON.
func (c LayoutFlow) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "LayoutFlow")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a LayoutFlow struct.
func (c *LayoutFlow) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// endregion Layout.Flow

// region GridArea

// GridArea represents an area within a grid layout, defined by its
// starting column and row, as well as how many columns and rows it spans.
// It can also have an optional name for reference.
type GridArea struct {
	Column     int    `json:"column,omitempty" version:"1.5"`
	ColumnSpan int    `json:"columnSpan,omitempty" version:"1.5"`
	Name       string `json:"name,omitempty" version:"1.5"`
	Row        int    `json:"row,omitempty" version:"1.5"`
	RowSpan    int    `json:"rowSpan,omitempty" version:"1.5"`
	Key        string `json:"key,omitempty" version:"1.0"`
}

// endregion GridArea

// region GridColumnWidth

// GridColumnWidth represents columns in the grid layout, defined as a
// percentage of the available width or in pixels using the `<number>px` format.
// Each column is specified as either a string ("50px") or as a number (1).
type GridColumnWidth struct {
	Value any `json:"-"`
}

// MarshalJSON allows GridColumnWidth to be marshaled as either a string
// (e.g., "50px") or a number (e.g., 1).
func (w GridColumnWidth) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.Value)
}

// UnmarshalJSON allows GridColumnWidth to be unmarshaled from either a string
// (e.g., "50px") or a number (e.g., 1). It validates the format and ensures
// that numbers are whole integers.
func (w *GridColumnWidth) UnmarshalJSON(data []byte) error {
	var i any
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}

	switch v := i.(type) {
	case float64:
		// JSON unmarshals all numbers as float64
		// Only accept whole numbers (integers)
		if v != float64(int(v)) {
			return fmt.Errorf("invalid value: GridColumnWidth must be a whole number, got %v", v)
		}
		w.Value = int(v)
		return nil
	case string:
		if !strings.HasSuffix(v, "px") {
			return fmt.Errorf("invalid format: string values must end in 'px'")
		}
		w.Value = v
	default:
		return fmt.Errorf("invalid type for GridColumnWidth: %T (%v)", v, v)
	}
	return nil
}

// endregion GridColumnWidth

// region Layout.AreaGrid

// LayoutAreaGrid represents a layout that arranges elements in a grid
// defined by areas and columns.
type LayoutAreaGrid struct {
	Areas         []GridArea        `json:"areas,omitempty" version:"1.5"`
	Columns       []GridColumnWidth `json:"columns,omitempty" version:"1.5"`
	ColumnSpacing *Spacing          `json:"columnSpacing,omitempty" version:"1.5"`
	RowSpacing    *Spacing          `json:"rowSpacing,omitempty" version:"1.5"`
	Key           string            `json:"key,omitempty" version:"1.0"`
	TargetWidth   *TargetWidth      `json:"targetWidth,omitempty" version:"1.0"`
}

func (c LayoutAreaGrid) isLayout() {}

// MarshalJSON ensures that the "type" field is included when marshaling a
// LayoutAreaGrid to JSON.
func (c LayoutAreaGrid) MarshalJSON() ([]byte, error) {
	return SmartMarshalFromJSON(c)
	// return marshalWithType(c, "LayoutAreaGrid")
}

// UnmarshalJSON ensures that the "type" field is validated when unmarshaling
// JSON into a LayoutAreaGrid struct.
func (c *LayoutAreaGrid) UnmarshalJSON(data []byte) error {
	return SmartUnmarshalJSON(data, c)
}

// endregion Layout.AreaGrid

// endregion Layout

// region Other

// ThemedURL represents a URL that can vary based on the current theme
// (light, dark, high contrast).
type ThemedURL struct {
	Theme Theme  `json:"theme,omitempty" version:"1.5"`
	URL   string `json:"url,omitempty" version:"1.5"`
	Key   string `json:"key,omitempty" version:"1.0"`
}

// endregion Other
