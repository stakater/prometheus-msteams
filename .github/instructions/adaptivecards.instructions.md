---
applyTo: "pkg/adaptivecards/**/*.go"
---
# Copilot Instructions for Adaptive Cards Package

## Project Overview

This project implements a complete Go package for Microsoft Adaptive Cards with full schema support for versions 1.0 through 1.6. The package is located in `pkg/adaptivecards/` and provides type-safe Go structs with JSON marshaling/unmarshaling and version validation capabilities.

## Schema Locations

- **`v1.6.0`** - https://adaptivecards.io/schemas/1.6.0/adaptive-card.json
- **`v1.5.0`** - https://adaptivecards.io/schemas/1.5.0/adaptive-card.json
- **`v1.4.0`** - https://adaptivecards.io/schemas/1.4.0/adaptive-card.json
- **`v1.3.0`** - https://adaptivecards.io/schemas/1.3.0/adaptive-card.json
- **`v1.2.1`** - https://adaptivecards.io/schemas/1.2.1/adaptive-card.json
- **`v1.2.0`** - https://adaptivecards.io/schemas/1.2.0/adaptive-card.json
- **`v1.1.0`** - https://adaptivecards.io/schemas/1.1.0/adaptive-card.json


## Package Structure

### Core Files

- **`adaptivecards.go`** - Main type definitions for all Adaptive Cards elements, actions, and containers
- **`constants.go`** - Enum types for styling, alignment, colors, fonts, spacing, etc.
- **`version.go`** - Version validation system using reflection and struct tags
- **`version_test.go`** - Comprehensive test suite for version validation
- **`textblock.go`** - (Legacy) Separate TextBlock implementation (now merged into adaptivecards.go)

### Key Concepts

1. **Element Interface**: All card body elements implement `isElement()` marker method
2. **Action Interface**: All actions implement `isAction()` marker method
3. **Version Tags**: Struct fields use `version:"x.y"` tags to indicate minimum required version
4. **Type Injection**: All types automatically inject a `"type"` field during JSON marshaling

## Type Organization

### Hierarchy

```
AdaptiveCard (root)
├── Body: []Element
│   ├── Containers: ActionSet, Container, ColumnSet, FactSet, ImageSet, Table
│   ├── Elements: TextBlock, Image, Media, RichTextBlock
│   └── Inputs: Input.ChoiceSet, Input.Date, Input.Number, Input.Text, Input.Time, Input.Toggle
├── Actions: []Action
│   ├── Action.Execute (1.4+)
│   ├── Action.OpenUrl
│   ├── Action.ShowCard
│   ├── Action.Submit
│   └── Action.ToggleVisibility (1.2+)
└── Supporting Types: BackgroundImage, Refresh, Authentication, Metadata
```

### Version 1.6 Types (Latest)

- **`Metadata`** - Card metadata with `webUrl` property
- **`CaptionSource`** - Media caption sources for accessibility (used in `Media.CaptionSources`)
- **`DataQuery`** - Dynamic choice filtering for Input.ChoiceSet (properties: `dataset`, `count`, `skip`)

## Version Validation System

### How It Works

The package uses Go reflection to inspect struct fields at runtime and validate them against the card's declared version.

**Struct Tag Format:**
```go
type ActionSubmit struct {
    Title     string `json:"title"`
    Tooltip   string `json:"tooltip,omitempty" version:"1.5"` // Requires 1.5+
    IsEnabled *bool  `json:"isEnabled,omitempty" version:"1.5"` // Requires 1.5+
}
```

**Validation Function:**
```go
errors := ValidateVersion(card, card.Version)
// Returns []ValidationError with detailed field-level errors
```

**Convenience Method:**
```go
card := AdaptiveCard{Version: "1.0", RTL: boolPtr(true)}
errors := card.Validate() // RTL requires 1.5, will return error
```

### Version Requirements by Feature

| Version | Key Features |
|---------|-------------|
| 1.0 | Base elements, actions, containers |
| 1.1 | `Height`, `SelectAction`, `VerticalContentAlignment` |
| 1.2 | `BackgroundImage`, `MinHeight`, `Bleed`, `FontType`, `Fallback`, `IsVisible`, `Requires`, `InlineAction` |
| 1.3 | Input validation: `ErrorMessage`, `IsRequired`, `Label`, `Regex`, `Underline` |
| 1.4 | `ActionExecute`, `Refresh`, `Authentication` |
| 1.5 | `RTL`, `Tooltip`, `IsEnabled`, `Mode`, `TextBlock.Style` (heading), `Table` element |
| 1.6 | `Metadata`, `CaptionSources`, `Data.Query`, `Refresh.Expires`, Input `LabelPosition`/`LabelWidth`/`InputStyle` |

