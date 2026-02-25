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

// ActionMode determines whether an action is displayed with a button or in the overflow menu
type ActionMode string

// ActionMode controls how an Action is displayed.
// "primary" actions are displayed as buttons, while
// "secondary" actions are displayed in the overflow menu.
const (
	ActionModePrimary   ActionMode = "primary"
	ActionModeSecondary ActionMode = "secondary"
)

// ActionStyle controls the style of an Action
type ActionStyle string

// ActionStyle controls the style of an Action.
// "default" is the default style,
// "positive" indicates a positive action, and
// "destructive" indicates a destructive action.
const (
	ActionStyleDefault     ActionStyle = "default"
	ActionStylePositive    ActionStyle = "positive"
	ActionStyleDestructive ActionStyle = "destructive"
)

// AssociatedInputs determines which inputs are associated with an action
type AssociatedInputs string

// AssociatedInputs controls which inputs are associated with an action.
// "auto" (the default) associates the action with all inputs in the same container.
// "none" indicates that the action isn't associated with any inputs.
const (
	AssociatedInputsAuto AssociatedInputs = "auto"
	AssociatedInputsNone AssociatedInputs = "none"
)

// BlockElementHeight controls the height of block elements
type BlockElementHeight string

// BlockElementHeight controls the height of block elements.
// "auto" (the default) allows the element to determine its own height based on its content,
//
//	while "stretch" forces the element to stretch to fill the available vertical space in its container.
const (
	BlockElementHeightAuto    BlockElementHeight = "auto"
	BlockElementHeightStretch BlockElementHeight = "stretch"
)

// ChoiceInputStyle is the style hint for Input.ChoiceSet
type ChoiceInputStyle string

// ChoiceInputStyle is the style hint for Input.ChoiceSet.
const (
	ChoiceInputStyleCompact  ChoiceInputStyle = "compact"
	ChoiceInputStyleExpanded ChoiceInputStyle = "expanded"
)

// Colors defines the color options for text and other elements
type Colors string

// Colors defines the color options for text and other elements.
const (
	ColorNone        Colors = ""
	ColorDefault     Colors = "Default"
	ColorDark        Colors = "Dark"
	ColorLight       Colors = "Light"
	ColorAccent      Colors = "Accent"
	ColorGood        Colors = "Good"
	ColorWarning     Colors = "Warning"
	ColorAttention   Colors = "Attention"
	ColorInformative Colors = "Informative"
	ColorSubtle      Colors = "Subtle"
)

// ContainerStyle defines style hints for containers
type ContainerStyle string

// ContainerStyle defines style hints for containers.
const (
	ContainerStyleDefault   ContainerStyle = "default"
	ContainerStyleEmphasis  ContainerStyle = "emphasis"
	ContainerStyleGood      ContainerStyle = "good"
	ContainerStyleAttention ContainerStyle = "attention"
	ContainerStyleWarning   ContainerStyle = "warning"
	ContainerStyleAccent    ContainerStyle = "accent"
)

// FallbackOption defines fallback behavior
type FallbackOption string

// FallbackOption controls fallback behavior when an element can't be rendered.
const (
	FallbackOptionDrop FallbackOption = "drop"
)

// FontSize controls the size of text
type FontSize string

// FontSize controls the size of text.
const (
	FontSizeDefault    FontSize = "Default"
	FontSizeSmall      FontSize = "Small"
	FontSizeMedium     FontSize = "Medium"
	FontSizeLarge      FontSize = "Large"
	FontSizeExtraLarge FontSize = "ExtraLarge"
)

// FontType defines the type of font to use
type FontType string

// FontType defines the type of font to use.
const (
	FontTypeDefault   FontType = "default"
	FontTypeMonospace FontType = "monospace"
)

// FontWeight controls the weight of text
type FontWeight string

// FontWeight controls the weight of text.
const (
	FontWeightDefault FontWeight = "Default"
	FontWeightLighter FontWeight = "Lighter"
	FontWeightBolder  FontWeight = "Bolder"
)

// HorizontalAlignment controls horizontal positioning within containers
type HorizontalAlignment string

