package remus

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Protocol version constants
const (
	ProtocolVersionMajor = 2
	ProtocolVersionMinor = 0
)

// MessageType defines the message types in the Remus protocol
type MessageType uint8

const (
	// Control messages (0x01-0x0F)
	MessageTypeHandshake        MessageType = 0x01
	MessageTypeHeartbeat        MessageType = 0x02
	MessageTypeGracefulShutdown MessageType = 0x03
	MessageTypeSchemaUpdate     MessageType = 0x04
	MessageTypeCapabilityUpdate MessageType = 0x05

	// Data operations (0x10-0x1F)
	MessageTypeRequest   MessageType = 0x10
	MessageTypeResponse  MessageType = 0x11
	MessageTypeError     MessageType = 0x12
	MessageTypeStream    MessageType = 0x13
	MessageTypeStreamEnd MessageType = 0x14

	// State management (0x20-0x2F)
	MessageTypeStateSync       MessageType = 0x20
	MessageTypeStateDelta      MessageType = 0x21
	MessageTypeStateValidation MessageType = 0x22
	MessageTypeStateConflict   MessageType = 0x23
	MessageTypeStateMerge      MessageType = 0x24

	// Edge operations (0x30-0x3F)
	MessageTypeEdgeCompute MessageType = 0x30
	MessageTypeEdgeState   MessageType = 0x31
	MessageTypeEdgeCache   MessageType = 0x32

	// Service Discovery (0x40-0x4F)
	MessageTypeServiceAnnounce MessageType = 0x40
	MessageTypeServiceQuery    MessageType = 0x41
	MessageTypeServiceUpdate   MessageType = 0x42
	MessageTypeHealthCheck     MessageType = 0x43

	// Observability (0x50-0x5F)
	MessageTypeMetrics   MessageType = 0x50
	MessageTypeTraceSpan MessageType = 0x51
	MessageTypeLog       MessageType = 0x52
	MessageTypeAudit     MessageType = 0x53
)

// Flags for the Remus message
type Flags uint16

const (
	FlagCompressed        Flags = 0x0001
	FlagRequiresAuth      Flags = 0x0002
	FlagHighPriority      Flags = 0x0004
	FlagCacheable         Flags = 0x0008
	FlagIdempotent        Flags = 0x0010
	FlagStreamEnd         Flags = 0x0020
	FlagBatchRequest      Flags = 0x0040
	FlagStateDelta        Flags = 0x0080
	FlagEdgeCompute       Flags = 0x0100
	FlagRequiresConsensus Flags = 0x0200
	FlagSensitiveData     Flags = 0x0400
	FlagCrossRegion       Flags = 0x0800
)

// RemusMessage represents the complete message structure according to spec
type RemusMessage struct {
	Length          uint32
	Type            MessageType
	Flags           Flags
	RequestID       uint32
	Timestamp       uint64 // microseconds
	Priority        uint16
	TTL             uint16
	RoutingInfo     []byte
	ContextMetadata []byte
	Payload         []byte
}

// VersionInfo represents the version negotiation structure
type VersionInfo struct {
	Major          uint16
	Minor          uint16
	FeatureFlags   uint32
	Compression    uint32
	ExtensionCount uint32
	Extensions     []Extension
}

// Extension represents a protocol extension
type Extension struct {
	Type   uint16
	Length uint16
	Data   []byte
}

// StateSync represents the state synchronization structure
type StateSync struct {
	Version         uint32
	VectorClock     uint64
	Dependencies    []uint32
	DeltaOps        []StateOperation
	ValidationRules []byte
}

// StateOperation represents a single state modification
type StateOperation struct {
	Type    uint8
	Key     string
	Value   []byte
	Version uint64
}

// EdgeFunction represents an edge computing function
type EdgeFunction struct {
	FunctionID    uint32
	RuntimeType   uint32
	Requirements  uint32
	Code          []byte
	Configuration []byte
}

// Service Discovery types
type ServiceInfo struct {
	ServiceID     uint32
	Version       uint32
	Capabilities  uint32
	HealthStatus  uint32
	LoadMetrics   uint32
	EndpointCount uint32
	Endpoints     []ServiceEndpoint
}

type ServiceEndpoint struct {
	Type     uint16
	Priority uint16
	Weight   uint16
	Port     uint16
	Host     string
}

// Observability types
type Metric struct {
	Timestamp uint64
	MetricID  uint32
	Type      uint8
	Value     float64
	Labels    map[string]string
}

