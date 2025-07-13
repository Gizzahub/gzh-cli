package memory

import (
	"bytes"
	"encoding/json"
	"sync"
)

// CommonPools provides pre-configured memory pools for common data types
type CommonPools struct {
	ByteBuffers    *MemoryPool
	StringBuilders *MemoryPool
	JSONBuffers    *MemoryPool
	SliceInt       *MemoryPool
	SliceString    *MemoryPool
	Maps           *MemoryPool
}

// NewCommonPools creates commonly used memory pools
func NewCommonPools() *CommonPools {
	pools := &CommonPools{}

	// Byte buffer pool for I/O operations
	pools.ByteBuffers = &MemoryPool{
		name: "byte-buffers",
		pool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
		resetFunc: func(obj interface{}) {
			if buf, ok := obj.(*bytes.Buffer); ok {
				buf.Reset()
			}
		},
	}

	// String builder pool for string concatenation
	pools.StringBuilders = &MemoryPool{
		name: "string-builders",
		pool: sync.Pool{
			New: func() interface{} {
				return &StringBuilder{
					builder: &bytes.Buffer{},
				}
			},
		},
		resetFunc: func(obj interface{}) {
			if sb, ok := obj.(*StringBuilder); ok {
				sb.Reset()
			}
		},
	}

	// JSON buffer pool for JSON marshaling/unmarshaling
	pools.JSONBuffers = &MemoryPool{
		name: "json-buffers",
		pool: sync.Pool{
			New: func() interface{} {
				return &JSONBuffer{
					Buffer:  &bytes.Buffer{},
					Encoder: nil, // Will be created when needed
					Decoder: nil, // Will be created when needed
				}
			},
		},
		resetFunc: func(obj interface{}) {
			if jb, ok := obj.(*JSONBuffer); ok {
				jb.Reset()
			}
		},
	}

	// Integer slice pool
	pools.SliceInt = &MemoryPool{
		name: "slice-int",
		pool: sync.Pool{
			New: func() interface{} {
				slice := make([]int, 0, 64) // Initial capacity of 64
				return &slice
			},
		},
		resetFunc: func(obj interface{}) {
			if slicePtr, ok := obj.(*[]int); ok {
				// Reset slice length but keep capacity
				*slicePtr = (*slicePtr)[:0]
			}
		},
	}

	// String slice pool
	pools.SliceString = &MemoryPool{
		name: "slice-string",
		pool: sync.Pool{
			New: func() interface{} {
				slice := make([]string, 0, 32) // Initial capacity of 32
				return &slice
			},
		},
		resetFunc: func(obj interface{}) {
			if slicePtr, ok := obj.(*[]string); ok {
				// Reset slice length but keep capacity
				*slicePtr = (*slicePtr)[:0]
			}
		},
	}

	// Map pool for string->interface{} maps
	pools.Maps = &MemoryPool{
		name: "string-maps",
		pool: sync.Pool{
			New: func() interface{} {
				return make(map[string]interface{}, 16) // Initial capacity of 16
			},
		},
		resetFunc: func(obj interface{}) {
			if m, ok := obj.(map[string]interface{}); ok {
				// Clear the map
				for k := range m {
					delete(m, k)
				}
			}
		},
	}

	return pools
}

// StringBuilder wraps bytes.Buffer for string building operations
type StringBuilder struct {
	builder *bytes.Buffer
}

// WriteString writes a string to the builder
func (sb *StringBuilder) WriteString(s string) {
	sb.builder.WriteString(s)
}

// WriteByte writes a byte to the builder
func (sb *StringBuilder) WriteByte(b byte) error {
	return sb.builder.WriteByte(b)
}

// String returns the built string
func (sb *StringBuilder) String() string {
	return sb.builder.String()
}

// Reset resets the builder for reuse
func (sb *StringBuilder) Reset() {
	sb.builder.Reset()
}

// Len returns the length of the built string
func (sb *StringBuilder) Len() int {
	return sb.builder.Len()
}

// JSONBuffer wraps bytes.Buffer with JSON encoder/decoder for efficient JSON operations
type JSONBuffer struct {
	*bytes.Buffer
	Encoder *json.Encoder
	Decoder *json.Decoder
}

// GetEncoder returns a JSON encoder for this buffer
func (jb *JSONBuffer) GetEncoder() *json.Encoder {
	if jb.Encoder == nil {
		jb.Encoder = json.NewEncoder(jb.Buffer)
	}
	return jb.Encoder
}

// GetDecoder returns a JSON decoder for this buffer
func (jb *JSONBuffer) GetDecoder() *json.Decoder {
	if jb.Decoder == nil {
		jb.Decoder = json.NewDecoder(jb.Buffer)
	}
	return jb.Decoder
}

// Reset resets the buffer and recreates encoder/decoder
func (jb *JSONBuffer) Reset() {
	jb.Buffer.Reset()
	jb.Encoder = nil
	jb.Decoder = nil
}

// EncodeJSON encodes an object to JSON
func (jb *JSONBuffer) EncodeJSON(obj interface{}) error {
	jb.Reset()
	return jb.GetEncoder().Encode(obj)
}