// HorizontalAlignment controls horizontal positioning within containers.
const (
	HorizontalAlignmentLeft   HorizontalAlignment = "Left"
	HorizontalAlignmentCenter HorizontalAlignment = "Center"
	HorizontalAlignmentRight  HorizontalAlignment = "Right"
)

// IconPosition controls the position of the icon.
type IconPosition string

// IconPosition controls the position of the icon.
const (
	IconPositionBefore IconPosition = "Before"
	IconPositionAfter  IconPosition = "After"
)

// IconSize controls the approximate size of images
type IconSize string

// IconSize controls the approximate size of images.
const (
	IconSizeXXSmall  IconSize = "XXSmall"
	IconSizeXSmall   IconSize = "XSmall"
	IconSizeSmall    IconSize = "Small"
	IconSizeStandard IconSize = "Standard"
	IconSizeMedium   IconSize = "Medium"
	IconSizeLarge    IconSize = "Large"
	IconSizeXLarge   IconSize = "XLarge"
	IconSizeXXLarge  IconSize = "XXLarge"
)

// IconStyle controls the style of the icon.
type IconStyle string

// IconStyle controls the style of the icon.
const (
	IconStyleRegular IconStyle = "Regular"
	IconStyleFilled  IconStyle = "Filled"
)

// ImageFit controls how item should fit inside the container.
type ImageFit string

// ImageFit controls how item should fit inside the container.
const (
	ImageFitCover   ImageFit = "Cover"
	ImageFitContain ImageFit = "Contain"
	ImageFitFill    ImageFit = "Fill"
)

// ImageFillMode controls how background images are displayed
type ImageFillMode string

// ImageFillMode controls how background images are displayed.
const (
	ImageFillModeCover              ImageFillMode = "Cover"
	ImageFillModeRepeatHorizontally ImageFillMode = "RepeatHorizontally"
	ImageFillModeRepeatVertically   ImageFillMode = "RepeatVertically"
	ImageFillModeRepeat             ImageFillMode = "Repeat"
)

// ImageSize controls the approximate size of images
type ImageSize string

// ImageSize controls the approximate size of images.
const (
	ImageSizeAuto    ImageSize = "auto"
	ImageSizeStretch ImageSize = "stretch"
	ImageSizeSmall   ImageSize = "small"
	ImageSizeMedium  ImageSize = "medium"
	ImageSizeLarge   ImageSize = "large"
)

// ImageStyle controls how an Image is displayed
type ImageStyle string

// ImageStyle controls how an Image is displayed.
const (
	ImageStyleDefault        ImageStyle = "Default"
	ImageStylePerson         ImageStyle = "Person"
	ImageStyleRoundedCorners ImageStyle = "RoundedCorners" // MS Teams only
)

// ItemFit controls how item should fit inside the container.
type ItemFit string

// ItemFit controls how item should fit inside the container.
const (
	ItemFitFit  ItemFit = "Fit"
	ItemFitFill ItemFit = "Fill"
)

// InputValidationStyle controls the validation style for inputs (version 1.6)
type InputValidationStyle string

// InputValidationStyle controls the validation style for inputs (version 1.6).
const (
	InputValidationStyleDefault       InputValidationStyle = "default"
	InputValidationStyleRevealOnHover InputValidationStyle = "revealOnHover"
)

// LabelPosition determines where input labels are positioned (version 1.6)
type LabelPosition string

// LabelPosition controls where input labels are positioned (version 1.6).
const (
	LabelPositionInline LabelPosition = "inline"
	LabelPositionAbove  LabelPosition = "above"
)

// Spacing specifies how much spacing should be used
type Spacing string

// Spacing specifies how much spacing should be used.
const (
	SpacingDefault    Spacing = "Default"
	SpacingNone       Spacing = "None"
	SpacingSmall      Spacing = "Small"
	SpacingMedium     Spacing = "Medium"
	SpacingLarge      Spacing = "Large"
	SpacingExtraLarge Spacing = "ExtraLarge"
	SpacingPadding    Spacing = "Padding"
)

// TextBlockStyle controls how a TextBlock behaves
type TextBlockStyle string