type TraceSpan struct {
	TraceID      uint64
	SpanID       uint64
	ParentSpanID uint64
	StartTime    uint64
	EndTime      uint64
	Name         string
	Attributes   map[string]string
}

type LogEntry struct {
	Timestamp uint64
	Level     uint32
	SourceID  uint32
	TraceID   uint32
	Message   string
	Metadata  map[string]string
}

// Network Topology types
type MeshNode struct {
	MeshID       uint32
	NodeType     uint32
	Capabilities uint32
	LoadMetrics  uint32
	RoutingTable []RoutingEntry
}

type RoutingEntry struct {
	Destination uint32
	NextHop     uint32
	Metric      uint32
	Flags       uint32
}

// Error definitions
var (
	ErrInvalidMessage    = errors.New("invalid message format")
	ErrInvalidLength     = errors.New("invalid message length")
	ErrInvalidType       = errors.New("invalid message type")
	ErrInvalidPayload    = errors.New("invalid payload")
	ErrBufferTooSmall    = errors.New("buffer too small")
	ErrVersionMismatch   = errors.New("protocol version mismatch")
	ErrInvalidExtension  = errors.New("invalid extension format")
	ErrInvalidState      = errors.New("invalid state format")
	ErrInvalidEdgeConfig = errors.New("invalid edge configuration")
	ErrInvalidService    = errors.New("invalid service information")
	ErrInvalidMetric     = errors.New("invalid metric format")
	ErrInvalidTrace      = errors.New("invalid trace span")
	ErrInvalidLog        = errors.New("invalid log entry")
	ErrInvalidMeshNode   = errors.New("invalid mesh node configuration")
)

const (
	HeaderSize       = 23               // 4 + 1 + 2 + 4 + 8 + 2 + 2 = 23 bytes
	LengthFieldSize  = 4                // Size of each length field
	MinMessageLength = 35               // HeaderSize + (3 * LengthFieldSize)
	MaxMessageLength = 64 * 1024 * 1024 // 64MB max message size
)

// EncodeMessage encodes a RemusMessage into a byte slice
func EncodeMessage(message RemusMessage) ([]byte, error) {
	var buffer bytes.Buffer

	// Calculate variable lengths
	routingLen := uint32(len(message.RoutingInfo))
	metadataLen := uint32(len(message.ContextMetadata))
	payloadLen := uint32(len(message.Payload))

	// Calculate total length
	// HeaderSize: all fixed fields
	// 3 * LengthFieldSize: size fields for RoutingInfo, ContextMetadata, and Payload
	// routingLen + metadataLen + payloadLen: actual variable data
	totalLength := HeaderSize + (3 * LengthFieldSize) + routingLen + metadataLen + payloadLen

	// Validate message length
	if totalLength > MaxMessageLength {
		return nil, fmt.Errorf("total message length %d exceeds maximum %d", totalLength, MaxMessageLength)
	}

	// Check if the provided Length field is incorrect
	if message.Length != 0 && message.Length != totalLength {
		return nil, ErrInvalidLength
	}

	// Update the message length field
	message.Length = totalLength

	// Write fixed-length header fields
	if err := binary.Write(&buffer, binary.BigEndian, message.Length); err != nil {
		return nil, err
	}
	if err := binary.Write(&buffer, binary.BigEndian, message.Type); err != nil {
		return nil, err
	}
	if err := binary.Write(&buffer, binary.BigEndian, message.Flags); err != nil {
		return nil, err
	}
	if err := binary.Write(&buffer, binary.BigEndian, message.RequestID); err != nil {
		return nil, err
	}
	if err := binary.Write(&buffer, binary.BigEndian, message.Timestamp); err != nil {
		return nil, err
	}
	if err := binary.Write(&buffer, binary.BigEndian, message.Priority); err != nil {
		return nil, err
	}
	if err := binary.Write(&buffer, binary.BigEndian, message.TTL); err != nil {
		return nil, err
	}

	// Write variable-length fields with their lengths
	if err := binary.Write(&buffer, binary.BigEndian, routingLen); err != nil {
		return nil, err
	}
	if routingLen > 0 {
		if _, err := buffer.Write(message.RoutingInfo); err != nil {
			return nil, err
		}
	}

	if err := binary.Write(&buffer, binary.BigEndian, metadataLen); err != nil {
		return nil, err
	}
	if metadataLen > 0 {
		if _, err := buffer.Write(message.ContextMetadata); err != nil {
			return nil, err
		}
	}

	if err := binary.Write(&buffer, binary.BigEndian, payloadLen); err != nil {
		return nil, err
	}
	if payloadLen > 0 {
		if _, err := buffer.Write(message.Payload); err != nil {
			return nil, err
		}
	}

	if uint32(buffer.Len()) != totalLength {
		return nil, fmt.Errorf("encoded length %d does not match calculated length %d", buffer.Len(), totalLength)
	}

	return buffer.Bytes(), nil
}

