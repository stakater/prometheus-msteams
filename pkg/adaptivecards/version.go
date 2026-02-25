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
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Version represents an Adaptive Card schema version
type Version struct {
	Major int
	Minor int
	// Patch int // Optional patch version, not used in current schema but included for future compatibility
	// Short string // Original version string for error messages
	// PreRelease string // Optional pre-release identifier (e.g. "beta", "rc1")
	// Build string // Optional build metadata (e.g. "20240101")
}

// ParseVersion parses a version string like "1.2" into a Version struct
func ParseVersion(v string) (Version, error) {
	parts := strings.Split(v, ".")
	if len(parts) != 2 {
		return Version{}, fmt.Errorf("invalid version format: %s", v)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	return Version{Major: major, Minor: minor}, nil
}

// MustParseVersion parses a version string and panics if it's invalid.
// This is useful for initializing package-level variables or when you're certain
// the version string is valid.
func MustParseVersion(v string) Version {
	ver, err := ParseVersion(v)
	if err != nil {
		panic(err)
	}
	return ver
}

// String returns the version as a string like "1.2"
func (v Version) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// Compare compares two versions. Returns:
// -1 if v < other
//
//	0 if v == other
//	1 if v > other
func (v Version) Compare(other Version) int {
	if v.Major < other.Major {
		return -1
	}
	if v.Major > other.Major {
		return 1
	}
	if v.Minor < other.Minor {
		return -1
	}
	if v.Minor > other.Minor {
		return 1
	}
	return 0
}

// SupportsVersion checks if this version supports the given version
func (v Version) SupportsVersion(required Version) bool {
	return v.Compare(required) >= 0
}

// ValidationError represents a version compatibility error
type ValidationError struct {
	FieldName       string
	RequiredVersion Version
	CardVersion     Version
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("field %s requires version %s but card version is %s",
		e.FieldName, e.RequiredVersion, e.CardVersion)
}

// ValidateVersion validates that all non-zero fields in a struct are supported by the given version
func ValidateVersion(v any, cardVersion string) []ValidationError {
	cv, err := ParseVersion(cardVersion)
	if err != nil {
		return []ValidationError{{
			FieldName:   "version",
			CardVersion: cv,
		}}
	}

	return validateStruct(reflect.ValueOf(v), cv, "")
}

func validateStruct(val reflect.Value, cardVersion Version, prefix string) []ValidationError {
	var errors []ValidationError

	// Dereference pointers
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		fieldName := fieldType.Name
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}

		// Check if field has a version tag
		versionTag := fieldType.Tag.Get("version")
		if versionTag != "" && !isZeroValue(field) {
			requiredVersion, err := ParseVersion(versionTag)
			if err == nil && !cardVersion.SupportsVersion(requiredVersion) {
				errors = append(errors, ValidationError{
					FieldName:       fieldName,
					RequiredVersion: requiredVersion,
					CardVersion:     cardVersion,
				})
			}
		}

		// Recursively validate nested structs
		switch field.Kind() {
		case reflect.Struct:
			errors = append(errors, validateStruct(field, cardVersion, fieldName)...)
		case reflect.Ptr:
			if !field.IsNil() {
				errors = append(errors, validateStruct(field, cardVersion, fieldName)...)
			}
		case reflect.Slice:
			for j := 0; j < field.Len(); j++ {
				item := field.Index(j)
				itemName := fmt.Sprintf("%s[%d]", fieldName, j)
				errors = append(errors, validateStruct(item, cardVersion, itemName)...)
			}
		}
	}

	return errors
}

func isZeroValue(v reflect.Value) bool {
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
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
