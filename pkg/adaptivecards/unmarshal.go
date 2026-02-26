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
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// region Type Registry for Dynamic Unmarshaling

// Type registries for dynamic unmarshaling
var typeRegistry = make(map[string]reflect.Type)

// RegisterType registers a type with its JSON type name for dynamic unmarshaling
func RegisterType(typeName string, example any) {
	typeRegistry[typeName] = reflect.TypeOf(example)
}

// endregion Type Registry for Dynamic Unmarshaling

// region Validation

// FindMapKey is a helper function to find the key in a map given a value.
// It returns the key and a boolean indicating if the value was found.
func FindMapKey[K comparable, V comparable](m map[K]V, value V) (key K, ok bool) {
	for k, v := range m {
		if v == value {
			key = k
			ok = true
			return
		}
	}
	return
}

// endregion Validation

// AdaptiveCardType is a helper struct to detect the "type" field in the JSON
// payload for dynamic unmarshaling.
type AdaptiveCardType struct {
	Type string `json:"type"`
}

// Custom error types for better error handling in dynamic unmarshaling

// AdaptiveCardErrorInvalid is returned when the "type" field in the JSON does
// not match the expected type for a struct.
type AdaptiveCardErrorInvalid struct {
	Original any
	Card     AdaptiveCardType
}

// Error implements the error interface for AdaptiveCardErrorInvalid.
func (e AdaptiveCardErrorInvalid) Error() string {
	return fmt.Sprintf("invalid type: %s for %T", e.Card.Type, e.Original)
}

// AdaptiveCardErrorUnknown is returned when the "type" field in the JSON does
// not match any known type in the registry.
type AdaptiveCardErrorUnknown struct {
	Card AdaptiveCardType
}

// Error implements the error interface for AdaptiveCardErrorUnknown.
func (e AdaptiveCardErrorUnknown) Error() string {
	return fmt.Sprintf("unknown type: %s", e.Card.Type)
}