## Coding Guidelines

### 1. Adding New Types

When adding a new Adaptive Cards type:

```go
// Step 1: Define the struct with proper tags
type NewElement struct {
    RequiredField string            `json:"requiredField"`
    OptionalField string            `json:"optionalField,omitempty" version:"1.x"`
    ID            string            `json:"id,omitempty"`
    IsVisible     *bool             `json:"isVisible,omitempty" version:"1.2"`
    Requires      map[string]string `json:"requires,omitempty" version:"1.2"`
}

// Step 2: Implement the appropriate interface
func (n NewElement) isElement() {} // or isAction()

// Step 3: Add MarshalJSON to inject type field
func (n NewElement) MarshalJSON() ([]byte, error) {
    type alias NewElement
    return json.Marshal(&struct {
        Type string `json:"type"`
        *alias
    }{
        Type:  "NewElement",
        alias: (*alias)(&n),
    })
}

// Step 4: Add UnmarshalJSON for type validation
func (n *NewElement) UnmarshalJSON(data []byte) error {
    type alias NewElement
    aux := &struct {
        Type string `json:"type"`
        *alias
    }{
        alias: (*alias)(n),
    }
    if err := json.Unmarshal(data, &aux); err != nil {
        return err
    }
    if aux.Type != "" && aux.Type != "NewElement" {
        return fmt.Errorf("invalid type %s for NewElement", aux.Type)
    }
    return nil
}
```

### 2. Version Tag Rules

- **Always add version tags** for properties introduced after 1.0
- Use the exact version from the schema (e.g., `version:"1.2"`)
- Properties without tags are assumed to be version 1.0
- Common versioned properties:
  - `version:"1.1"`: Height, SelectAction, VerticalContentAlignment
  - `version:"1.2"`: Fallback, IsVisible, Requires, MinHeight, Bleed, BackgroundImage
  - `version:"1.3"`: ErrorMessage, IsRequired, Label (inputs)
  - `version:"1.5"`: Tooltip, IsEnabled, Mode, RTL
  - `version:"1.6"`: LabelPosition, LabelWidth, InputStyle, Metadata

### 3. Handling Polymorphic Types

For fields that can contain multiple types (Element or Action):

```go
// In parent struct
Body []Element `json:"body,omitempty"` // Can be any type implementing Element interface

// Custom marshal/unmarshal
func (a *AdaptiveCard) UnmarshalJSON(data []byte) error {
    // Use json.RawMessage for polymorphic fields
    type Alias AdaptiveCard
    aux := &struct {
        Body []json.RawMessage `json:"Body,omitempty"`
        *Alias
    }{
        Alias: (*Alias)(a),
    }
    
    if err := json.Unmarshal(data, &aux); err != nil {
        return err
    }
    
    // Unmarshal each raw element
    for _, rawElem := range aux.Body {
        elem, err := unmarshalElement(rawElem)
        if err != nil {
            return err
        }
        a.Body = append(a.Body, elem)
    }
    return nil
}
```

### 4. Testing Requirements

When adding new features, always add tests:

```go
func TestNewFeatureValidation(t *testing.T) {
    tests := []struct {
        name        string
        version     string
        expectError bool
    }{
        {"valid version", "1.5", false},
        {"invalid version", "1.0", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            card := AdaptiveCard{
                Version: tt.version,
                // Use new feature
            }
            errors := card.Validate()
            
            if tt.expectError && len(errors) == 0 {
                t.Error("expected validation errors but got none")
            }
            if !tt.expectError && len(errors) > 0 {
                t.Errorf("expected no errors but got: %v", errors)
            }
        })
    }
}
```

### 5. Pointer vs Value Types

**Use pointers for:**
- Optional boolean fields (to distinguish between false and unset)
- Optional complex types (BackgroundImage, Refresh, Authentication, Metadata)

```go
RTL       *bool            `json:"rtl,omitempty" version:"1.5"`
IsVisible *bool            `json:"isVisible,omitempty" version:"1.2"`
Metadata  *Metadata        `json:"metadata,omitempty" version:"1.6"`
```

**Use values for:**
- Required fields
- String/int fields (empty string/0 is valid "unset" state)
- Slices (nil slice works as "unset")

