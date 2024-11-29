package remus

import (
	"encoding/json"
	"time"
)

// MessageBuilder provides a fluent interface for building messages
type MessageBuilder struct {
	msg RemusMessage
}

// NewMessage creates a new MessageBuilder with default values
func NewMessage() *MessageBuilder {
	return &MessageBuilder{
		msg: RemusMessage{
			Timestamp: uint64(time.Now().UnixMicro()),
			Priority:  1,    // Default priority
			TTL:       5000, // Default 5 second TTL
		},
	}
}

// WithType sets the message type
func (b *MessageBuilder) WithType(t MessageType) *MessageBuilder {
	b.msg.Type = t
	return b
}

// WithFlags sets the message flags
func (b *MessageBuilder) WithFlags(f Flags) *MessageBuilder {
	b.msg.Flags = f
	return b
}

// WithRequestID sets the request ID
func (b *MessageBuilder) WithRequestID(id uint32) *MessageBuilder {
	b.msg.RequestID = id
	return b
}

// WithPriority sets the priority
func (b *MessageBuilder) WithPriority(p uint16) *MessageBuilder {
	b.msg.Priority = p
	return b
}

// WithTTL sets the TTL
func (b *MessageBuilder) WithTTL(ttl uint16) *MessageBuilder {
	b.msg.TTL = ttl
	return b
}

// WithRoutingInfo sets the routing info
func (b *MessageBuilder) WithRoutingInfo(info []byte) *MessageBuilder {
	b.msg.RoutingInfo = info
	return b
}

// WithContextMetadata sets the context metadata
func (b *MessageBuilder) WithContextMetadata(meta []byte) *MessageBuilder {
	b.msg.ContextMetadata = meta
	return b
}

// WithPayload sets the payload
func (b *MessageBuilder) WithPayload(payload []byte) *MessageBuilder {
	b.msg.Payload = payload
	return b
}

// Build returns the constructed message
func (b *MessageBuilder) Build() RemusMessage {
	return b.msg
}

// Quick helper functions for common message types

// NewHandshake creates a handshake message with sensible defaults
func NewHandshake(requestID uint32) (RemusMessage, error) {
	version := VersionInfo{
		Major:        ProtocolVersionMajor,
		Minor:        ProtocolVersionMinor,
		FeatureFlags: 0x0001,
	}

	payload, err := json.Marshal(version)
	if err != nil {
		return RemusMessage{}, err
	}

	return NewMessage().
		WithType(MessageTypeHandshake).
		WithRequestID(requestID).
		WithPriority(2).
		WithTTL(1000).
		WithPayload(payload).
		Build(), nil
}

// NewRequest creates a request message with sensible defaults
func NewRequest(requestID uint32, payload []byte) RemusMessage {
	return NewMessage().
		WithType(MessageTypeRequest).
		WithRequestID(requestID).
		WithPayload(payload).
		Build()
}

// NewResponse creates a response message with sensible defaults
func NewResponse(requestID uint32, payload []byte) RemusMessage {
	return NewMessage().
		WithType(MessageTypeResponse).
		WithRequestID(requestID).
		WithPayload(payload).
		Build()
}

// NewError creates an error message with sensible defaults
func NewError(requestID uint32, code uint32, message string) (RemusMessage, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"code":    code,
		"message": message,
	})
	if err != nil {
		return RemusMessage{}, err
	}

	return NewMessage().
		WithType(MessageTypeError).
		WithRequestID(requestID).
		WithPriority(2).
		WithPayload(payload).
		Build(), nil
}

// NewStateSync creates a state sync message with sensible defaults
func NewStateSync(requestID uint32, state interface{}) (RemusMessage, error) {
	payload, err := json.Marshal(state)
	if err != nil {
		return RemusMessage{}, err
	}

	return NewMessage().
		WithType(MessageTypeStateSync).
		WithRequestID(requestID).
		WithFlags(FlagStateDelta).
		WithPayload(payload).
		Build(), nil
}

// NewMetric creates a metric message with sensible defaults
func NewMetric(requestID uint32, metrics []Metric) (RemusMessage, error) {
	payload, err := json.Marshal(metrics)
	if err != nil {
		return RemusMessage{}, err
	}

	return NewMessage().
		WithType(MessageTypeMetrics).
		WithRequestID(requestID).
		WithPayload(payload).
		Build(), nil
}

// NewStreamData creates a stream data message with sensible defaults
func NewStreamData(requestID uint32, data []byte, isEnd bool) RemusMessage {
	flags := Flags(0)
	if isEnd {
		flags |= FlagStreamEnd
	}

	return NewMessage().
		WithType(MessageTypeStream).
		WithRequestID(requestID).
		WithFlags(flags).
		WithPayload(data).
		Build()
}

// NewServiceAnnounce creates a service announcement message with sensible defaults
func NewServiceAnnounce(requestID uint32, service ServiceInfo) (RemusMessage, error) {
	payload, err := json.Marshal(service)
	if err != nil {
		return RemusMessage{}, err
	}

	return NewMessage().
		WithType(MessageTypeServiceAnnounce).
		WithRequestID(requestID).
		WithTTL(30000). // 30 second TTL for service announcements
		WithPayload(payload).
		Build(), nil
}
