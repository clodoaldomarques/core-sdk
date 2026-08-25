# Core SDK

> Shared Go infrastructure library for building backend services with AWS, messaging, observability and common application components.

## Overview

Core SDK is a reusable Go library created to centralize common infrastructure components used across backend services.

The goal is to reduce duplication, standardize integrations and provide consistent building blocks for Go-based applications, particularly services running in cloud-native and distributed environments.

The SDK currently provides components for:

- AWS configuration
- Amazon SNS
- Amazon SQS
- Structured logging
- OpenTelemetry
- Distributed tracing
- Environment configuration
- Expression evaluation
- Common reusable services and utilities

## Architecture

The project follows a modular package-oriented structure, exposing reusable components through the `pkg` directory while keeping internal implementation details under `internal`.

```text
core-sdk/
├── internal/
│   └── request/
│       └── service.go
│
├── pkg/
│   ├── aws/
│   ├── commons/
│   ├── env/
│   ├── expression/
│   ├── logger/
│   ├── opentelemetry/
│   ├── sns/
│   ├── sqs/
│   └── tracer/
│
├── go.mod
└── go.sum