// ValidateTypes validates a slice of json.RawMessage against a source type.
// It checks if each item in the slice can be unmarshaled into the source type
// based on the "type" field in the JSON and the type registry.
func ValidateTypes(data []json.RawMessage, sourceType reflect.Type) error {
	errs := []error{}
	kind := sourceType.Kind()
	if kind == reflect.Slice {
		sourceType = sourceType.Elem()
	}

	for i, raw := range data {
		if err := ValidateType(raw, sourceType); err != nil {
			errs = append(errs, fmt.Errorf("validation failed for item at index %d: %w", i, err))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// ValidateType validates a single json.RawMessage against a source type.
// It checks if the item can be unmarshaled into the source type based on the
// "type" field in the JSON and the type registry.
func ValidateType(raw json.RawMessage, sourceType reflect.Type) error {
	tc := &AdaptiveCardType{}
	if err := json.Unmarshal(raw, tc); err != nil {
		return errors.Join(errors.New("failed to detect element type"), err)
	}

	kind := sourceType.Kind()
	if kind == reflect.Interface {
		return fmt.Errorf("can't validate against interface, defer to dynamic unmarshaling")
	}
	if kind == reflect.Pointer {
		sourceType = sourceType.Elem()
	}

	typeStr, ok := FindMapKey(typeRegistry, sourceType)
	if !ok {
		return fmt.Errorf("unable to find type for %T in registry! Is it registered?", sourceType)
	}
	if tc.Type != "" && tc.Type != typeStr {
		return AdaptiveCardErrorInvalid{Original: sourceType, Card: *tc}
	}
	return nil
}

// region Dynamic Marshaling

// SmartMarshalFromJSON injects a "type" field into any struct during JSON marshaling.
// It uses reflection to extract struct fields and build a map, avoiding the infinite
// recursion that would occur if we called json.Marshal directly on the value.
func SmartMarshalFromJSON(v any) ([]byte, error) {
	sourceValue := reflect.ValueOf(v)
	sourceType := sourceValue.Type()

	if sourceValue.Kind() == reflect.Pointer {
		sourceType = sourceType.Elem()
	}

	typeStr, ok := FindMapKey(typeRegistry, sourceType)
	if !ok {
		return nil, fmt.Errorf("unable to find type for %T in registry! Is it registered?", sourceType)
	}

	return marshalWithType(v, typeStr)
}

// marshalWithType injects a "type" field into any struct during JSON marshaling.
// It uses reflection to extract struct fields and build a map, avoiding the infinite
// recursion that would occur if we called json.Marshal directly on the value.
func marshalWithType(v any, typeName string) ([]byte, error) {
	// Create a map to hold all fields plus the type
	m := make(map[string]any)
	m["type"] = typeName

	// Use reflection to extract struct fields
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	// Handle pointer types by dereferencing
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return json.Marshal(m)
		}
		val = val.Elem()
		typ = typ.Elem()
	}

	// Ensure we're working with a struct
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("marshalWithType requires a struct, got %v", val.Kind())
	}

	// Extract all exported fields
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Handle embedded structs (like *Common, *CommonActionProperties)
		if field.Anonymous {
			// If it's a pointer to a struct, dereference it
			if fieldValue.Kind() == reflect.Pointer && !fieldValue.IsNil() {
				fieldValue = fieldValue.Elem()
			}

			// If it's a struct, add all its fields to our map
			if fieldValue.Kind() == reflect.Struct {
				embeddedType := fieldValue.Type()
				for j := 0; j < fieldValue.NumField(); j++ {
					embeddedField := embeddedType.Field(j)
					embeddedFieldValue := fieldValue.Field(j)

					if !embeddedField.IsExported() {
						continue
					}

					jsonTag := embeddedField.Tag.Get("json")
					if jsonTag == "" || jsonTag == "-" {
						continue
					}

					// Parse the json tag (format: "name,omitempty")
					tagParts := strings.Split(jsonTag, ",")
					fieldName := tagParts[0]
					omitEmpty := len(tagParts) > 1 && tagParts[1] == "omitempty"

					// Skip if omitempty and the field is empty
					if omitEmpty && isEmptyValue(embeddedFieldValue) {
						continue
					}

					m[fieldName] = embeddedFieldValue.Interface()
				}
			}
			continue
		}

		// Get the JSON tag for this field
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse the json tag (format: "name,omitempty")
		tagParts := strings.Split(jsonTag, ",")
		fieldName := tagParts[0]
		omitEmpty := len(tagParts) > 1 && tagParts[1] == "omitempty"

		// Skip if omitempty and the field is empty
		if omitEmpty && isEmptyValue(fieldValue) {
			continue
		}

		// Add the field to the map
		m[fieldName] = fieldValue.Interface()
	}

	// Marshal the final map
	return json.Marshal(m)
}

// isEmptyValue checks if a reflect.Value is empty (zero value)
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	}
	return false
}

// endregion Dynamic Marshaling

// region Dynamic Unmarshaling

// parseJSONTag extracts the field name from a JSON struct tag
func parseJSONTag(tag string) string {
	if tag == "" || tag == "-" {
		return ""
	}
	// Find comma separator for options like "name,omitempty"
	if idx := strings.IndexByte(tag, ','); idx != -1 {
		return tag[:idx]
	}
	return tag
}

// unmarshalFallback handles the special fallback field which can be a string, Element, or Action
func unmarshalFallback(payload json.RawMessage, data *any) error {
	// Try string first (most common case)
	var fallbackStr string
	if err := json.Unmarshal(payload, &fallbackStr); err == nil {
		*data = FallbackOption(fallbackStr)
		return nil
	}

	// Try Element
	var fallbackElement Element
	if err := unmarshalMessage(payload, &fallbackElement); err == nil {
		*data = fallbackElement
		return nil
	}

	// Try Action
	var fallbackAction Action
	if err := unmarshalMessage(payload, &fallbackAction); err == nil {
		*data = fallbackAction
		return nil
	}

	return fmt.Errorf("fallback must be either FallbackOption, Element, or Action")
}

// unmarshalMessages unmarshals a slice of raw JSON messages into a typed slice
// Unknown types are skipped for forward/backward compatibility
func unmarshalMessages[T any](payload []json.RawMessage, data *[]T) error {
	result := make([]T, 0, len(payload))
	for i, rawData := range payload {
		item, err := dynamicUnmarshaling[T](rawData)
		if err != nil {
			// Skip unknown types for forward/backward compatibility
			var unknownErr AdaptiveCardErrorUnknown
			if errors.As(err, &unknownErr) {
				continue
			}
			return fmt.Errorf("failed to unmarshal type %T at index %d: %w", *data, i, err)
		}
		result = append(result, item)
	}
	*data = result
	return nil
}

