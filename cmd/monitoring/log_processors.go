package monitoring

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
)

// FilterProcessor filters log entries based on configured criteria
type FilterProcessor struct {
	name            string
	config          *ProcessorConfig
	levelFilter     map[string]bool
	messagePatterns []*regexp.Regexp
	fieldFilters    map[string]interface{}
}

// TransformProcessor transforms log entries
type TransformProcessor struct {
	name            string
	config          *ProcessorConfig
	fieldMappings   map[string]string
	messageTemplate string
	addFields       map[string]interface{}
	removeFields    []string
}

// EnrichProcessor enriches log entries with additional context
type EnrichProcessor struct {
	name           string
	config         *ProcessorConfig
	staticFields   map[string]interface{}
	timestampField string
	hostField      string
	processField   string
}

// SampleProcessor samples log entries at a configured rate
type SampleProcessor struct {
	name       string
	config     *ProcessorConfig
	sampleRate float64
	random     *rand.Rand
}

// NewFilterProcessor creates a new filter processor
func NewFilterProcessor(name string, config *ProcessorConfig) (*FilterProcessor, error) {
	fp := &FilterProcessor{
		name:         name,
		config:       config,
		levelFilter:  make(map[string]bool),
		fieldFilters: make(map[string]interface{}),
	}

	settings := config.Settings

	// Configure level filter
	if levels, ok := settings["allowed_levels"].([]interface{}); ok {
		for _, level := range levels {
			if levelStr, ok := level.(string); ok {
				fp.levelFilter[levelStr] = true
			}
		}
	}

	// Configure message pattern filters
	if patterns, ok := settings["message_patterns"].([]interface{}); ok {
		for _, pattern := range patterns {
			if patternStr, ok := pattern.(string); ok {
				if regex, err := regexp.Compile(patternStr); err == nil {
					fp.messagePatterns = append(fp.messagePatterns, regex)
				}
			}
		}
	}

	// Configure field filters
	if fields, ok := settings["field_filters"].(map[string]interface{}); ok {
		fp.fieldFilters = fields
	}

	return fp, nil
}

func (fp *FilterProcessor) Process(entry *LogEntry) (*LogEntry, error) {
	// Level filter
	if len(fp.levelFilter) > 0 && !fp.levelFilter[entry.Level] {
		return nil, nil // Filter out
	}

	// Message pattern filter
	if len(fp.messagePatterns) > 0 {
		matched := false
		for _, pattern := range fp.messagePatterns {
			if pattern.MatchString(entry.Message) {
				matched = true
				break
			}
		}
		if !matched {
			return nil, nil // Filter out
		}
	}

	// Field filters
	for fieldName, expectedValue := range fp.fieldFilters {
		if actualValue, exists := entry.Fields[fieldName]; exists {
			if actualValue != expectedValue {
				return nil, nil // Filter out
			}
		}
	}

	return entry, nil
}

func (fp *FilterProcessor) Name() string {
	return fp.name
}

// NewTransformProcessor creates a new transform processor
func NewTransformProcessor(name string, config *ProcessorConfig) (*TransformProcessor, error) {
	tp := &TransformProcessor{
		name:          name,
		config:        config,
		fieldMappings: make(map[string]string),
		addFields:     make(map[string]interface{}),
	}

	settings := config.Settings

	// Configure field mappings
	if mappings, ok := settings["field_mappings"].(map[string]interface{}); ok {
		for from, to := range mappings {
			if toStr, ok := to.(string); ok {
				tp.fieldMappings[from] = toStr
			}
		}
	}

	// Configure message template
	if template, ok := settings["message_template"].(string); ok {
		tp.messageTemplate = template
	}

	// Configure fields to add
	if fields, ok := settings["add_fields"].(map[string]interface{}); ok {
		tp.addFields = fields
	}

	// Configure fields to remove
	if fields, ok := settings["remove_fields"].([]interface{}); ok {
		for _, field := range fields {
			if fieldStr, ok := field.(string); ok {
				tp.removeFields = append(tp.removeFields, fieldStr)
			}
		}
	}

	return tp, nil
}

func (tp *TransformProcessor) Process(entry *LogEntry) (*LogEntry, error) {
	// Create a copy to avoid modifying original
	transformed := &LogEntry{
		Timestamp: entry.Timestamp,
		Level:     entry.Level,
		Message:   entry.Message,
		Logger:    entry.Logger,
		Fields:    make(map[string]interface{}),
		Labels:    make(map[string]string),
		TraceID:   entry.TraceID,
		SpanID:    entry.SpanID,
		Source:    entry.Source,
	}

	// Copy fields
	for k, v := range entry.Fields {
		transformed.Fields[k] = v
	}

	// Copy labels
	for k, v := range entry.Labels {
		transformed.Labels[k] = v
	}

	// Apply field mappings
	for from, to := range tp.fieldMappings {
		if value, exists := transformed.Fields[from]; exists {
			transformed.Fields[to] = value
			delete(transformed.Fields, from)
		}
	}

	// Transform message if template is provided
	if tp.messageTemplate != "" {
		transformed.Message = tp.applyMessageTemplate(transformed)
	}

	// Add fields
	for k, v := range tp.addFields {
		transformed.Fields[k] = v
	}

	// Remove fields
	for _, field := range tp.removeFields {
		delete(transformed.Fields, field)
	}

	return transformed, nil
}