// DecodeJSON decodes JSON from the buffer
func (jb *JSONBuffer) DecodeJSON(obj interface{}) error {
	return jb.GetDecoder().Decode(obj)
}

// GetByteBuffer gets a byte buffer from the pool
func (cp *CommonPools) GetByteBuffer() *bytes.Buffer {
	return cp.ByteBuffers.Get().(*bytes.Buffer)
}

// PutByteBuffer returns a byte buffer to the pool
func (cp *CommonPools) PutByteBuffer(buf *bytes.Buffer) {
	cp.ByteBuffers.Put(buf)
}

// GetStringBuilder gets a string builder from the pool
func (cp *CommonPools) GetStringBuilder() *StringBuilder {
	return cp.StringBuilders.Get().(*StringBuilder)
}

// PutStringBuilder returns a string builder to the pool
func (cp *CommonPools) PutStringBuilder(sb *StringBuilder) {
	cp.StringBuilders.Put(sb)
}

// GetJSONBuffer gets a JSON buffer from the pool
func (cp *CommonPools) GetJSONBuffer() *JSONBuffer {
	return cp.JSONBuffers.Get().(*JSONBuffer)
}

// PutJSONBuffer returns a JSON buffer to the pool
func (cp *CommonPools) PutJSONBuffer(jb *JSONBuffer) {
	cp.JSONBuffers.Put(jb)
}

// GetIntSlice gets an integer slice from the pool
func (cp *CommonPools) GetIntSlice() []int {
	slicePtr := cp.SliceInt.Get().(*[]int)
	return *slicePtr
}

// PutIntSlice returns an integer slice to the pool
func (cp *CommonPools) PutIntSlice(slice []int) {
	cp.SliceInt.Put(&slice)
}

// GetStringSlice gets a string slice from the pool
func (cp *CommonPools) GetStringSlice() []string {
	slicePtr := cp.SliceString.Get().(*[]string)
	return *slicePtr
}

// PutStringSlice returns a string slice to the pool
func (cp *CommonPools) PutStringSlice(slice []string) {
	cp.SliceString.Put(&slice)
}

// GetStringMap gets a string map from the pool
func (cp *CommonPools) GetStringMap() map[string]interface{} {
	return cp.Maps.Get().(map[string]interface{})
}

// PutStringMap returns a string map to the pool
func (cp *CommonPools) PutStringMap(m map[string]interface{}) {
	cp.Maps.Put(m)
}

// GetAllStats returns statistics for all common pools
func (cp *CommonPools) GetAllStats() map[string]PoolStats {
	return map[string]PoolStats{
		"byte-buffers":    cp.ByteBuffers.GetStats(),
		"string-builders": cp.StringBuilders.GetStats(),
		"json-buffers":    cp.JSONBuffers.GetStats(),
		"slice-int":       cp.SliceInt.GetStats(),
		"slice-string":    cp.SliceString.GetStats(),
		"string-maps":     cp.Maps.GetStats(),
	}
}

// GlobalPools provides globally accessible memory pools
var GlobalPools = NewCommonPools()

// Convenience functions for global pool access

// GetByteBuffer gets a byte buffer from the global pool
func GetByteBuffer() *bytes.Buffer {
	return GlobalPools.GetByteBuffer()
}

// PutByteBuffer returns a byte buffer to the global pool
func PutByteBuffer(buf *bytes.Buffer) {
	GlobalPools.PutByteBuffer(buf)
}

// GetStringBuilder gets a string builder from the global pool
func GetStringBuilder() *StringBuilder {
	return GlobalPools.GetStringBuilder()
}

// PutStringBuilder returns a string builder to the global pool
func PutStringBuilder(sb *StringBuilder) {
	GlobalPools.PutStringBuilder(sb)
}

// GetJSONBuffer gets a JSON buffer from the global pool
func GetJSONBuffer() *JSONBuffer {
	return GlobalPools.GetJSONBuffer()
}

// PutJSONBuffer returns a JSON buffer to the global pool
func PutJSONBuffer(jb *JSONBuffer) {
	GlobalPools.PutJSONBuffer(jb)
}

// WithByteBuffer executes a function with a pooled byte buffer
func WithByteBuffer(fn func(*bytes.Buffer) error) error {
	buf := GetByteBuffer()
	defer PutByteBuffer(buf)
	return fn(buf)
}

// WithStringBuilder executes a function with a pooled string builder
func WithStringBuilder(fn func(*StringBuilder) error) error {
	sb := GetStringBuilder()
	defer PutStringBuilder(sb)
	return fn(sb)
}

// WithJSONBuffer executes a function with a pooled JSON buffer
func WithJSONBuffer(fn func(*JSONBuffer) error) error {
	jb := GetJSONBuffer()
	defer PutJSONBuffer(jb)
	return fn(jb)
}

// PooledOperation represents a function that uses pooled resources
type PooledOperation[T any] func() (T, error)

// WithPooledResource executes an operation with automatic resource pooling
func WithPooledResource[T any](getResource func() T, putResource func(T), operation func(T) error) error {
	resource := getResource()
	defer putResource(resource)
	return operation(resource)
}
