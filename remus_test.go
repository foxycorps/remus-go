package remus

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"math"
	"testing"
	"time"
)

// Helper function to compare RemusMessages
func assertMessagesEqual(t *testing.T, expected, actual RemusMessage) {
	t.Helper()
	if expected.Type != actual.Type {
		t.Errorf("Message type mismatch: expected %v, got %v", expected.Type, actual.Type)
	}
	if expected.Flags != actual.Flags {
		t.Errorf("Flags mismatch: expected %v, got %v", expected.Flags, actual.Flags)
	}
	if expected.RequestID != actual.RequestID {
		t.Errorf("RequestID mismatch: expected %v, got %v", expected.RequestID, actual.RequestID)
	}
	if expected.Priority != actual.Priority {
		t.Errorf("Priority mismatch: expected %v, got %v", expected.Priority, actual.Priority)
	}
	if expected.TTL != actual.TTL {
		t.Errorf("TTL mismatch: expected %v, got %v", expected.TTL, actual.TTL)
	}
	if !bytes.Equal(expected.RoutingInfo, actual.RoutingInfo) {
		t.Errorf("RoutingInfo mismatch: expected %v, got %v", expected.RoutingInfo, actual.RoutingInfo)
	}
	if !bytes.Equal(expected.ContextMetadata, actual.ContextMetadata) {
		t.Errorf("ContextMetadata mismatch: expected %v, got %v", expected.ContextMetadata, actual.ContextMetadata)
	}
	if !bytes.Equal(expected.Payload, actual.Payload) {
		t.Errorf("Payload mismatch: expected %v, got %v", expected.Payload, actual.Payload)
	}
}

// Test basic message encoding/decoding
func TestMessageEncodeDecode(t *testing.T) {
	original := RemusMessage{
		Type:            MessageTypeRequest,
		Flags:           FlagHighPriority | FlagRequiresAuth,
		RequestID:       12345,
		Timestamp:       uint64(time.Now().UnixMicro()),
		Priority:        2,
		TTL:             1000,
		RoutingInfo:     []byte("test-routing"),
		ContextMetadata: []byte("test-metadata"),
		Payload:         []byte("test-payload"),
	}

	encoded, err := EncodeMessage(original)
	if err != nil {
		t.Fatalf("Failed to encode message: %v", err)
	}

	decoded, err := DecodeMessage(encoded)
	if err != nil {
		t.Fatalf("Failed to decode message: %v", err)
	}

	assertMessagesEqual(t, original, decoded)
}

// Test Handshake message creation and handling
func TestHandshakeMessage(t *testing.T) {
	version := VersionInfo{
		Major:          2,
		Minor:          0,
		FeatureFlags:   0x0001,
		Compression:    0x0002,
		ExtensionCount: 1,
		Extensions: []Extension{
			{
				Type:   1,
				Length: 4,
				Data:   []byte("test"),
			},
		},
	}

	msg, err := CreateHandshakeMessage(1, version)
	if err != nil {
		t.Fatalf("Failed to create handshake message: %v", err)
	}

	if msg.Type != MessageTypeHandshake {
		t.Errorf("Incorrect message type: expected %v, got %v", MessageTypeHandshake, msg.Type)
	}

	// Decode payload and verify version info
	var decodedVersion VersionInfo
	err = json.Unmarshal(msg.Payload, &decodedVersion)
	if err != nil {
		t.Fatalf("Failed to decode version info: %v", err)
	}

	if decodedVersion.Major != version.Major || decodedVersion.Minor != version.Minor {
		t.Errorf("Version mismatch: expected %v.%v, got %v.%v",
			version.Major, version.Minor, decodedVersion.Major, decodedVersion.Minor)
	}
}

// Test State Sync message creation and handling
func TestStateSyncMessage(t *testing.T) {
	state := StateSync{
		Version:      1,
		VectorClock:  123456,
		Dependencies: []uint32{1, 2, 3},
		DeltaOps: []StateOperation{
			{
				Type:    1,
				Key:     "test-key",
				Value:   []byte("test-value"),
				Version: 1,
			},
		},
		ValidationRules: []byte("test-rules"),
	}

	msg, err := CreateStateSyncMessage(1, state)
	if err != nil {
		t.Fatalf("Failed to create state sync message: %v", err)
	}

	if msg.Type != MessageTypeStateSync {
		t.Errorf("Incorrect message type: expected %v, got %v", MessageTypeStateSync, msg.Type)
	}

	if !HasFlag(msg.Flags, FlagStateDelta) {
		t.Error("StateDelta flag not set")
	}

	// Decode payload and verify state
	var decodedState StateSync
	err = json.Unmarshal(msg.Payload, &decodedState)
	if err != nil {
		t.Fatalf("Failed to decode state: %v", err)
	}

	if decodedState.Version != state.Version || decodedState.VectorClock != state.VectorClock {
		t.Error("State mismatch")
	}
}