// DecodeMessage decodes a byte slice into a RemusMessage
func DecodeMessage(data []byte) (RemusMessage, error) {
	var message RemusMessage

	// Check minimum length requirement
	if len(data) < MinMessageLength {
		return message, fmt.Errorf("message length %d is less than minimum required size %d", len(data), MinMessageLength)
	}

	buffer := bytes.NewReader(data)

	// Read declared length first
	if err := binary.Read(buffer, binary.BigEndian, &message.Length); err != nil {
		return message, err
	}

	// Validate declared length
	if message.Length < MinMessageLength {
		return message, fmt.Errorf("declared length %d is less than minimum size %d", message.Length, MinMessageLength)
	}
	if message.Length > MaxMessageLength {
		return message, fmt.Errorf("declared length %d exceeds maximum size %d", message.Length, MaxMessageLength)
	}

	// Check if actual data length matches declared length
	if uint32(len(data)) != message.Length {
		return message, fmt.Errorf("actual data length %d does not match declared length %d", len(data), message.Length)
	}

	// Read remaining fixed-length fields
	if err := binary.Read(buffer, binary.BigEndian, &message.Type); err != nil {
		return message, err
	}
	if err := binary.Read(buffer, binary.BigEndian, &message.Flags); err != nil {
		return message, err
	}
	if err := binary.Read(buffer, binary.BigEndian, &message.RequestID); err != nil {
		return message, err
	}
	if err := binary.Read(buffer, binary.BigEndian, &message.Timestamp); err != nil {
		return message, err
	}
	if err := binary.Read(buffer, binary.BigEndian, &message.Priority); err != nil {
		return message, err
	}
	if err := binary.Read(buffer, binary.BigEndian, &message.TTL); err != nil {
		return message, err
	}

	// Read variable-length fields
	var routingLen, metadataLen, payloadLen uint32

	if err := binary.Read(buffer, binary.BigEndian, &routingLen); err != nil {
		return message, err
	}
	if routingLen > 0 {
		message.RoutingInfo = make([]byte, routingLen)
		if _, err := buffer.Read(message.RoutingInfo); err != nil {
			return message, err
		}
	}

	if err := binary.Read(buffer, binary.BigEndian, &metadataLen); err != nil {
		return message, err
	}
	if metadataLen > 0 {
		message.ContextMetadata = make([]byte, metadataLen)
		if _, err := buffer.Read(message.ContextMetadata); err != nil {
			return message, err
		}
	}

	if err := binary.Read(buffer, binary.BigEndian, &payloadLen); err != nil {
		return message, err
	}
	if payloadLen > 0 {
		message.Payload = make([]byte, payloadLen)
		if _, err := buffer.Read(message.Payload); err != nil {
			return message, err
		}
	}

	// Verify total length
	expectedLength := HeaderSize + (3 * LengthFieldSize) + routingLen + metadataLen + payloadLen
	if message.Length != expectedLength {
		return message, fmt.Errorf("message length %d does not match expected length %d", message.Length, expectedLength)
	}

	return message, nil
}

// Helper functions for creating specific message types

func CreateHandshakeMessage(requestID uint32, version VersionInfo) (RemusMessage, error) {
	payload, err := json.Marshal(version)
	if err != nil {
		return RemusMessage{}, err
	}

	return RemusMessage{
		Type:      MessageTypeHandshake,
		Flags:     0,
		RequestID: requestID,
		Timestamp: uint64(time.Now().UnixMicro()),
		Priority:  0,
		TTL:       1000, // Default 1 second TTL for handshake
		Payload:   payload,
	}, nil
}

func CreateStateSyncMessage(requestID uint32, state StateSync) (RemusMessage, error) {
	payload, err := json.Marshal(state)
	if err != nil {
		return RemusMessage{}, err
	}

	return RemusMessage{
		Type:      MessageTypeStateSync,
		Flags:     FlagStateDelta,
		RequestID: requestID,
		Timestamp: uint64(time.Now().UnixMicro()),
		Priority:  1,    // Medium priority
		TTL:       5000, // 5 second TTL for state sync
		Payload:   payload,
	}, nil
}

