# Remus

Remus is a flexible and efficient network messaging library for Go that provides a comprehensive protocol for building distributed systems. It supports various message types including control messages, data operations, state management, edge computing, service discovery, and observability.

[![Go Reference](https://pkg.go.dev/badge/github.com/foxycorps/remus-go.svg)](https://pkg.go.dev/github.com/foxycorps/remus-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/foxycorps/remus-go)](https://goreportcard.com/report/github.com/foxycorps/remus-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- **Flexible Message Protocol**: Support for multiple message types with customizable flags, priorities, and TTLs
- **Builder Pattern**: Fluent interface for constructing messages
- **Binary Encoding**: Efficient binary message encoding and decoding
- **State Management**: Built-in support for state synchronization and delta updates
- **Service Discovery**: Service announcement and health check capabilities
- **Edge Computing**: Support for deploying and managing edge functions
- **Observability**: Integrated metrics, tracing, and logging
- **Stream Processing**: Support for streaming data with start/end markers

## Installation

```go
go get github.com/foxycorps/remus-go
```

## Quick Start

### Basic Message Creation

```go
// Create a simple request message
requestMsg := NewRequest(123, []byte("Hello, World!"))

// Create a response message
responseMsg := NewResponse(123, []byte("Response received"))
```

### Using the Message Builder

```go
// Create a custom message using the builder pattern
msg := NewMessage().
    WithType(MessageTypeRequest).
    WithRequestID(123).
    WithPriority(2).
    WithPayload([]byte("Custom payload")).
    Build()
```

## Documentation

For detailed documentation covering all aspects of the library, including:
- Message Types and Structure
- Builder Pattern Usage
- Service Discovery
- State Management
- Edge Computing
- Observability
- Stream Processing
- Best Practices
- Security Considerations

Please refer to our [Technical Documentation](DOCUMENTATION.md).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

**Brayden Moon** ([@crazywolf132](https://github.com/crazywolf132))

## Organization

[FoxyCorps](https://github.com/foxycorps)

## Protocol Version

Current protocol version: 2.0

## Notes

- Maximum message size is 64MB
- Default TTL varies by message type
- All timestamps are in microseconds
- Messages are encoded in big-endian format 