// Test Edge Compute message creation and handling
func TestEdgeComputeMessage(t *testing.T) {
	function := EdgeFunction{
		FunctionID:    1,
		RuntimeType:   1,
		Requirements:  2,
		Code:          []byte("test-code"),
		Configuration: []byte("test-config"),
	}

	msg, err := CreateEdgeComputeMessage(1, function)
	if err != nil {
		t.Fatalf("Failed to create edge compute message: %v", err)
	}

	if msg.Type != MessageTypeEdgeCompute {
		t.Errorf("Incorrect message type: expected %v, got %v", MessageTypeEdgeCompute, msg.Type)
	}

	if !HasFlag(msg.Flags, FlagEdgeCompute) {
		t.Error("EdgeCompute flag not set")
	}

	// Decode payload and verify function
	var decodedFunction EdgeFunction
	err = json.Unmarshal(msg.Payload, &decodedFunction)
	if err != nil {
		t.Fatalf("Failed to decode function: %v", err)
	}

	if decodedFunction.FunctionID != function.FunctionID {
		t.Error("Function mismatch")
	}
}

// Test Service Discovery message creation and handling
func TestServiceDiscoveryMessage(t *testing.T) {
	service := ServiceInfo{
		ServiceID:     1,
		Version:       1,
		Capabilities:  0x0001,
		HealthStatus:  1,
		LoadMetrics:   50,
		EndpointCount: 1,
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

	msg, err := CreateServiceAnnounceMessage(1, service)
	if err != nil {
		t.Fatalf("Failed to create service announce message: %v", err)
	}

	if msg.Type != MessageTypeServiceAnnounce {
		t.Errorf("Incorrect message type: expected %v, got %v", MessageTypeServiceAnnounce, msg.Type)
	}

	// Decode payload and verify service info
	var decodedService ServiceInfo
	err = json.Unmarshal(msg.Payload, &decodedService)
	if err != nil {
		t.Fatalf("Failed to decode service info: %v", err)
	}

	if decodedService.ServiceID != service.ServiceID || decodedService.Version != service.Version {
		t.Error("Service info mismatch")
	}
}

// Test Observability message creation and handling
func TestObservabilityMessages(t *testing.T) {
	t.Run("Metrics", func(t *testing.T) {
		metrics := []Metric{
			{
				Timestamp: uint64(time.Now().UnixMicro()),
				MetricID:  1,
				Type:      1,
				Value:     42.0,
				Labels:    map[string]string{"test": "label"},
			},
		}

		msg, err := CreateMetricMessage(1, metrics)
		if err != nil {
			t.Fatalf("Failed to create metric message: %v", err)
		}

		if msg.Type != MessageTypeMetrics {
			t.Errorf("Incorrect message type: expected %v, got %v", MessageTypeMetrics, msg.Type)
		}

		var decodedMetrics []Metric
		err = json.Unmarshal(msg.Payload, &decodedMetrics)
		if err != nil {
			t.Fatalf("Failed to decode metrics: %v", err)
		}

		if len(decodedMetrics) != len(metrics) {
			t.Error("Metrics count mismatch")
		}
	})

	t.Run("Trace", func(t *testing.T) {
		span := TraceSpan{
			TraceID:      1,
			SpanID:       2,
			ParentSpanID: 0,
			StartTime:    uint64(time.Now().UnixMicro()),
			EndTime:      uint64(time.Now().UnixMicro()),
			Name:         "test-span",
			Attributes:   map[string]string{"test": "attr"},
		}

		msg, err := CreateTraceMessage(1, span)
		if err != nil {
			t.Fatalf("Failed to create trace message: %v", err)
		}

		if msg.Type != MessageTypeTraceSpan {
			t.Errorf("Incorrect message type: expected %v, got %v", MessageTypeTraceSpan, msg.Type)
		}

		var decodedSpan TraceSpan
		err = json.Unmarshal(msg.Payload, &decodedSpan)
		if err != nil {
			t.Fatalf("Failed to decode trace span: %v", err)
		}

		if decodedSpan.TraceID != span.TraceID || decodedSpan.SpanID != span.SpanID {
			t.Error("Trace span mismatch")
		}
	})

	t.Run("Log", func(t *testing.T) {
		entry := LogEntry{
			Timestamp: uint64(time.Now().UnixMicro()),
			Level:     1,
			SourceID:  1,
			TraceID:   1,
			Message:   "test-log",
			Metadata:  map[string]string{"test": "meta"},
		}

		msg, err := CreateLogMessage(1, entry)
		if err != nil {
			t.Fatalf("Failed to create log message: %v", err)
		}

		if msg.Type != MessageTypeLog {
			t.Errorf("Incorrect message type: expected %v, got %v", MessageTypeLog, msg.Type)
		}

		var decodedEntry LogEntry
		err = json.Unmarshal(msg.Payload, &decodedEntry)
		if err != nil {
			t.Fatalf("Failed to decode log entry: %v", err)
		}

		if decodedEntry.Message != entry.Message || decodedEntry.Level != entry.Level {
			t.Error("Log entry mismatch")
		}
	})
}

// Test utility functions
func TestUtilityFunctions(t *testing.T) {
	tests := []struct {
		msgType MessageType
		checks  []struct {
			fn     func(MessageType) bool
			expect bool
		}
	}{
		{
			MessageTypeHandshake,
			[]struct {
				fn     func(MessageType) bool
				expect bool
			}{
				{IsControlMessage, true},
				{IsDataOperation, false},
				{IsStateManagement, false},
				{IsEdgeOperation, false},
				{IsServiceDiscovery, false},
				{IsObservability, false},
			},
		},
		{
			MessageTypeMetrics,
			[]struct {
				fn     func(MessageType) bool
				expect bool
			}{
				{IsControlMessage, false},
				{IsDataOperation, false},
				{IsStateManagement, false},
				{IsEdgeOperation, false},
				{IsServiceDiscovery, false},
				{IsObservability, true},
			},
		},
	}

	for _, test := range tests {
		for _, check := range test.checks {
			if result := check.fn(test.msgType); result != check.expect {
				t.Errorf("For message type %v, expected %v, got %v", test.msgType, check.expect, result)
			}
		}
	}
}

// Test flag operations
func TestFlagOperations(t *testing.T) {
	flags := FlagCompressed | FlagHighPriority | FlagCacheable

	tests := []struct {
		flag     Flags
		expected bool
	}{
		{FlagCompressed, true},
		{FlagHighPriority, true},
		{FlagCacheable, true},
		{FlagRequiresAuth, false},
		{FlagIdempotent, false},
	}

	for _, test := range tests {
		if HasFlag(flags, test.flag) != test.expected {
			t.Errorf("Flag test failed for flag %v: expected %v", test.flag, test.expected)
		}
	}
}

// Test message validation
func TestMessageValidation(t *testing.T) {
	t.Run("Zero Length Message", func(t *testing.T) {
		_, err := DecodeMessage([]byte{})
		if err == nil {
			t.Error("Expected error for zero length message")
		}
	})

	t.Run("Invalid Length Field", func(t *testing.T) {
		msg := RemusMessage{
			Length:  1, // Invalid length
			Type:    MessageTypeRequest,
			Payload: []byte("test"),
		}
		_, err := EncodeMessage(msg)
		if err == nil {
			t.Error("Expected error for invalid length field")
		}
	})

	t.Run("Maximum Message Size", func(t *testing.T) {
		largePayload := make([]byte, 16*1024*1024) // 16MB
		rand.Read(largePayload)

		msg := RemusMessage{
			Type:    MessageTypeRequest,
			Payload: largePayload,
		}

		encoded, err := EncodeMessage(msg)
		if err != nil {
			t.Fatalf("Failed to encode large message: %v", err)
		}

		decoded, err := DecodeMessage(encoded)
		if err != nil {
			t.Fatalf("Failed to decode large message: %v", err)
		}

		if !bytes.Equal(msg.Payload, decoded.Payload) {
			t.Error("Large payload mismatch")
		}
	})

	t.Run("Invalid Message Type", func(t *testing.T) {
		msg := RemusMessage{
			Type: MessageType(0xFF), // Invalid type
		}
		if IsControlMessage(msg.Type) || IsDataOperation(msg.Type) ||
			IsStateManagement(msg.Type) || IsEdgeOperation(msg.Type) ||
			IsServiceDiscovery(msg.Type) || IsObservability(msg.Type) {
			t.Error("Invalid message type should not be categorized")
		}
	})
}

// Test edge cases in encoding/decoding
func TestEncodingEdgeCases(t *testing.T) {
	t.Run("Empty Fields", func(t *testing.T) {
		msg := RemusMessage{
			Type:            MessageTypeRequest,
			Flags:           0,
			RequestID:       0,
			Timestamp:       0,
			Priority:        0,
			TTL:             0,
			RoutingInfo:     nil,
			ContextMetadata: nil,
			Payload:         nil,
		}

		encoded, err := EncodeMessage(msg)
		if err != nil {
			t.Fatalf("Failed to encode message with empty fields: %v", err)
		}

		decoded, err := DecodeMessage(encoded)
		if err != nil {
			t.Fatalf("Failed to decode message with empty fields: %v", err)
		}

		assertMessagesEqual(t, msg, decoded)
	})

	t.Run("Maximum Values", func(t *testing.T) {
		msg := RemusMessage{
			Type:            MessageType(math.MaxUint8),
			Flags:           Flags(math.MaxUint16),
			RequestID:       math.MaxUint32,
			Timestamp:       math.MaxUint64,
			Priority:        math.MaxUint16,
			TTL:             math.MaxUint16,
			RoutingInfo:     bytes.Repeat([]byte{0xFF}, 1000),
			ContextMetadata: bytes.Repeat([]byte{0xFF}, 1000),
			Payload:         bytes.Repeat([]byte{0xFF}, 1000),
		}

		encoded, err := EncodeMessage(msg)
		if err != nil {
			t.Fatalf("Failed to encode message with maximum values: %v", err)
		}

		decoded, err := DecodeMessage(encoded)
		if err != nil {
			t.Fatalf("Failed to decode message with maximum values: %v", err)
		}

		assertMessagesEqual(t, msg, decoded)
	})
}

// Test protocol versioning
func TestProtocolVersioning(t *testing.T) {
	t.Run("Version Negotiation", func(t *testing.T) {
		versions := []struct {
			major, minor uint16
			valid        bool
		}{
			{2, 0, true},
			{1, 0, false},
			{2, 1, true},
			{3, 0, false},
		}

		for _, v := range versions {
			version := VersionInfo{
				Major:        v.major,
				Minor:        v.minor,
				FeatureFlags: 0x0001,
			}

			msg, err := CreateHandshakeMessage(1, version)
			if err != nil {
				t.Fatalf("Failed to create handshake message: %v", err)
			}

			var decoded VersionInfo
			err = json.Unmarshal(msg.Payload, &decoded)
			if err != nil {
				t.Fatalf("Failed to decode version info: %v", err)
			}

			if (decoded.Major == ProtocolVersionMajor) != v.valid {
				t.Errorf("Unexpected version validation result for %v.%v", v.major, v.minor)
			}
		}
	})
}

// Test state management
func TestStateManagement(t *testing.T) {
	t.Run("State Conflict Resolution", func(t *testing.T) {
		states := []StateSync{
			{
				Version:     1,
				VectorClock: 100,
				DeltaOps: []StateOperation{
					{Type: 1, Key: "key1", Value: []byte("value1"), Version: 1},
				},
			},
			{
				Version:     1,
				VectorClock: 200,
				DeltaOps: []StateOperation{
					{Type: 1, Key: "key1", Value: []byte("value2"), Version: 2},
				},
			},
		}

		// Create messages for both states
		for _, state := range states {
			msg, err := CreateStateSyncMessage(1, state)
			if err != nil {
				t.Fatalf("Failed to create state sync message: %v", err)
			}

			var decoded StateSync
			err = json.Unmarshal(msg.Payload, &decoded)
			if err != nil {
				t.Fatalf("Failed to decode state: %v", err)
			}

			if decoded.VectorClock != state.VectorClock {
				t.Error("Vector clock mismatch")
			}
		}
	})

	t.Run("State Delta Operations", func(t *testing.T) {
		ops := []StateOperation{
			{Type: 1, Key: "key1", Value: []byte("value1"), Version: 1},
			{Type: 2, Key: "key2", Value: []byte("value2"), Version: 1},
			{Type: 3, Key: "key3", Value: []byte("value3"), Version: 1},
		}

		state := StateSync{
			Version:     1,
			VectorClock: 100,
			DeltaOps:    ops,
		}

		msg, err := CreateStateSyncMessage(1, state)
		if err != nil {
			t.Fatalf("Failed to create state sync message: %v", err)
		}

		if !HasFlag(msg.Flags, FlagStateDelta) {
			t.Error("StateDelta flag not set for delta operations")
		}

		var decoded StateSync
		err = json.Unmarshal(msg.Payload, &decoded)
		if err != nil {
			t.Fatalf("Failed to decode state: %v", err)
		}

		if len(decoded.DeltaOps) != len(ops) {
			t.Error("Delta operations count mismatch")
		}
	})
}

// Test edge computing functionality
func TestEdgeComputing(t *testing.T) {
	t.Run("Edge Function Deployment", func(t *testing.T) {
		functions := []EdgeFunction{
			{
				FunctionID:    1,
				RuntimeType:   1, // WASM
				Requirements:  0x0001,
				Code:          []byte("test-wasm-code"),
				Configuration: []byte(`{"memory": 128, "timeout": 30}`),
			},
			{
				FunctionID:    2,
				RuntimeType:   2, // Native
				Requirements:  0x0002,
				Code:          []byte("test-native-code"),
				Configuration: []byte(`{"cpu": 2, "memory": 256}`),
			},
		}

		for _, fn := range functions {
			msg, err := CreateEdgeComputeMessage(1, fn)
			if err != nil {
				t.Fatalf("Failed to create edge compute message: %v", err)
			}

			if !HasFlag(msg.Flags, FlagEdgeCompute) {
				t.Error("EdgeCompute flag not set")
			}

			var decoded EdgeFunction
			err = json.Unmarshal(msg.Payload, &decoded)
			if err != nil {
				t.Fatalf("Failed to decode edge function: %v", err)
			}

			if decoded.FunctionID != fn.FunctionID || decoded.RuntimeType != fn.RuntimeType {
				t.Error("Edge function mismatch")
			}
		}
	})
}

// Test service discovery
func TestServiceDiscovery(t *testing.T) {
	t.Run("Service Registration", func(t *testing.T) {
		services := []ServiceInfo{
			{
				ServiceID:    1,
				Version:      1,
				Capabilities: 0x0001,
				HealthStatus: 1,
				LoadMetrics:  50,
				Endpoints: []ServiceEndpoint{
					{Type: 1, Priority: 1, Weight: 1, Port: 8080, Host: "host1"},
					{Type: 2, Priority: 2, Weight: 2, Port: 8081, Host: "host2"},
				},
			},
			{
				ServiceID:    2,
				Version:      1,
				Capabilities: 0x0002,
				HealthStatus: 1,
				LoadMetrics:  75,
				Endpoints: []ServiceEndpoint{
					{Type: 1, Priority: 1, Weight: 1, Port: 9090, Host: "host3"},
				},
			},
		}

		for _, svc := range services {
			msg, err := CreateServiceAnnounceMessage(1, svc)
			if err != nil {
				t.Fatalf("Failed to create service announce message: %v", err)
			}

			var decoded ServiceInfo
			err = json.Unmarshal(msg.Payload, &decoded)
			if err != nil {
				t.Fatalf("Failed to decode service info: %v", err)
			}

			if decoded.ServiceID != svc.ServiceID || len(decoded.Endpoints) != len(svc.Endpoints) {
				t.Error("Service info mismatch")
			}
		}
	})
}

// Test observability components
func TestObservabilityComponents(t *testing.T) {
	t.Run("Complex Metrics", func(t *testing.T) {
		metrics := []Metric{
			{
				Timestamp: uint64(time.Now().UnixMicro()),
				MetricID:  1,
				Type:      1,
				Value:     42.0,
				Labels: map[string]string{
					"service": "test",
					"env":     "prod",
					"region":  "us-west",
				},
			},
			{
				Timestamp: uint64(time.Now().UnixMicro()),
				MetricID:  2,
				Type:      2,
				Value:     99.9,
				Labels: map[string]string{
					"service": "test",
					"metric":  "cpu",
					"host":    "server1",
				},
			},
		}

		msg, err := CreateMetricMessage(1, metrics)
		if err != nil {
			t.Fatalf("Failed to create metric message: %v", err)
		}

		var decoded []Metric
		err = json.Unmarshal(msg.Payload, &decoded)
		if err != nil {
			t.Fatalf("Failed to decode metrics: %v", err)
		}

		if len(decoded) != len(metrics) {
			t.Error("Metrics count mismatch")
		}

		for i, m := range decoded {
			if len(m.Labels) != len(metrics[i].Labels) {
				t.Error("Metric labels mismatch")
			}
		}
	})

	t.Run("Trace Chain", func(t *testing.T) {
		spans := []TraceSpan{
			{
				TraceID:      1,
				SpanID:       1,
				ParentSpanID: 0,
				StartTime:    uint64(time.Now().UnixMicro()),
				EndTime:      uint64(time.Now().UnixMicro()),
				Name:         "root",
				Attributes: map[string]string{
					"service": "api",
				},
			},
			{
				TraceID:      1,
				SpanID:       2,
				ParentSpanID: 1,
				StartTime:    uint64(time.Now().UnixMicro()),
				EndTime:      uint64(time.Now().UnixMicro()),
				Name:         "db-query",
				Attributes: map[string]string{
					"service": "db",
					"query":   "select",
				},
			},
		}

		for _, span := range spans {
			msg, err := CreateTraceMessage(1, span)
			if err != nil {
				t.Fatalf("Failed to create trace message: %v", err)
			}

			var decoded TraceSpan
			err = json.Unmarshal(msg.Payload, &decoded)
			if err != nil {
				t.Fatalf("Failed to decode trace span: %v", err)
			}

			if decoded.TraceID != span.TraceID || decoded.SpanID != span.SpanID {
				t.Error("Trace span mismatch")
			}
		}
	})

	t.Run("Log Levels", func(t *testing.T) {
		levels := []uint32{1, 2, 3, 4, 5} // DEBUG, INFO, WARN, ERROR, FATAL
		for _, level := range levels {
			entry := LogEntry{
				Timestamp: uint64(time.Now().UnixMicro()),
				Level:     level,
				SourceID:  1,
				TraceID:   1,
				Message:   "test log message",
				Metadata: map[string]string{
					"level": "test",
				},
			}

			msg, err := CreateLogMessage(1, entry)
			if err != nil {
				t.Fatalf("Failed to create log message: %v", err)
			}

			var decoded LogEntry
			err = json.Unmarshal(msg.Payload, &decoded)
			if err != nil {
				t.Fatalf("Failed to decode log entry: %v", err)
			}

			if decoded.Level != entry.Level {
				t.Error("Log level mismatch")
			}
		}
	})
}

func TestMessageCorruption(t *testing.T) {
	t.Run("Corrupted Length", func(t *testing.T) {
		original := RemusMessage{
			Type:      MessageTypeRequest,
			RequestID: 1,
			Payload:   []byte("test"),
		}

		encoded, err := EncodeMessage(original)
		if err != nil {
			t.Fatalf("Failed to encode message: %v", err)
		}

		// Corrupt the length field
		binary.BigEndian.PutUint32(encoded[0:4], 99999)

		_, err = DecodeMessage(encoded)
		if err == nil {
			t.Error("Expected error for corrupted length")
		}
	})

	t.Run("Truncated Message", func(t *testing.T) {
		original := RemusMessage{
			Type:      MessageTypeRequest,
			RequestID: 1,
			Payload:   []byte("test"),
		}

		encoded, err := EncodeMessage(original)
		if err != nil {
			t.Fatalf("Failed to encode message: %v", err)
		}

		// Truncate the message
		truncated := encoded[:len(encoded)-10]

		_, err = DecodeMessage(truncated)
		if err == nil {
			t.Error("Expected error for truncated message")
		}
	})

	t.Run("Corrupted Type", func(t *testing.T) {
		original := RemusMessage{
			Type:      MessageTypeRequest,
			RequestID: 1,
			Payload:   []byte("test"),
		}

		encoded, err := EncodeMessage(original)
		if err != nil {
			t.Fatalf("Failed to encode message: %v", err)
		}

		// Corrupt the type field
		encoded[4] = 0xFF

		decoded, err := DecodeMessage(encoded)
		if err != nil {
			t.Fatalf("Failed to decode message: %v", err)
		}

		if decoded.Type == original.Type {
			t.Error("Expected different type after corruption")
		}
	})

	t.Run("Invalid JSON Payload", func(t *testing.T) {
		// Create a message with invalid JSON
		msg := RemusMessage{
			Type:      MessageTypeRequest,
			RequestID: 1,
			Payload:   []byte("{invalid-json}"),
		}

		encoded, err := EncodeMessage(msg)
		if err != nil {
			t.Fatalf("Failed to encode message: %v", err)
		}

		decoded, err := DecodeMessage(encoded)
		if err != nil {
			t.Fatalf("Failed to decode message: %v", err)
		}

		var jsonData map[string]interface{}
		if err := json.Unmarshal(decoded.Payload, &jsonData); err == nil {
			t.Error("Expected JSON unmarshal to fail")
		}
	})
}
