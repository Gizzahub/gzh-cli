package memory

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommonPools(t *testing.T) {
	pools := NewCommonPools()
	require.NotNil(t, pools)

	t.Run("ByteBuffers", func(t *testing.T) {
		buf := pools.GetByteBuffer()
		require.NotNil(t, buf)

		// Test writing to buffer
		buf.WriteString("test data")
		assert.Equal(t, "test data", buf.String())

		// Return to pool (should reset)
		pools.PutByteBuffer(buf)

		// Get another buffer - should be clean
		buf2 := pools.GetByteBuffer()
		assert.Equal(t, 0, buf2.Len())
	})

	t.Run("StringBuilders", func(t *testing.T) {
		sb := pools.GetStringBuilder()
		require.NotNil(t, sb)

		// Test building string
		sb.WriteString("hello")
		sb.WriteString(" ")
		sb.WriteString("world")
		assert.Equal(t, "hello world", sb.String())
		assert.Equal(t, 11, sb.Len())

		// Return to pool (should reset)
		pools.PutStringBuilder(sb)

		// Get another builder - should be clean
		sb2 := pools.GetStringBuilder()
		assert.Equal(t, 0, sb2.Len())
	})

	t.Run("JSONBuffers", func(t *testing.T) {
		jb := pools.GetJSONBuffer()
		require.NotNil(t, jb)

		// Test JSON encoding
		testData := map[string]interface{}{
			"name": "test",
			"age":  25,
		}

		err := jb.EncodeJSON(testData)
		assert.NoError(t, err)
		assert.True(t, jb.Len() > 0)

		// Test JSON decoding
		jb.Reset()
		jb.WriteString(`{"name":"decoded","value":42}`)

		var decoded map[string]interface{}
		err = jb.DecodeJSON(&decoded)
		assert.NoError(t, err)
		assert.Equal(t, "decoded", decoded["name"])
		assert.Equal(t, float64(42), decoded["value"])

		// Return to pool
		pools.PutJSONBuffer(jb)
	})

	t.Run("IntSlices", func(t *testing.T) {
		slice := pools.GetIntSlice()
		require.NotNil(t, slice)
		assert.Equal(t, 0, len(slice))
		assert.True(t, cap(slice) >= 64) // Should have initial capacity

		// Add some data
		slice = append(slice, 1, 2, 3, 4, 5)
		assert.Equal(t, 5, len(slice))

		// Return to pool (should reset length)
		pools.PutIntSlice(slice)

		// Get another slice - should be clean but keep capacity
		slice2 := pools.GetIntSlice()
		assert.Equal(t, 0, len(slice2))
	})

	t.Run("StringSlices", func(t *testing.T) {
		slice := pools.GetStringSlice()
		require.NotNil(t, slice)
		assert.Equal(t, 0, len(slice))
		assert.True(t, cap(slice) >= 32)

		// Add some data
		slice = append(slice, "a", "b", "c")
		assert.Equal(t, 3, len(slice))

		// Return to pool
		pools.PutStringSlice(slice)

		// Get another slice - should be clean
		slice2 := pools.GetStringSlice()
		assert.Equal(t, 0, len(slice2))
	})

	t.Run("StringMaps", func(t *testing.T) {
		m := pools.GetStringMap()
		require.NotNil(t, m)
		assert.Equal(t, 0, len(m))

		// Add some data
		m["key1"] = "value1"
		m["key2"] = 42
		assert.Equal(t, 2, len(m))

		// Return to pool (should clear)
		pools.PutStringMap(m)

		// Get another map - should be clean
		m2 := pools.GetStringMap()
		assert.Equal(t, 0, len(m2))
	})

	t.Run("GetAllStats", func(t *testing.T) {
		// Create a fresh pool instance to avoid interference from other tests
		freshPools := NewCommonPools()

		// Use all pools once
		buf := freshPools.GetByteBuffer()
		freshPools.PutByteBuffer(buf)

		sb := freshPools.GetStringBuilder()
		freshPools.PutStringBuilder(sb)

		stats := freshPools.GetAllStats()
		assert.NotNil(t, stats)
		assert.Contains(t, stats, "byte-buffers")
		assert.Contains(t, stats, "string-builders")
		assert.Contains(t, stats, "json-buffers")
		assert.Contains(t, stats, "slice-int")
		assert.Contains(t, stats, "slice-string")
		assert.Contains(t, stats, "string-maps")

		// Check that byte-buffers has activity
		bufStats := stats["byte-buffers"]
		assert.Equal(t, int64(1), bufStats.Gets)
		assert.Equal(t, int64(1), bufStats.Puts)
	})
}

