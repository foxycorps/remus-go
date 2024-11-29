# Remus Technical Documentation

## Table of Contents
1. [Message Types](#message-types)
2. [Message Structure](#message-structure)
3. [Builder Pattern](#builder-pattern)
4. [Service Discovery](#service-discovery)
5. [State Management](#state-management)
6. [Edge Computing](#edge-computing)
7. [Observability](#observability)
8. [Stream Processing](#stream-processing)

## Message Types

### Control Messages (0x01-0x0F)

#### Handshake
```go
handshake := NewMessage().
    WithType(MessageTypeHandshake).
    WithPayload([]byte("client-id-123")).
    Build()
```

#### Heartbeat
```go
heartbeat := NewMessage().
    WithType(MessageTypeHeartbeat).
    WithTTL(1000). // 1 second
    Build()
```

### Data Operations (0x10-0x1F)

#### Request/Response Pattern
```go
// Creating a request
request := NewMessage().
    WithType(MessageTypeRequest).
    WithRequestID(123).
    WithPayload([]byte("request-data")).
    Build()

// Creating a response
response := NewMessage().
    WithType(MessageTypeResponse).
    WithRequestID(123). // Same as request ID
    WithPayload([]byte("response-data")).
    Build()
```

#### Error Handling
```go
errorMsg := NewMessage().
    WithType(MessageTypeError).
    WithRequestID(123).
    WithPayload([]byte("error-details")).
    Build()
```

## Message Structure

### Basic Message Components
```go
type RemusMessage struct {
    Length    uint32
    Type      MessageType
    Flags     Flags
    RequestID uint32
    Timestamp uint64
    Priority  uint16
    TTL       uint16
    Payload   []byte
}
```

### Using Flags
```go
msg := NewMessage().
    WithFlags(FlagEncrypted | FlagCompressed).
    WithType(MessageTypeRequest).
    Build()
```

## Builder Pattern

### Basic Usage
```go
msg := NewMessage().
    WithType(MessageTypeRequest).
    WithRequestID(123).
    WithPriority(2).
    WithTTL(5000).
    WithPayload([]byte("payload")).
    Build()
```

### Custom Builders
```go
// Creating a request builder
func NewRequestBuilder() *MessageBuilder {
    return NewMessage().
        WithType(MessageTypeRequest).
        WithPriority(1).
        WithTTL(5000)
}

// Usage
request := NewRequestBuilder().
    WithRequestID(123).
    WithPayload([]byte("data")).
    Build()
```

## Service Discovery

### Announcing a Service
```go
service := ServiceInfo{
    ServiceID:    1,
    Version:      1,
    Capabilities: 0x0001,
    HealthStatus: 1,
    LoadMetrics:  0,
    Endpoints: []ServiceEndpoint{
        {
            Type:     1,
            Priority: 1,
            Weight:   1,
            Port:     8080,
            Host:     "localhost",
        },
    },
}

msg, err := NewServiceAnnounce(1, service)
```

### Service Query
```go
query := ServiceInfo{
    ServiceID: 1,
    Version:   1,
}

msg, err := NewServiceQuery(1, query)
```

## State Management

### State Synchronization
```go
state := map[string]interface{}{
    "key":     "value",
    "version": 1,
    "data":    []byte("binary-data"),
}

msg, err := NewStateSync(1, state)
```

### Delta Updates
```go
delta := map[string]interface{}{
    "updated_key": "new_value",
    "version":     2,
}

msg, err := NewStateDelta(1, delta)
```

## Edge Computing

### Deploying Edge Functions
```go
function := EdgeFunction{
    FunctionID:    1,
    RuntimeType:   1, // e.g., 1 for JavaScript
    Requirements:  0x0001,
    Code:         []byte("function code here"),
    Configuration: []byte("config data"),
}

msg, err := NewEdgeCompute(1, function)
```

## Observability

### Metrics
```go
metric := Metric{
    Timestamp: uint64(time.Now().UnixMicro()),
    MetricID:  1,
    Type:      1, // e.g., 1 for counter
    Value:     42.0,
    Labels:    map[string]string{"service": "api"},
}

msg, err := CreateMetricMessage(1, metric)
```

### Tracing
```go
span := TraceSpan{
    TraceID:      1,
    SpanID:       2,
    ParentSpanID: 0,
    StartTime:    uint64(time.Now().UnixMicro()),
    EndTime:      uint64(time.Now().UnixMicro()),
    Name:         "http-request",
    Attributes:   map[string]string{"method": "GET"},
}

msg, err := CreateTraceMessage(1, span)
```

### Logging
```go
entry := LogEntry{
    Timestamp: uint64(time.Now().UnixMicro()),
    Level:     1, // e.g., 1 for INFO
    SourceID:  1,
    TraceID:   1,
    Message:   "API request processed",
    Metadata:  map[string]string{"status": "200"},
}

msg, err := CreateLogMessage(1, entry)
```

## Stream Processing

### Streaming Data
```go
// Starting a stream
startMsg := NewStreamData(123, []byte("chunk1"), false)

// Middle chunks
chunkMsg := NewStreamData(123, []byte("chunk2"), false)

// Ending a stream
endMsg := NewStreamData(123, []byte("final-chunk"), true)
```

### Stream Processing Patterns
```go
// Stream processor example
func ProcessStream(streamID uint32, handler func([]byte) error) {
    for {
        msg := receiveMessage()
        if msg.Type != MessageTypeStream {
            continue
        }
        
        if err := handler(msg.Payload); err != nil {
            // Handle error
            break
        }
        
        if msg.Flags&FlagStreamEnd != 0 {
            break
        }
    }
}
```

## Best Practices

1. **Message Sizing**
   - Keep messages under 64MB
   - Use streaming for large data transfers
   - Compress payloads when appropriate

2. **TTL Management**
   - Set appropriate TTLs based on message type
   - Use shorter TTLs for time-sensitive messages
   - Consider network latency when setting TTLs

3. **Error Handling**
   - Always check error returns
   - Use appropriate error messages
   - Include context in error responses

4. **Performance Optimization**
   - Use appropriate message priorities
   - Batch messages when possible
   - Implement proper cleanup of resources

## Protocol Versioning

The current protocol version is 2.0. Version compatibility can be checked using the handshake message type.

```go
handshake := NewMessage().
    WithType(MessageTypeHandshake).
    WithPayload([]byte("v2.0")).
    Build()
```

## Security Considerations

1. **Encryption**
   - Use FlagEncrypted for sensitive data
   - Implement proper key management
   - Validate message integrity

2. **Authentication**
   - Implement proper service authentication
   - Validate service announcements
   - Use secure endpoints

3. **Authorization**
   - Implement proper access control
   - Validate message permissions
   - Log security events 