// unmarshalMessage unmarshals a single raw JSON message into a typed value
// Unknown types are ignored for forward/backward compatibility
func unmarshalMessage[T any](payload json.RawMessage, data *T) error {
	result, err := dynamicUnmarshaling[T](payload)
	if err != nil {
		// Ignore unknown types for forward/backward compatibility
		var unknownErr AdaptiveCardErrorUnknown
		if errors.As(err, &unknownErr) {
			return nil
		}
		return fmt.Errorf("failed to unmarshal %T: %w", data, err)
	}
	*data = result
	return nil
}

func getCardType(data json.RawMessage) (*AdaptiveCardType, error) {
	typeCheck := &AdaptiveCardType{}
	if err := json.Unmarshal(data, typeCheck); err != nil {
		return nil, errors.Join(errors.New("failed to detect element type"), err)
	}
	return typeCheck, nil
}

func reflectUnmarshaling(data json.RawMessage) (reflect.Value, error) {
	zero := reflect.Value{}
	typeCheck, err := getCardType(data)
	if err != nil {
		return zero, err
	}

	// Look up the type in the registry
	var regPtr reflect.Value
	if regType, ok := typeRegistry[typeCheck.Type]; ok {
		regPtr = reflect.New(regType)
	} else {
		return zero, AdaptiveCardErrorUnknown{*typeCheck}
	}

	// Unmarshal into the new instance
	if err := json.Unmarshal(data, regPtr.Interface()); err != nil {
		return zero, errors.Join(fmt.Errorf("failed to unmarshal %s", typeCheck.Type), err)
	}
	// Return the pointer value instead of dereferencing
	// This ensures that when unmarshaling into interface slices ([]Element, []Action),
	// we get pointers (*Image, *TextBlock) instead of value types
	return regPtr, nil
}

func dynamicUnmarshaling[T any](data json.RawMessage) (T, error) {
	var zero T
	regPtr, err := reflectUnmarshaling(data)
	if err != nil {
		return zero, err
	}

	// Convert to T interface
	elem, ok := regPtr.Interface().(T)
	if !ok {
		typeCheck, _ := getCardType(data)
		return zero, fmt.Errorf("type %s does not implement interface", typeCheck.Type)
	}

	return elem, nil
}

// unmarshalSlice dynamically unmarshals a slice field based on its element type
func unmarshalSlice(rawMessages []json.RawMessage, fieldValue reflect.Value) error {
	elemType := fieldValue.Type().Elem()

	// Look up the type in the registry
	slice := reflect.MakeSlice(fieldValue.Type(), len(rawMessages), len(rawMessages))
	for i, raw := range rawMessages {
		elem := reflect.New(elemType)
		if err := json.Unmarshal(raw, elem.Interface()); err != nil {
			return errors.Join(fmt.Errorf("failed to unmarshal slice element %d of type %s", i, elemType), err)
		}
		slice.Index(i).Set(elem.Elem())
	}
	fieldValue.Set(slice)
	return nil
}

func unmarshalSingle(rawMessage json.RawMessage, fieldValue reflect.Value) error {
	fieldType := fieldValue.Type()

	elem := reflect.New(fieldType)
	if err := json.Unmarshal(rawMessage, elem.Interface()); err != nil {
		return errors.Join(fmt.Errorf("failed to unmarshal %s", fieldType), err)
	}
	fieldValue.Set(elem.Elem())
	return nil
}

// interfaceUnmarshaler defines known polymorphic interface types
type interfaceUnmarshaler struct {
	interfaceType reflect.Type
	unmarshal     func([]json.RawMessage) (reflect.Value, error)
}