func TestStringBuilder(t *testing.T) {
	sb := &StringBuilder{
		builder: &bytes.Buffer{},
	}

	t.Run("BasicOperations", func(t *testing.T) {
		sb.WriteString("hello")
		_ = sb.WriteByte(' ')
		sb.WriteString("world")

		assert.Equal(t, "hello world", sb.String())
		assert.Equal(t, 11, sb.Len())

		sb.Reset()
		assert.Equal(t, 0, sb.Len())
		assert.Equal(t, "", sb.String())
	})
}

func TestJSONBuffer(t *testing.T) {
	jb := &JSONBuffer{
		Buffer: &bytes.Buffer{},
	}

	t.Run("EncoderDecoder", func(t *testing.T) {
		// Test encoder creation
		encoder := jb.GetEncoder()
		assert.NotNil(t, encoder)
		assert.Same(t, encoder, jb.GetEncoder()) // Should return same instance

		// Test decoder creation
		jb.WriteString(`{"test": "data"}`)
		decoder := jb.GetDecoder()
		assert.NotNil(t, decoder)
		assert.Same(t, decoder, jb.GetDecoder()) // Should return same instance

		// Test reset
		jb.Reset()
		assert.Equal(t, 0, jb.Len())
		assert.Nil(t, jb.Encoder)
		assert.Nil(t, jb.Decoder)
	})

	t.Run("EncodeDecodeJSON", func(t *testing.T) {
		testData := map[string]interface{}{
			"name":   "test",
			"age":    30,
			"active": true,
			"scores": []int{1, 2, 3},
		}

		// Encode
		err := jb.EncodeJSON(testData)
		assert.NoError(t, err)
		assert.True(t, jb.Len() > 0)

		// Decode
		jb.Reset()
		err = json.NewEncoder(jb).Encode(testData)
		assert.NoError(t, err)

		var decoded map[string]interface{}
		err = jb.DecodeJSON(&decoded)
		assert.NoError(t, err)

		assert.Equal(t, "test", decoded["name"])
		assert.Equal(t, float64(30), decoded["age"])
		assert.Equal(t, true, decoded["active"])
	})
}

func TestGlobalPools(t *testing.T) {
	t.Run("ConvenienceFunctions", func(t *testing.T) {
		// Test byte buffer convenience functions
		buf := GetByteBuffer()
		assert.NotNil(t, buf)
		buf.WriteString("test")
		PutByteBuffer(buf)

		// Test string builder convenience functions
		sb := GetStringBuilder()
		assert.NotNil(t, sb)
		sb.WriteString("test")
		PutStringBuilder(sb)

		// Test JSON buffer convenience functions
		jb := GetJSONBuffer()
		assert.NotNil(t, jb)
		PutJSONBuffer(jb)
	})

	t.Run("WithFunctions", func(t *testing.T) {
		// Test WithByteBuffer
		var result string
		err := WithByteBuffer(func(buf *bytes.Buffer) error {
			buf.WriteString("test data")
			result = buf.String()
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "test data", result)

		// Test WithStringBuilder
		err = WithStringBuilder(func(sb *StringBuilder) error {
			sb.WriteString("built string")
			result = sb.String()
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "built string", result)

		// Test WithJSONBuffer
		err = WithJSONBuffer(func(jb *JSONBuffer) error {
			testObj := map[string]string{"key": "value"}
			return jb.EncodeJSON(testObj)
		})
		assert.NoError(t, err)
	})
}

func BenchmarkPools(b *testing.B) {
	pools := NewCommonPools()

	b.Run("ByteBufferPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := pools.GetByteBuffer()
			buf.WriteString("benchmark data")
			pools.PutByteBuffer(buf)
		}
	})

	b.Run("ByteBufferDirect", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := &bytes.Buffer{}
			buf.WriteString("benchmark data")
		}
	})

	b.Run("StringBuilderPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sb := pools.GetStringBuilder()
			sb.WriteString("benchmark")
			sb.WriteString(" data")
			pools.PutStringBuilder(sb)
		}
	})

	b.Run("StringBuilderDirect", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sb := &StringBuilder{builder: &bytes.Buffer{}}
			sb.WriteString("benchmark")
			sb.WriteString(" data")
		}
	})

	b.Run("IntSlicePool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice := pools.GetIntSlice()
			slice = append(slice, 1, 2, 3, 4, 5)
			pools.PutIntSlice(slice)
		}
	})

	b.Run("IntSliceDirect", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice := make([]int, 0, 64)
			_ = append(slice, 1, 2, 3, 4, 5) // Benchmark slice append operation
		}
	})
}
