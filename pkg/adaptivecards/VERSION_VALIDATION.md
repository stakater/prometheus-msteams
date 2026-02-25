# Adaptive Cards Version Validation

This package provides comprehensive support for Adaptive Cards schema versions 1.0 through 1.6, with built-in validation to ensure cards don't use features incompatible with their declared version.

## Quick Start

```go
import "github.com/stakater/prometheus-msteams/pkg/adaptivecards"

// Create a card
card := adaptivecards.AdaptiveCard{
    Version: "1.0",
    Body: []adaptivecards.Element{
        adaptivecards.TextBlock{
            Text: "Hello World",
        },
    },
}

// Validate it
if errors := card.Validate(); len(errors) > 0 {
    for _, err := range errors {
        log.Printf("Validation error: %v", err)
    }
}
```

## Version Support

The package tracks which Adaptive Cards version introduced each feature using `version:"x.y"` struct tags:

```go
type ActionSubmit struct {
    Title     string  `json:"title"`
    Tooltip   string  `json:"tooltip,omitempty" version:"1.5"`   // Requires 1.5+
    IsEnabled *bool   `json:"isEnabled,omitempty" version:"1.5"` // Requires 1.5+
    Mode      string  `json:"mode,omitempty" version:"1.5"`      // Requires 1.5+
}
```

### Feature Versions

| Version | Notable Features |
|---------|-----------------|
| 1.0 | Basic elements, actions, and containers |
| 1.1 | Height property, SelectAction |
| 1.2 | BackgroundImage, MinHeight, Bleed, FontType, Fallback |
| 1.3 | RGBA colors, TargetWidth, Input validation (ErrorMessage, IsRequired, Label) |
| 1.5 | RTL support, Tooltip, IsEnabled, Mode, Style (heading), Table element |
| 1.6 | Metadata, CaptionSource, Dynamic choice filtering (Data.Query), Refresh.Expires, Input label positioning |

## Validation

### Automatic Validation

Use the `Validate()` method on any `AdaptiveCard`:

```go
card := adaptivecards.AdaptiveCard{
    Version: "1.0",
    RTL:     boolPtr(true), // ERROR! RTL requires version 1.5
}

errors := card.Validate()
// Returns: "AdaptiveCard.RTL requires version 1.5 but card version is 1.0"
```

### Manual Validation

You can also validate any struct with version tags:

```go
errors := adaptivecards.ValidateVersion(myStruct, "1.2")
```

### Validation Errors

`ValidationError` provides detailed information:

```go
type ValidationError struct {
    Field           string  // e.g., "AdaptiveCard.RTL"
    RequiredVersion Version // e.g., "1.5"
    CardVersion     Version // e.g., "1.0"
}

err := errors[0]
fmt.Printf("%s requires version %s but card version is %s\n",
    err.Field, err.RequiredVersion, err.CardVersion)
```

## Version Comparison

Compare versions programmatically:

```go
v1 := adaptivecards.MustParseVersion("1.5")
v2 := adaptivecards.MustParseVersion("1.2")

v1.Compare(v2)  // Returns: 1 (v1 > v2)
v1.SupportsVersion(v2) // Returns: true (1.5 supports all 1.2 features)
```

## Examples

### Example 1: Simple 1.0 Card

```go
card := adaptivecards.AdaptiveCard{
    Version: "1.0",
    Body: []adaptivecards.Element{
        adaptivecards.TextBlock{
            Text:   "Hello World",
            Size:   stringPtr("large"),
            Weight: stringPtr("bolder"),
        },
    },
}

if errors := card.Validate(); len(errors) > 0 {
    // Handle validation errors
}
```

### Example 2: Advanced 1.5 Card

```go
card := adaptivecards.AdaptiveCard{
    Version:   "1.5",
    MinHeight: "100px",
    RTL:       boolPtr(true),
    Body: []adaptivecards.Element{
        adaptivecards.TextBlock{
            Text:  "Advanced Card",
            Style: stringPtr("heading"),
        },
        adaptivecards.Table{
            FirstRowAsHeader: true,
            Columns: []adaptivecards.TableColumnDefinition{
                {Width: "auto"},
                {Width: "stretch"},
            },
            Rows: []adaptivecards.TableRow{
                {Cells: []adaptivecards.TableCell{
                    {Items: []adaptivecards.Element{adaptivecards.TextBlock{Text: "Name"}}},
                    {Items: []adaptivecards.Element{adaptivecards.TextBlock{Text: "Value"}}},
                }},
            },
        },
    },
}

if errors := card.Validate(); len(errors) == 0 {
    // Card is valid - safe to send
}
```

### Example 3: Catching Errors

```go
// Card claims to be 1.0 but uses 1.5 features
card := adaptivecards.AdaptiveCard{
    Version:   "1.0",
    RTL:       boolPtr(true), // Requires 1.5!
    Body: []adaptivecards.Element{
        adaptivecards.TextBlock{
            Text:  "Invalid",
            Style: stringPtr("heading"), // Requires 1.5!
        },
    },
}

errors := card.Validate()
// Returns multiple errors:
// - "AdaptiveCard.RTL requires version 1.5 but card version is 1.0"
// - "TextBlock.Style requires version 1.5 but card version is 1.0"
```

## Best Practices

1. **Always validate cards before sending**
   ```go
   if errors := card.Validate(); len(errors) > 0 {
       return fmt.Errorf("invalid card: %v", errors[0])
   }
   ```

2. **Use the lowest version possible**
   - More clients support older versions
   - Only bump version when you need new features

3. **Test with different versions**
   ```go
   testVersions := []string{"1.0", "1.2", "1.5"}
   for _, v := range testVersions {
       card.Version = v
       errors := card.Validate()
       // Verify expected behavior
   }
   ```

4. **Handle validation errors gracefully**
   ```go
   if errors := card.Validate(); len(errors) > 0 {
       log.Printf("Card validation failed:")
       for _, err := range errors {
           log.Printf("  - %s", err.Error())
       }
       // Fallback to simpler card or return error
   }
   ```

## Running the Examples

See [examples/validation_example.go](examples/validation_example.go) for complete working examples:

```bash
go run pkg/adaptivecards/examples/validation_example.go
```

## References

- [Adaptive Cards Schema](https://adaptivecards.io/schemas/adaptive-card.json)
- [Adaptive Cards Versioning](https://docs.microsoft.com/en-us/adaptive-cards/authoring-cards/versioning)
- [Schema Explorer](https://adaptivecards.io/explorer/)