var sliceUnmarshalers = []interfaceUnmarshaler{
	// Element
	{
		interfaceType: reflect.TypeFor[Element](),
		unmarshal: func(raw []json.RawMessage) (reflect.Value, error) {
			var elements []Element
			err := unmarshalMessages(raw, &elements)
			return reflect.ValueOf(elements), err
		},
	},
	// Action
	{
		interfaceType: reflect.TypeFor[Action](),
		unmarshal: func(raw []json.RawMessage) (reflect.Value, error) {
			var actions []Action
			err := unmarshalMessages(raw, &actions)
			return reflect.ValueOf(actions), err
		},
	},
	// ActionData
	{
		interfaceType: reflect.TypeFor[ActionData](),
		unmarshal: func(raw []json.RawMessage) (reflect.Value, error) {
			var actiondata []ActionData
			err := unmarshalMessages(raw, &actiondata)
			return reflect.ValueOf(actiondata), err
		},
	},
	// Reference
	{
		interfaceType: reflect.TypeFor[Reference](),
		unmarshal: func(raw []json.RawMessage) (reflect.Value, error) {
			var references []Reference
			err := unmarshalMessages(raw, &references)
			return reflect.ValueOf(references), err
		},
	},
	// Layout
	{
		interfaceType: reflect.TypeFor[Layout](),
		unmarshal: func(raw []json.RawMessage) (reflect.Value, error) {
			var layouts []Layout
			err := unmarshalMessages(raw, &layouts)
			return reflect.ValueOf(layouts), err
		},
	},
	// RichTextInline
	{
		interfaceType: reflect.TypeFor[RichTextInline](),
		unmarshal: func(raw []json.RawMessage) (reflect.Value, error) {
			var richTexts []RichTextInline
			err := unmarshalMessages(raw, &richTexts)
			return reflect.ValueOf(richTexts), err
		},
	},
}

var singleUnmarshalers = []struct {
	interfaceType reflect.Type
	unmarshal     func(json.RawMessage) (reflect.Value, error)
}{
	// Element
	{
		interfaceType: reflect.TypeFor[Element](),
		unmarshal: func(raw json.RawMessage) (reflect.Value, error) {
			var element Element
			err := unmarshalMessage(raw, &element)
			return reflect.ValueOf(element), err
		},
	},
	// Action
	{
		interfaceType: reflect.TypeFor[Action](),
		unmarshal: func(raw json.RawMessage) (reflect.Value, error) {
			var action Action
			err := unmarshalMessage(raw, &action)
			return reflect.ValueOf(action), err
		},
	},
	// ActionData
	{
		interfaceType: reflect.TypeFor[ActionData](),
		unmarshal: func(raw json.RawMessage) (reflect.Value, error) {
			var action ActionData
			err := unmarshalMessage(raw, &action)
			return reflect.ValueOf(action), err
		},
	},
	// Reference
	{
		interfaceType: reflect.TypeFor[Reference](),
		unmarshal: func(raw json.RawMessage) (reflect.Value, error) {
			var reference Reference
			err := unmarshalMessage(raw, &reference)
			return reflect.ValueOf(reference), err
		},
	},
	// Layout
	{
		interfaceType: reflect.TypeFor[Layout](),
		unmarshal: func(raw json.RawMessage) (reflect.Value, error) {
			var layout Layout
			err := unmarshalMessage(raw, &layout)
			return reflect.ValueOf(layout), err
		},
	},
	// RichTextInline
	{
		interfaceType: reflect.TypeFor[RichTextInline](),
		unmarshal: func(raw json.RawMessage) (reflect.Value, error) {
			var richText RichTextInline
			err := unmarshalMessage(raw, &richText)
			return reflect.ValueOf(richText), err
		},
	},
}

// unmarshalSliceField dynamically unmarshals a slice field based on its element type
// It handles polymorphic slices like []Element, []Action, []Reference, and []Layout
func unmarshalSliceField(rawMessages []json.RawMessage, fieldValue reflect.Value) error {
	// Try concrete type first (optimized path)
	if err := ValidateTypes(rawMessages, fieldValue.Type()); err == nil {
		return unmarshalSlice(rawMessages, fieldValue)
	}

	// Check if element type implements any known interface
	elemType := fieldValue.Type().Elem()
	for _, unmarshaler := range sliceUnmarshalers {
		if elemType.Implements(unmarshaler.interfaceType) {
			value, err := unmarshaler.unmarshal(rawMessages)
			if err != nil {
				return err
			}
			fieldValue.Set(value)
			return nil
		}
	}

	// Fallback to concrete type unmarshaling
	return unmarshalSlice(rawMessages, fieldValue)
}