// TextBlockStyle controls how a TextBlock behaves.
const (
	TextBlockStyleDefault TextBlockStyle = "default"
	TextBlockStyleHeading TextBlockStyle = "heading"
)

// TextInputStyle is the style hint for text input
type TextInputStyle string

// TextInputStyle is the style hint for text input.
const (
	TextInputStyleText     TextInputStyle = "Text"
	TextInputStyleTel      TextInputStyle = "Tel"
	TextInputStyleURL      TextInputStyle = "Url"
	TextInputStyleEmail    TextInputStyle = "Email"
	TextInputStylePassword TextInputStyle = "Password"
)

// Theme controls the overall theme of the card.
type Theme string

// Theme controls the overall theme of the card.
const (
	ThemeLight Theme = "Light"
	ThemeDark  Theme = "Dark"
)

// VerticalAlignment controls vertical positioning
type VerticalAlignment string

// VerticalAlignment controls vertical positioning.
const (
	VerticalAlignmentTop    VerticalAlignment = "Top"
	VerticalAlignmentCenter VerticalAlignment = "Center"
	VerticalAlignmentBottom VerticalAlignment = "Bottom"
)

// InsertPosition controls where the image should be inserted when using ActionInsertImage.
type InsertPosition string

// InsertPosition controls where the image should be inserted when using ActionInsertImage.
const (
	InsertPositionSelection InsertPosition = "Selection"
	InsertPositionTop       InsertPosition = "Top"
	InsertPositionBottom    InsertPosition = "Bottom"
)

// TargetWidth controls for which card width the element should be displayed.
// If targetWidth isn't specified, the element is rendered at all card widths.
// Using targetWidth makes it possible to author responsive cards that adapt
// their layout to the available horizontal space.
type TargetWidth string

// TargetWidth controls for which card width the element should be displayed.
const (
	TargetWidthVeryNarrow        TargetWidth = "VeryNarrow"
	TargetWidthNarrow            TargetWidth = "Narrow"
	TargetWidthStandard          TargetWidth = "Standard"
	TargetWidthWide              TargetWidth = "Wide"
	TargetWidthAtLeastVeryNarrow TargetWidth = "AtLeast:VeryNarrow"
	TargetWidthAtMostVeryNarrow  TargetWidth = "AtMost:VeryNarrow"
	TargetWidthAtLeastNarrow     TargetWidth = "AtLeast:Narrow"
	TargetWidthAtMostNarrow      TargetWidth = "AtMost:Narrow"
	TargetWidthAtLeastStandard   TargetWidth = "AtLeast:Standard"
	TargetWidthAtMostStandard    TargetWidth = "AtMost:Standard"
	TargetWidthAtLeastWide       TargetWidth = "AtLeast:Wide"
	TargetWidthAtMostWide        TargetWidth = "AtMost:Wide"
)

// Requires is a list of capabilities the element requires the host application
// to support. If the host application doesn't support at least one of the
// listed capabilities, the element is not rendered (or its fallback is rendered
// if provided).
type Requires map[string]string

// CodeLanguage controls the language the code snippet is expressed in.
type CodeLanguage string

// CodeLanguage controls the language the code snippet is expressed in.
const (
	CodeLanguageBash       CodeLanguage = "Bash"
	CodeLanguageC          CodeLanguage = "C"
	CodeLanguageCpp        CodeLanguage = "Cpp"
	CodeLanguageCSharp     CodeLanguage = "CSharp"
	CodeLanguageCSS        CodeLanguage = "Css"
	CodeLanguageDos        CodeLanguage = "Dos"
	CodeLanguageGo         CodeLanguage = "Go"
	CodeLanguageGraphql    CodeLanguage = "Graphql"
	CodeLanguageHTML       CodeLanguage = "Html"
	CodeLanguageJava       CodeLanguage = "Java"
	CodeLanguageJavaScript CodeLanguage = "JavaScript"
	CodeLanguageJSON       CodeLanguage = "Json"
	CodeLanguageObjectiveC CodeLanguage = "ObjectiveC"
	CodeLanguagePerl       CodeLanguage = "Perl"
	CodeLanguagePHP        CodeLanguage = "Php"
	CodeLanguagePlainText  CodeLanguage = "PlainText"
	CodeLanguagePowerShell CodeLanguage = "PowerShell"
	CodeLanguagePython     CodeLanguage = "Python"
	CodeLanguageSQL        CodeLanguage = "Sql"
	CodeLanguageTypeScript CodeLanguage = "TypeScript"
	CodeLanguageVbNet      CodeLanguage = "VbNet"
	CodeLanguageVerilog    CodeLanguage = "Verilog"
	CodeLanguageVHDL       CodeLanguage = "Vhdl"
	CodeLanguageXML        CodeLanguage = "Xml"
)