## Common Patterns

### Creating Cards

```go
card := adaptivecards.AdaptiveCard{
    Version: "1.5",
    Body: []adaptivecards.Element{
        adaptivecards.TextBlock{
            Text:   "Hello World",
            Size:   stringPtr("large"),
            Weight: stringPtr("bolder"),
        },
    },
    Actions: []adaptivecards.Action{
        adaptivecards.ActionSubmit{
            Title:     "Submit",
            Tooltip:   "Submit the form",
            IsEnabled: boolPtr(true),
        },
    },
}

// Validate before sending
if errors := card.Validate(); len(errors) > 0 {
    // Handle errors
}
```

### Helper Functions

```go
func boolPtr(b bool) *bool { return &b }
func stringPtr(s string) *string { return &s }
func intPtr(i int) *int { return &i }
```

## Schema Compliance

The package follows the official Microsoft Adaptive Cards schema:
- **Schema URL**: https://github.com/microsoft/AdaptiveCards/raw/refs/heads/main/schemas/1.6.0/adaptive-card.json
- **Documentation**: https://adaptivecards.io/explorer/
- **Versioning Guide**: https://docs.microsoft.com/en-us/adaptive-cards/authoring-cards/versioning

### Verification Checklist

When updating schema compliance:
1. ✅ All types from schema definitions are implemented
2. ✅ All properties have correct JSON tags
3. ✅ Version tags match schema version annotations
4. ✅ MarshalJSON/UnmarshalJSON methods inject/validate type field
5. ✅ Polymorphic fields use interfaces (Element, Action)
6. ✅ New features have corresponding tests
7. ✅ Documentation is updated (VERSION_VALIDATION.md)

## File Organization

### Region Comments

Use region comments to organize code:

```go
// region ContainerName

// Type definitions and methods

// endregion ContainerName
```

Current regions in adaptivecards.go:
- AdaptiveCard
- Containers (ActionSet, Container, ColumnSet, Column, FactSet, ImageSet)
- Actions (Execute, OpenURL, ShowCard, Submit, ToggleVisibility)
- Elements (TextBlock, Image, Media, RichTextBlock, TextRun)
- Inputs (Choice, ChoiceSet, Date, Number, Text, Time, Toggle, DataQuery)
- Tables (Table, TableRow, TableCell, TableColumnDefinition)
- Supporting Types (BackgroundImage, Refresh, Authentication, Metadata, CaptionSource)

## Error Handling

### Validation Errors

```go
type ValidationError struct {
    FieldName       string  // e.g., "AdaptiveCard.RTL"
    RequiredVersion Version // e.g., Version{1, 5}
    CardVersion     Version // e.g., Version{1, 0}
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("field %s requires version %s but card version is %s",
        e.FieldName, e.RequiredVersion, e.CardVersion)
}
```

### Unmarshal Errors

Always validate type fields during unmarshaling:

```go
if aux.Type != "" && aux.Type != "ExpectedType" {
    return fmt.Errorf("invalid type %s for ExpectedType", aux.Type)
}
```

## Best Practices

1. **Always validate cards before sending** to catch version compatibility issues early
2. **Use the lowest compatible version** to maximize client support
3. **Add version tags to all new properties** based on official schema
4. **Test with multiple versions** to ensure backward compatibility
5. **Keep MarshalJSON/UnmarshalJSON methods consistent** across all types
6. **Use meaningful type names** matching the schema (e.g., "Action.Submit", not "ActionSubmit")
7. **Document version requirements** in comments for complex features
8. **Run tests after any changes**: `go test ./pkg/adaptivecards -v`

## Related Files

- **VERSION_VALIDATION.md** - User-facing documentation for version validation
- **examples/validation_example.go** - Working examples of version validation
- **version_test.go** - Comprehensive test coverage (aim for 90%+)

## Future Enhancements

When new Adaptive Cards schema versions are released:

1. Fetch the new schema from the official repository
2. Compare with current implementation to identify new types/properties
3. Add new types with proper version tags
4. Update version.go if new validation patterns are needed
5. Add tests for new features
6. Update VERSION_VALIDATION.md with new version features
7. Update this file with any new patterns or guidelines

## Contact & Resources

- **Official Schema**: https://github.com/microsoft/AdaptiveCards/tree/main/schemas
- **Explorer**: https://adaptivecards.io/explorer/
- **Designer**: https://adaptivecards.io/designer/
- **Documentation**: https://docs.microsoft.com/en-us/adaptive-cards/