// unmarshalSingleField dynamically unmarshals a single field based on its type
// It handles polymorphic fields like Element, Action, Reference, and Layout
func unmarshalSingleField(rawMessage json.RawMessage, fieldValue reflect.Value) error {
	// Try concrete type first (optimized path)
	if err := ValidateType(rawMessage, fieldValue.Type()); err == nil {
		return unmarshalSingle(rawMessage, fieldValue)
	}

	// Check if field type implements any known interface
	fieldType := fieldValue.Type()
	for _, unmarshaler := range singleUnmarshalers {
		if fieldType.Implements(unmarshaler.interfaceType) {
			value, err := unmarshaler.unmarshal(rawMessage)
			if err != nil {
				return err
			}
			// Check if the returned value is valid before setting
			if !value.IsValid() {
				// Unknown type was ignored, skip setting
				return nil
			}
			fieldValue.Set(value)
			return nil
		}
	}

	// Fallback to concrete type unmarshaling
	return unmarshalSingle(rawMessage, fieldValue)
}

// hasEmbeddedFields checks if any fields from an embedded struct are present in the JSON
func hasEmbeddedFields(embeddedType reflect.Type, rawFields map[string]json.RawMessage) bool {
	for i := 0; i < embeddedType.NumField(); i++ {
		jsonTag := embeddedType.Field(i).Tag.Get("json")
		if fieldName := parseJSONTag(jsonTag); fieldName != "" {
			if _, exists := rawFields[fieldName]; exists {
				return true
			}
		}
	}
	return false
}

// unmarshalEmbeddedFields unmarshals all fields from an embedded struct
func unmarshalEmbeddedFields(embeddedType reflect.Type, embeddedValue reflect.Value, rawFields map[string]json.RawMessage) error {
	for i := 0; i < embeddedType.NumField(); i++ {
		if err := smartUnmarshalSingleField(rawFields, embeddedType, embeddedValue, i); err != nil {
			return err
		}
	}
	return nil
}

// smartUnmarshalAnonymousField handles embedded/anonymous struct fields
func smartUnmarshalAnonymousField(rawFields map[string]json.RawMessage, targetElem reflect.Value, field int) error {
	fieldValue := targetElem.Field(field)

	if fieldValue.Kind() == reflect.Pointer {
		embeddedType := fieldValue.Type().Elem()

		// Only initialize pointer if at least one field exists in JSON
		if !hasEmbeddedFields(embeddedType, rawFields) {
			return nil
		}

		if fieldValue.IsNil() {
			fieldValue.Set(reflect.New(embeddedType))
		}
		return unmarshalEmbeddedFields(embeddedType, fieldValue.Elem(), rawFields)
	}

	// Non-pointer embedded struct - always unmarshal
	return unmarshalEmbeddedFields(fieldValue.Type(), fieldValue, rawFields)
}

// unmarshalSliceFieldFromRaw unmarshals a slice field from raw JSON
func unmarshalSliceFieldFromRaw(rawData json.RawMessage, fieldValue reflect.Value, fieldName string) error {
	var rawMessages []json.RawMessage
	if err := json.Unmarshal(rawData, &rawMessages); err != nil {
		// Not a slice of objects, use standard unmarshaling
		return json.Unmarshal(rawData, fieldValue.Addr().Interface())
	}

	// Use dynamic unmarshaling for polymorphic slices
	if err := unmarshalSliceField(rawMessages, fieldValue); err != nil {
		return fmt.Errorf("failed to unmarshal slice field %s: %w", fieldName, err)
	}
	return nil
}

// unmarshalInterfaceFieldFromRaw unmarshals an interface field from raw JSON
func unmarshalInterfaceFieldFromRaw(rawData json.RawMessage, fieldValue reflect.Value, fieldName string) error {
	// Validate fieldValue before proceeding
	if !fieldValue.IsValid() {
		return fmt.Errorf("field %s has invalid reflect.Value", fieldName)
	}
	if !fieldValue.CanSet() {
		return fmt.Errorf("field %s cannot be set", fieldName)
	}

	// Special case: Fallback field can be string, Element, or Action
	if fieldName == "fallback" {
		var fallbackValue any
		if err := unmarshalFallback(rawData, &fallbackValue); err != nil {
			return fmt.Errorf("failed to unmarshal fallback field %s: %w", fieldName, err)
		}
		fieldValue.Set(reflect.ValueOf(fallbackValue))
		return nil
	}

	// Handle polymorphic interface fields (Element, Action, etc.)
	if err := unmarshalSingleField(rawData, fieldValue); err != nil {
		return fmt.Errorf("failed to unmarshal interface field %s: %w", fieldName, err)
	}
	return nil
}