// Appearance controls the strength of the background color.
type Appearance string

// Appearance controls the strength of the background color.
const (
	AppearanceFilled Appearance = "Filled"
	AppearanceTint   Appearance = "Tint"
)

// BadgeShape controls the shape of the badge.
type BadgeShape string

// BadgeShape controls the shape of the badge.
const (
	BadgeShapeSquare   BadgeShape = "Square"
	BadgeShapeRounded  BadgeShape = "Rounded"
	BadgeShapeCircular BadgeShape = "Circular"
)

// BadgeSize controls the size of the badge.
type BadgeSize string

// BadgeSize controls the size of the badge.
const (
	BadgeSizeMedium     BadgeSize = "Medium"
	BadgeSizeLarge      BadgeSize = "Large"
	BadgeSizeExtraLarge BadgeSize = "ExtraLarge"
)

// RatingColor controls the color of the rating element.
type RatingColor string

// RatingColor controls the color of the rating element.
const (
	RatingColorNeutral  RatingColor = "Neutral"
	RatingColorMarigold RatingColor = "Marigold"
)

// RatingSize controls the size of the rating element.
type RatingSize string

// RatingSize controls the size of the rating element.
const (
	RatingSizeMedium RatingSize = "Medium"
	RatingSizeLarge  RatingSize = "Large"
)

// RatingStyle controls the style of the rating element.
type RatingStyle string

// RatingStyle controls the style of the rating element.
const (
	RatingStyleDefault RatingStyle = "Default"
	RatingStyleCompact RatingStyle = "Compact"
)

// TeamsCardWidth controls the width of a card in Microsoft Teams.
type TeamsCardWidth string

// TeamsCardWidth controls the width of a card in Microsoft Teams.
const (
	TeamsCardWidthFull TeamsCardWidth = "Full"
)

// MentionType controls the type of the mentioned entity.
type MentionType string

// MentionType controls the type of the mentioned entity in a mention element.
const (
	MentionTypePerson MentionType = "Person"
	MentionTypeTag    MentionType = "Tag"
)

// ProgressRingLabelPosition controls the position of the label in a progress ring.
type ProgressRingLabelPosition string

// ProgressRingLabelPosition controls the position of the label in a progress ring.
const (
	ProgressRingLabelPositionBefore ProgressRingLabelPosition = "Before"
	ProgressRingLabelPositionAfter  ProgressRingLabelPosition = "After"
	ProgressRingLabelPositionAbove  ProgressRingLabelPosition = "Above"
	ProgressRingLabelPositionBelow  ProgressRingLabelPosition = "Below"
)

// ProgressRingSize controls the size of the progress ring.
type ProgressRingSize string

// ProgressRingSize controls the size of the progress ring.
const (
	ProgressRingSizeTiny   ProgressRingSize = "Tiny"
	ProgressRingSizeSmall  ProgressRingSize = "Small"
	ProgressRingSizeMedium ProgressRingSize = "Medium"
	ProgressRingSizeLarge  ProgressRingSize = "Large"
)

// PopoverPosition controls the position of the label in a progress ring.
type PopoverPosition string

// PopoverPosition controls the position of the label in a progress ring.
const (
	PopoverPositionBefore PopoverPosition = "Before"
	PopoverPositionAfter  PopoverPosition = "After"
	PopoverPositionAbove  PopoverPosition = "Above"
	PopoverPositionBelow  PopoverPosition = "Below"
)
