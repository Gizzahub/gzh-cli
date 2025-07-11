package monitoring

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogParsingAndIndexing(t *testing.T) {
	// Test JSON log parsing
	t.Run("JSON Log Parsing", func(t *testing.T) {
		parser := NewJSONLogParser()

		jsonLog := `{
			"timestamp": "2024-01-15T10:30:00Z",
			"level": "info",
			"message": "User login successful",
			"logger": "auth-service",
			"fields": {
				"user_id": 12345,
				"ip_address": "192.168.1.100"
			},
			"labels": {
				"environment": "production",
				"service": "authentication"
			}
		}`

		entry, err := parser.Parse([]byte(jsonLog))
		require.NoError(t, err)
		assert.Equal(t, "info", entry.Level)
		assert.Equal(t, "User login successful", entry.Message)
		assert.Equal(t, "auth-service", entry.Logger)
		assert.Equal(t, float64(12345), entry.Fields["user_id"])
		assert.Equal(t, "production", entry.Labels["environment"])
	})

	// Test Syslog parsing
	t.Run("Syslog Parsing", func(t *testing.T) {
		parser := NewSyslogParser()

		syslogMessage := "<134>Dec 25 14:30:00 server01 nginx: 192.168.1.1 - - [25/Dec/2023:14:30:00 +0000] \"GET /api/health HTTP/1.1\" 200 15"

		entry, err := parser.Parse([]byte(syslogMessage))
		require.NoError(t, err)
		assert.Equal(t, "info", entry.Level)
		assert.Equal(t, "nginx", entry.Logger)
		assert.Contains(t, entry.Message, "GET /api/health")
	})

	// Test Common Log Format parsing
	t.Run("Common Log Format Parsing", func(t *testing.T) {
		parser := NewCommonLogParser()

		accessLog := `192.168.1.100 - - [25/Dec/2023:14:30:00 +0000] "GET /api/users HTTP/1.1" 200 1024 "https://example.com" "Mozilla/5.0"`

		entry, err := parser.Parse([]byte(accessLog))
		require.NoError(t, err)
		assert.Equal(t, "info", entry.Level)
		assert.Equal(t, "access_log", entry.Logger)
		assert.Equal(t, "GET", entry.Fields["method"])
		assert.Equal(t, "/api/users", entry.Fields["uri"])
		assert.Equal(t, 200, entry.Fields["status"])
	})

	// Test Memory Indexer
	t.Run("Memory Indexer", func(t *testing.T) {
		indexConfig := &IndexConfig{
			Name: "test-logs",
			Settings: &IndexSettings{
				Shards:          1,
				Replicas:        0,
				RefreshInterval: "1s",
			},
		}

		indexer := NewMemoryIndexer("test-logs", indexConfig)
		require.NotNil(t, indexer)

		// Index some test entries
		entries := []*LogEntry{
			{
				Timestamp: time.Now(),
				Level:     "error",
				Logger:    "payment-service",
				Message:   "Payment processing failed for order 12345",
				Fields: map[string]interface{}{
					"order_id":   12345,
					"error_code": "CARD_DECLINED",
					"amount":     99.99,
				},
				Labels: map[string]string{
					"service": "payment",
					"team":    "checkout",
				},
			},
			{
				Timestamp: time.Now(),
				Level:     "info",
				Logger:    "user-service",
				Message:   "New user registered successfully",
				Fields: map[string]interface{}{
					"user_id": 67890,
					"email":   "user@example.com",
				},
				Labels: map[string]string{
					"service": "user",
					"team":    "identity",
				},
			},
			{
				Timestamp: time.Now(),
				Level:     "warn",
				Logger:    "auth-service",
				Message:   "Multiple failed login attempts detected",
				Fields: map[string]interface{}{
					"ip_address":      "192.168.1.200",
					"failed_attempts": 5,
				},
				Labels: map[string]string{
					"service": "auth",
					"team":    "security",
				},
			},
		}

		for _, entry := range entries {
			err := indexer.Index(entry)
			require.NoError(t, err)
		}

		// Test search functionality
		t.Run("Search by keyword", func(t *testing.T) {
			query := &SearchQuery{
				Query: "payment",
				Limit: 10,
			}

			result, err := indexer.Search(query)
			require.NoError(t, err)
			assert.Equal(t, int64(1), result.Total)
			assert.Len(t, result.Hits, 1)
			assert.Contains(t, result.Hits[0].Source.Message, "Payment")
		})

		t.Run("Search by level filter", func(t *testing.T) {
			query := &SearchQuery{
				Query: "",
				Filters: map[string]interface{}{
					"level": "error",
				},
				Limit: 10,
			}

			result, err := indexer.Search(query)
			require.NoError(t, err)
			assert.Equal(t, int64(1), result.Total)
			assert.Equal(t, "error", result.Hits[0].Source.Level)
		})

		t.Run("Search with aggregations", func(t *testing.T) {
			query := &SearchQuery{
				Query: "",
				Limit: 10,
				Aggregations: map[string]*Aggregation{
					"levels": {
						Type:  "terms",
						Field: "level",
						Size:  10,
					},
					"services": {
						Type:  "terms",
						Field: "logger",
						Size:  10,
					},
				},
			}

			result, err := indexer.Search(query)
			require.NoError(t, err)
			assert.Equal(t, int64(3), result.Total)

			// Check aggregations
			require.NotNil(t, result.Aggregations)
			levelsAgg := result.Aggregations["levels"]
			require.NotNil(t, levelsAgg)
			assert.Len(t, levelsAgg.Buckets, 3) // error, info, warn
		})

		// Test index statistics
		stats := indexer.GetStats()
		assert.Equal(t, "test-logs", stats.Name)
		assert.Equal(t, int64(3), stats.DocCount)
		assert.Equal(t, "green", stats.Health)
	})

	// Test Parse Processor with multiple parsers
	t.Run("Parse Processor Integration", func(t *testing.T) {
		config := &ProcessorConfig{
			Type:    "parse",
			Enabled: true,
			Settings: map[string]interface{}{
				"parsers": []interface{}{
					map[string]interface{}{
						"type": "json",
					},
					map[string]interface{}{
						"type": "syslog",
					},
					map[string]interface{}{
						"type": "common_log",
					},
				},
			},
		}

		processor, err := NewParseProcessor("test-parser", config)
		require.NoError(t, err)

		// Test with a log entry containing JSON in message
		entry := &LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Logger:    "app",
			Message:   `{"user_id": 123, "action": "login", "success": true}`,
		}

		processedEntry, err := processor.Process(entry)
		require.NoError(t, err)
		require.NotNil(t, processedEntry)

		// The JSON should be parsed and added to fields
		assert.NotNil(t, processedEntry.Fields)
	})
}

func TestGrokPatterns(t *testing.T) {
	config := map[string]interface{}{
		"patterns": map[string]interface{}{
			"COMMONAPACHELOG": `(?P<clientip>\S+) \S+ (?P<auth>\S+) \[(?P<timestamp>[^\]]+)\] "(?P<verb>\S+) (?P<request>\S+) HTTP/(?P<httpversion>[^"]*)" (?P<response>\d+) (?P<bytes>\d+)`,
		},
	}

	parser, err := NewGrokParser(config)
	require.NoError(t, err)

	logLine := `127.0.0.1 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`

	entry, err := parser.Parse([]byte(logLine))
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1", entry.Fields["clientip"])
	assert.Equal(t, "GET", entry.Fields["verb"])
	assert.Equal(t, "/apache_pb.gif", entry.Fields["request"])
	assert.Equal(t, "200", entry.Fields["response"])
}