// unmarshalPointerFieldFromRaw unmarshals a pointer field from raw JSON
func unmarshalPointerFieldFromRaw(rawData json.RawMessage, fieldValue reflect.Value, fieldType reflect.Type, fieldName string) error {
	ptrElem := fieldType.Elem()

	if ptrElem.Kind() == reflect.Interface {
		// Pointer to interface (like *Action)
		newValue := reflect.New(ptrElem)
		if err := unmarshalSingleField(rawData, newValue.Elem()); err != nil {
			return fmt.Errorf("failed to unmarshal pointer-to-interface field %s: %w", fieldName, err)
		}
		fieldValue.Set(newValue)
		return nil
	}

	// Standard pointer field
	newPtr := reflect.New(ptrElem)
	if err := json.Unmarshal(rawData, newPtr.Interface()); err != nil {
		return fmt.Errorf("failed to unmarshal pointer field %s: %w", fieldName, err)
	}
	fieldValue.Set(newPtr)
	return nil
}

// smartUnmarshalSingleField unmarshals a single struct field based on its type
func smartUnmarshalSingleField(rawFields map[string]json.RawMessage, targetType reflect.Type, targetElem reflect.Value, field int) error {
	structField := targetType.Field(field)
	fieldValue := targetElem.Field(field)

	// Skip unexported fields
	if !fieldValue.CanSet() {
		return nil
	}

	// Handle embedded anonymous structs (like *Common)
	if structField.Anonymous {
		return smartUnmarshalAnonymousField(rawFields, targetElem, field)
	}

	// Get field name from JSON tag
	fieldName := parseJSONTag(structField.Tag.Get("json"))
	if fieldName == "" {
		return nil
	}

	// Get raw JSON for this field
	rawData, exists := rawFields[fieldName]
	if !exists {
		return nil
	}

	/*
		logger.Debug(
			"function", "smartUnmarshalSingleField",
			"targetType", targetType.Name(),
			"structField", structField.Name,
			"fieldName", fieldName,
			"fieldType", fieldValue.Type())
	*/

	// Dispatch to appropriate unmarshaler based on field kind
	switch structField.Type.Kind() {
	case reflect.Slice:
		return unmarshalSliceFieldFromRaw(rawData, fieldValue, fieldName)

	case reflect.Interface:
		return unmarshalInterfaceFieldFromRaw(rawData, fieldValue, fieldName)

	case reflect.Pointer:
		return unmarshalPointerFieldFromRaw(rawData, fieldValue, structField.Type, fieldName)

	default:
		// Standard fields (string, int, bool, struct, etc.)
		if err := json.Unmarshal(rawData, fieldValue.Addr().Interface()); err != nil {
			return fmt.Errorf("failed to unmarshal field %s: %w", fieldName, err)
		}
		return nil
	}
}

// SmartUnmarshalJSON is a generic function that automatically handles polymorphic unmarshaling
// for structs containing Element, Action, Reference, or Layout interfaces.
//
// Usage:
//
//	func (a *AdaptiveCard) UnmarshalJSON(data []byte) error {
//	    return SmartUnmarshalJSON(data, a)
//	}
//
// This eliminates the need to manually handle polymorphic fields in every UnmarshalJSON implementation.
func SmartUnmarshalJSON[T any](data []byte, target *T) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Pointer || targetValue.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	targetElem := targetValue.Elem()
	targetType := targetElem.Type()

	if targetType.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to a struct, got %s", targetType.Kind())
	}

	// First, unmarshal into a map to preserve raw JSON for polymorphic fields
	rawFields := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &rawFields); err != nil {
		return err
	}

	// Iterate through struct fields and unmarshal appropriately
	for i := 0; i < targetType.NumField(); i++ {
		if err := smartUnmarshalSingleField(rawFields, targetType, targetElem, i); err != nil {
			return err
		}
	}

	return nil
}

// endregion Dynamic Unmarshaling