func (tp *TransformProcessor) applyMessageTemplate(entry *LogEntry) string {
	message := tp.messageTemplate

	// Replace placeholders
	message = strings.ReplaceAll(message, "{{.Message}}", entry.Message)
	message = strings.ReplaceAll(message, "{{.Level}}", entry.Level)
	message = strings.ReplaceAll(message, "{{.Logger}}", entry.Logger)
	message = strings.ReplaceAll(message, "{{.Timestamp}}", entry.Timestamp.Format(time.RFC3339))

	// Replace field placeholders
	for k, v := range entry.Fields {
		placeholder := fmt.Sprintf("{{.Fields.%s}}", k)
		message = strings.ReplaceAll(message, placeholder, fmt.Sprintf("%v", v))
	}

	return message
}

func (tp *TransformProcessor) Name() string {
	return tp.name
}

// NewEnrichProcessor creates a new enrich processor
func NewEnrichProcessor(name string, config *ProcessorConfig) (*EnrichProcessor, error) {
	ep := &EnrichProcessor{
		name:         name,
		config:       config,
		staticFields: make(map[string]interface{}),
	}

	settings := config.Settings

	// Configure static fields
	if fields, ok := settings["static_fields"].(map[string]interface{}); ok {
		ep.staticFields = fields
	}

	// Configure timestamp field
	if field, ok := settings["timestamp_field"].(string); ok {
		ep.timestampField = field
	}

	// Configure host field
	if field, ok := settings["host_field"].(string); ok {
		ep.hostField = field
	}

	// Configure process field
	if field, ok := settings["process_field"].(string); ok {
		ep.processField = field
	}

	return ep, nil
}

func (ep *EnrichProcessor) Process(entry *LogEntry) (*LogEntry, error) {
	// Create a copy
	enriched := &LogEntry{
		Timestamp: entry.Timestamp,
		Level:     entry.Level,
		Message:   entry.Message,
		Logger:    entry.Logger,
		Fields:    make(map[string]interface{}),
		Labels:    make(map[string]string),
		TraceID:   entry.TraceID,
		SpanID:    entry.SpanID,
		Source:    entry.Source,
	}

	// Copy existing fields and labels
	for k, v := range entry.Fields {
		enriched.Fields[k] = v
	}
	for k, v := range entry.Labels {
		enriched.Labels[k] = v
	}

	// Add static fields
	for k, v := range ep.staticFields {
		enriched.Fields[k] = v
	}

	// Add timestamp field if configured
	if ep.timestampField != "" {
		enriched.Fields[ep.timestampField] = entry.Timestamp.Unix()
	}

	// Add host information if configured
	if ep.hostField != "" {
		hostname := getHostname()
		enriched.Fields[ep.hostField] = hostname
	}

	// Add process information if configured
	if ep.processField != "" {
		processInfo := getProcessInfo()
		enriched.Fields[ep.processField] = processInfo
	}

	// Add environment information
	enriched.Fields["environment"] = getLogEnvironment()
	enriched.Fields["service"] = "gzh-manager"
	enriched.Fields["version"] = getVersion()

	return enriched, nil
}

func (ep *EnrichProcessor) Name() string {
	return ep.name
}

// NewSampleProcessor creates a new sample processor
func NewSampleProcessor(name string, config *ProcessorConfig) (*SampleProcessor, error) {
	sp := &SampleProcessor{
		name:   name,
		config: config,
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	settings := config.Settings

	// Configure sample rate
	if rate, ok := settings["sample_rate"].(float64); ok {
		sp.sampleRate = rate
	} else {
		sp.sampleRate = 1.0 // Default to 100% sampling
	}

	// Ensure sample rate is between 0 and 1
	if sp.sampleRate < 0 {
		sp.sampleRate = 0
	} else if sp.sampleRate > 1 {
		sp.sampleRate = 1
	}

	return sp, nil
}

func (sp *SampleProcessor) Process(entry *LogEntry) (*LogEntry, error) {
	// Always pass through error and warn levels
	if entry.Level == "error" || entry.Level == "warn" {
		return entry, nil
	}

	// Sample based on configured rate
	if sp.random.Float64() < sp.sampleRate {
		return entry, nil
	}

	// Drop the entry
	return nil, nil
}

func (sp *SampleProcessor) Name() string {
	return sp.name
}

// Helper functions

func getHostname() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return "unknown"
}

func getProcessInfo() map[string]interface{} {
	return map[string]interface{}{
		"pid":  os.Getpid(),
		"ppid": os.Getppid(),
		"uid":  os.Getuid(),
		"gid":  os.Getgid(),
	}
}

func getLogEnvironment() string {
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return env
	}
	if env := os.Getenv("ENV"); env != "" {
		return env
	}
	return "development"
}

func getVersion() string {
	if version := os.Getenv("GZH_VERSION"); version != "" {
		return version
	}
	return "dev"
}
