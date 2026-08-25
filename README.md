# Core SDK

> Shared Go infrastructure library for AWS, messaging, observability and reusable backend components.

## Overview

Core SDK is a reusable Go library created to centralize common infrastructure components used across backend services.

The goal is to reduce code duplication, standardize integrations and provide consistent building blocks for Go-based applications, particularly services running in cloud-native and distributed environments.

The SDK provides components for:

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
```

## Components

### AWS

Provides common AWS configuration and credential handling based on the AWS SDK for Go v2.

The configuration abstraction supports:

- AWS Region
- Custom AWS endpoint
- Access Key ID
- Secret Access Key

Custom endpoints also make the package suitable for local development environments using AWS-compatible services.

### Amazon SNS

Provides a reusable publisher abstraction for publishing application events to Amazon SNS topics.

The publisher encapsulates AWS SDK interaction and standardizes event publication across services.

### Amazon SQS

Provides reusable producers and consumers for Amazon SQS.

The consumer supports:

- Long polling
- Configurable visibility timeout
- Configurable message batch size
- Concurrent workers using goroutines
- Message handlers
- Dead Letter Queue (DLQ) handling
- Receive-count based retry handling
- Structured logging

The publisher supports both individual and batch message publishing.

#### SQS Consumer Architecture

```text
                 ┌──────────────┐
                 │   SQS Queue  │
                 └──────┬───────┘
                        │
                  Long Polling
                        │
                 ┌──────▼───────┐
                 │ SQS Consumer │
                 └──────┬───────┘
                        │
              ┌─────────┼─────────┐
              │         │         │
          ┌───▼───┐ ┌──▼────┐ ┌──▼────┐
          │Worker │ │Worker │ │Worker │
          │  #1   │ │  #2   │ │  #N   │
          └───┬───┘ └──┬────┘ └──┬────┘
              │         │         │
              └─────────┼─────────┘
                        │
                  Message Handler
                        │
                 ┌──────▼───────┐
                 │   Success?   │
                 └──────┬───────┘
                    ┌───┴───┐
                   YES      NO
                    │        │
                 Delete    Retry
                              │
                       Max attempts?
                         ┌────┴────┐
                        NO         YES
                        │           │
                      Retry        DLQ
```

The consumer implements concurrent message processing using Go goroutines while providing configurable retry and failure handling.

### Structured Logging

Provides a centralized logging abstraction based on Uber Zap.

Structured fields can be attached to application logs to improve diagnostics and correlation across distributed services.

### OpenTelemetry

Provides initialization and lifecycle management for OpenTelemetry tracing and metrics.

The SDK supports OTLP-based telemetry export and provides common configuration for:

- Trace provider
- Meter provider
- Trace context propagation
- Batch span processing
- Periodic metric collection

### Distributed Tracing

The tracing package provides a higher-level abstraction over OpenTelemetry spans.

It supports:

- Span creation
- Custom attributes
- Events
- Error recording
- Trace ID and Span ID access
- Correlation identifiers
- Integration between tracing and structured logging

Application context such as correlation and organization identifiers can be associated with spans.

### Expression Evaluation

Provides a reusable abstraction for evaluating expressions dynamically using the `expr` library, including error handling and tests.

## Design Goals

The project was created around the following principles:

- **Reusability** — common infrastructure should not be duplicated across services.
- **Consistency** — services should use standardized integrations and behaviors.
- **Separation of concerns** — infrastructure details remain encapsulated.
- **Cloud-native development** — provide components commonly required by distributed backend services.
- **Observability** — logging, metrics and tracing are first-class concerns.
- **Go idioms** — keep APIs simple and leverage Go's standard patterns.

## Technology Stack

| Technology | Purpose |
|---|---|
| Go | Core language |
| AWS SDK for Go v2 | AWS integration |
| Amazon SNS | Event publishing |
| Amazon SQS | Asynchronous messaging |
| OpenTelemetry | Metrics and distributed tracing |
| OTLP | Telemetry export |
| Zap | Structured logging |
| Echo | HTTP-related infrastructure |
| expr | Expression evaluation |
| Testify | Testing |

## Project Structure

### `internal/`

Contains implementation details that should not be imported directly by external packages.

### `pkg/`

Contains reusable packages intended to be consumed by backend services.

Each package focuses on a specific infrastructure concern, such as AWS, messaging, logging or observability.

## Getting Started

### Requirements

- Go 1.26+

### Clone

```bash
git clone https://github.com/clodoaldomarques/core-sdk.git
cd core-sdk
```

### Download dependencies

```bash
go mod download
```

### Run tests

```bash
go test ./...
```

### Run tests with race detection

```bash
go test -race ./...
```

## Use Cases

Core SDK is particularly useful as a foundation for Go backend services that require common infrastructure such as:

- REST APIs
- Microservices
- Event-driven applications
- Asynchronous workers
- AWS integrations
- Distributed tracing
- Structured logging
- Cloud-native services

## Known Trade-offs

### SQS to DLQ

When a message is moved manually to a Dead Letter Queue, publishing to the DLQ and deleting the message from the original queue are separate operations.

If publishing succeeds but deletion fails, the original message may be processed again.

This is a distributed-systems trade-off and highlights the importance of idempotent message processing and explicit delivery-semantics decisions.

## Project Status

This project is primarily intended as a reusable foundation for Go backend projects and as a laboratory for exploring infrastructure components and patterns commonly used in distributed systems.

## Author

**Clodoaldo Marques**

Backend Software Engineer focused on Go, Microservices, Distributed Systems and Cloud-Native architectures.

- GitHub: https://github.com/clodoaldomarques
- LinkedIn: https://www.linkedin.com/in/clodoaldomarques/