func CreateEdgeComputeMessage(requestID uint32, function EdgeFunction) (RemusMessage, error) {
	payload, err := json.Marshal(function)
	if err != nil {
		return RemusMessage{}, err
	}

	return RemusMessage{
		Type:      MessageTypeEdgeCompute,
		Flags:     FlagEdgeCompute,
		RequestID: requestID,
		Timestamp: uint64(time.Now().UnixMicro()),
		Priority:  2,    // Higher priority
		TTL:       3000, // 3 second TTL for edge compute
		Payload:   payload,
	}, nil
}

func CreateStreamMessage(requestID uint32, data []byte, isEnd bool) RemusMessage {
	flags := Flags(0)
	if isEnd {
		flags |= FlagStreamEnd
	}

	return RemusMessage{
		Type:      MessageTypeStream,
		Flags:     flags,
		RequestID: requestID,
		Timestamp: uint64(time.Now().UnixMicro()),
		Priority:  1,    // Medium priority
		TTL:       2000, // 2 second TTL for stream messages
		Payload:   data,
	}
}

func CreateErrorMessage(requestID uint32, errorCode uint32, errorMessage string) RemusMessage {
	payload, _ := json.Marshal(map[string]interface{}{
		"code":    errorCode,
		"message": errorMessage,
	})

	return RemusMessage{
		Type:      MessageTypeError,
		Flags:     0,
		RequestID: requestID,
		Timestamp: uint64(time.Now().UnixMicro()),
		Priority:  2,    // Higher priority for errors
		TTL:       1000, // 1 second TTL for errors
		Payload:   payload,
	}
}

// Utility functions

func IsControlMessage(msgType MessageType) bool {
	return msgType >= 0x01 && msgType <= 0x0F
}

func IsDataOperation(msgType MessageType) bool {
	return msgType >= 0x10 && msgType <= 0x1F
}

func IsStateManagement(msgType MessageType) bool {
	return msgType >= 0x20 && msgType <= 0x2F
}

func IsEdgeOperation(msgType MessageType) bool {
	return msgType >= 0x30 && msgType <= 0x3F
}

func HasFlag(flags Flags, flag Flags) bool {
	return (flags & flag) == flag
}

func CreateServiceAnnounceMessage(requestID uint32, service ServiceInfo) (RemusMessage, error) {
	payload, err := json.Marshal(service)
	if err != nil {
		return RemusMessage{}, err
	}

	return RemusMessage{
		Type:      MessageTypeServiceAnnounce,
		Flags:     0,
		RequestID: requestID,
		Timestamp: uint64(time.Now().UnixMicro()),
		Priority:  1,
		TTL:       30000, // 30 second TTL for service announcements
		Payload:   payload,
	}, nil
}

func CreateMetricMessage(requestID uint32, metrics []Metric) (RemusMessage, error) {
	payload, err := json.Marshal(metrics)
	if err != nil {
		return RemusMessage{}, err
	}

	return RemusMessage{
		Type:      MessageTypeMetrics,
		Flags:     0,
		RequestID: requestID,
		Timestamp: uint64(time.Now().UnixMicro()),
		Priority:  1,
		TTL:       5000, // 5 second TTL for metrics
		Payload:   payload,
	}, nil
}

func CreateTraceMessage(requestID uint32, span TraceSpan) (RemusMessage, error) {
	payload, err := json.Marshal(span)
	if err != nil {
		return RemusMessage{}, err
	}

	return RemusMessage{
		Type:      MessageTypeTraceSpan,
		Flags:     0,
		RequestID: requestID,
		Timestamp: uint64(time.Now().UnixMicro()),
		Priority:  1,
		TTL:       10000, // 10 second TTL for trace spans
		Payload:   payload,
	}, nil
}

func CreateLogMessage(requestID uint32, entry LogEntry) (RemusMessage, error) {
	payload, err := json.Marshal(entry)
	if err != nil {
		return RemusMessage{}, err
	}

	return RemusMessage{
		Type:      MessageTypeLog,
		Flags:     0,
		RequestID: requestID,
		Timestamp: uint64(time.Now().UnixMicro()),
		Priority:  1,
		TTL:       5000, // 5 second TTL for logs
		Payload:   payload,
	}, nil
}

// Additional utility functions

func IsServiceDiscovery(msgType MessageType) bool {
	return msgType >= 0x40 && msgType <= 0x4F
}

func IsObservability(msgType MessageType) bool {
	return msgType >= 0x50 && msgType <= 0x5F
}
