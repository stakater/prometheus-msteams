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

// Package main demonstrates how to use the adaptivecards package to create and
// validate Adaptive Cards with different versions and features.
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/stakater/prometheus-msteams/pkg/adaptivecards"
)

func main() {
	// Example 1: Create a version 1.0 compatible card
	example1()

	// Example 2: Create a version 1.5 card with advanced features
	example2()

	// Example 3: Validate a card before sending
	example3()

	// Example 4: Show version compatibility errors
	example4()
}

func example1() {
	fmt.Println("\n=== Example 1: Version 1.0 Compatible Card ===")

	card := adaptivecards.AdaptiveCard{
		Version: "1.0",
		Body: []adaptivecards.Element{
			adaptivecards.TextBlock{
				Text:   adaptivecards.AsPtr("Hello World"),
				Size:   adaptivecards.AsPtr(adaptivecards.FontSizeLarge),
				Weight: adaptivecards.AsPtr(adaptivecards.FontWeightBolder),
			},
			adaptivecards.TextBlock{
				Text: adaptivecards.AsPtr("This is a simple card compatible with version 1.0"),
				Wrap: true,
			},
		},
		Actions: []adaptivecards.Action{
			adaptivecards.ActionOpenURL{
				CommonActionProperties: &adaptivecards.CommonActionProperties{
					Title: "Learn More",
				},
				URL: "https://adaptivecards.io",
			},
		},
	}

	// Using the convenience method
	if errors := card.Validate(); len(errors) > 0 {
		fmt.Println("Validation errors:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err.Error())
		}
		return
	}

	fmt.Println("✓ Card is valid for version", card.Version)
	printCard(card)
}

func example2() {
	fmt.Println("\n=== Example 2: Version 1.5 Card with Advanced Features ===")

	card := adaptivecards.AdaptiveCard{
		Version:   "1.5",
		MinHeight: "100px",                   // Requires 1.2
		RTL:       adaptivecards.AsPtr(true), // Requires 1.5
		Body: []adaptivecards.Element{
			adaptivecards.TextBlock{
				Text:  adaptivecards.AsPtr("Advanced Card"),
				Style: adaptivecards.AsPtr(adaptivecards.TextBlockStyleHeading), // Requires 1.5
			},
			adaptivecards.Container{
				Items: []adaptivecards.Element{
					adaptivecards.TextBlock{
						Text:     adaptivecards.AsPtr("Container with bleed"),
						FontType: adaptivecards.AsPtr(adaptivecards.FontTypeMonospace), // Requires 1.2
					},
				},
				Bleed:     true,                       // Requires 1.2
				RTL:       adaptivecards.AsPtr(false), // Requires 1.5
				MinHeight: "50px",                     // Requires 1.2
			},
			adaptivecards.Table{ // Requires 1.5
				FirstRowAsHeader: true,
				Columns: []adaptivecards.TableColumnDefinition{
					{Width: "auto"},
					{Width: "stretch"},
				},
				Rows: []adaptivecards.TableRow{
					{
						Cells: []adaptivecards.TableCell{
							{Items: []adaptivecards.Element{adaptivecards.TextBlock{Text: adaptivecards.AsPtr("Name")}}},
							{Items: []adaptivecards.Element{adaptivecards.TextBlock{Text: adaptivecards.AsPtr("Value")}}},
						},
					},
					{
						Cells: []adaptivecards.TableCell{
							{Items: []adaptivecards.Element{adaptivecards.TextBlock{Text: adaptivecards.AsPtr("Version")}}},
							{Items: []adaptivecards.Element{adaptivecards.TextBlock{Text: adaptivecards.AsPtr("1.5")}}},
						},
					},
				},
			},
		},
		Actions: []adaptivecards.Action{
			adaptivecards.ActionSubmit{
				CommonActionProperties: &adaptivecards.CommonActionProperties{
					Title:     "Submit",
					Tooltip:   "Submit the form",
					IsEnabled: adaptivecards.AsPtr(true),
					Mode:      adaptivecards.AsPtr(adaptivecards.ActionModePrimary),
				},
			},
		},
	}

	// Using the convenience method
	if errors := card.Validate(); len(errors) > 0 {
		fmt.Println("Validation errors:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err.Error())
		}
		return
	}

	fmt.Println("✓ Card is valid for version", card.Version)
	printCard(card)
}

func example3() {
	fmt.Println("\n=== Example 3: Validate Before Sending ===")

	card := adaptivecards.AdaptiveCard{
		Version: "1.2",
		Body: []adaptivecards.Element{
			adaptivecards.TextBlock{
				Text:     adaptivecards.AsPtr("Valid 1.2 card"),
				FontType: adaptivecards.AsPtr(adaptivecards.FontTypeMonospace), // OK - introduced in 1.2
			},
			adaptivecards.Container{
				Items: []adaptivecards.Element{
					adaptivecards.TextBlock{Text: adaptivecards.AsPtr("Content")},
				},
				Bleed:     true,   // OK - introduced in 1.2
				MinHeight: "50px", // OK - introduced in 1.2
			},
		},
	}

	fmt.Printf("Validating card version %s...\n", card.Version)

	// Using the convenience method
	if errors := card.Validate(); len(errors) > 0 {
		fmt.Println("❌ Validation failed:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err.Error())
		}
	} else {
		fmt.Println("✓ Card validation passed! Safe to send.")
	}
}

func example4() {
	fmt.Println("\n=== Example 4: Version Compatibility Errors ===")

	// Create a card that claims to be 1.0 but uses 1.5 features
	card := adaptivecards.AdaptiveCard{
		Version:   "1.0",
		MinHeight: "100px",                   // ERROR - requires 1.2
		RTL:       adaptivecards.AsPtr(true), // ERROR - requires 1.5
		Body: []adaptivecards.Element{
			adaptivecards.TextBlock{
				Text:  adaptivecards.AsPtr("This won't validate"),
				Style: adaptivecards.AsPtr(adaptivecards.TextBlockStyleHeading), // ERROR - requires 1.5
			},
			adaptivecards.Table{ // ERROR - Table type requires 1.5
				Rows: []adaptivecards.TableRow{
					{Cells: []adaptivecards.TableCell{}},
				},
			},
		},
	}

	fmt.Printf("Validating card version %s with incompatible features...\n", card.Version)

	// Using the convenience method
	errors := card.Validate()

	if len(errors) > 0 {
		fmt.Printf("❌ Found %d compatibility issues:\n", len(errors))
		for i, err := range errors {
			fmt.Printf("  %d. %s\n", i+1, err.Error())
		}
		fmt.Println("\nRecommendation: Update card version to 1.5 to use these features")
	}
}

// Helper functions

func printCard(card adaptivecards.AdaptiveCard) {
	data, err := json.MarshalIndent(card, "", "  ")
	if err != nil {
		log.Printf("Error marshaling card: %v", err)
		return
	}
	fmt.Printf("\nCard JSON:\n%s\n", string(data))
}
