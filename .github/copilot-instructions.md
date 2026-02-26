# Copilot Instructions for Prometheus MS Teams

## Project Overview

A lightweight Go Web Server that receives POST alert messages from Prometheus Alert Manager and sends it to a Microsoft Teams Channel using an incoming webhook URL[s].
The package provides a complete implementation of the Microsoft Adaptive Cards schema, supporting all versions from 1.0 to 1.6. It includes type definitions for all card elements, actions, and containers, along with JSON marshaling/unmarshaling and a robust version validation system.

## Package Structure

- **`cmd`** - Contains the main application entry point and server setup.
- **`pkg`** - Contains the core logic for handling incoming alerts, processing them, and sending messages to Microsoft Teams.
  - **`adaptivecards`** - Implements the Adaptive Cards schema with version validation.
  - **`service`** - Defines the service structure.
  - **`testutils`** - Test utilities and helpers.
  - **`transport`** - Transport layer for handling HTTP[s] requests and responses.
  - **`version`** - Contains versioning information and utilities for the application.
- **`test`** - Contains unit and integration tests for the application.

## Coding Guidelines

- Follow Go conventions for naming, formatting, and structuring code.
- Use clear and descriptive names for variables, functions, and types.
- Write modular and reusable code, breaking down complex functions into smaller, focused ones.
- Include comments and documentation for public functions and types to explain their purpose and usage.
- Ensure proper error handling and logging throughout the application.
- When adding new features, always add tests:
  - Use `github.com/stretchr/testify/assert` for assertions.
  - Use `github.com/onsi/ginkgo/v2/ginkgo` for BDD-style testing.
  - Use `github.com/onsi/gomega` for matcher assertions.
- Always run tests and ensure they pass before committing code.
  - Use `make test` to run all tests in the project.
  - Use `make test-coverage` to run tests with coverage reporting.
- Code must lint without errors using `golangci-lint` and should be formatted using `gofmt`.
  - Use `make lint` to run the linter and `make fmt` to format the code.
