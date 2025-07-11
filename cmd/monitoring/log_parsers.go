package monitoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// LogParser interface for different log formats
type LogParser interface {
	Parse(data []byte) (*LogEntry, error)
	CanParse(data []byte) bool
	Name() string
}

// ParseProcessor processes logs through multiple parsers
type ParseProcessor struct {
	name    string
	config  *ProcessorConfig
	parsers []LogParser
}

// JSONLogParser parses JSON formatted logs
type JSONLogParser struct {
	name string
}

// SyslogParser parses syslog format logs (RFC3164/RFC5424)
type SyslogParser struct {
	name        string
	rfc3164     *regexp.Regexp
	rfc5424     *regexp.Regexp
	timeFormats []string
}

// CommonLogParser parses Apache/nginx common log format
type CommonLogParser struct {
	name    string
	pattern *regexp.Regexp
}

// CustomPatternParser uses configurable regex patterns
type CustomPatternParser struct {
	name     string
	patterns []*LogPattern
}

// GrokParser supports Logstash grok patterns
type GrokParser struct {
	name         string
	patterns     map[string]*regexp.Regexp
	grokPatterns map[string]string
}

// MultilineParser handles multi-line log entries
type MultilineParser struct {
	name            string
	startPattern    *regexp.Regexp
	continuePattern *regexp.Regexp
	endPattern      *regexp.Regexp
	buffer          bytes.Buffer
	timeout         time.Duration
}

// LogPattern represents a custom parsing pattern
type LogPattern struct {
	Name       string            `yaml:"name" json:"name"`
	Pattern    string            `yaml:"pattern" json:"pattern"`
	Regex      *regexp.Regexp    `yaml:"-" json:"-"`
	Fields     map[string]string `yaml:"fields" json:"fields"`
	TimeFormat string            `yaml:"time_format" json:"time_format"`
}

// ParserConfig represents parser configuration
type ParserConfig struct {
	Type     string                 `yaml:"type" json:"type"`
	Enabled  bool                   `yaml:"enabled" json:"enabled"`
	Priority int                    `yaml:"priority" json:"priority"`
	Settings map[string]interface{} `yaml:"settings" json:"settings"`
}

// NewParseProcessor creates a new parse processor
func NewParseProcessor(name string, config *ProcessorConfig) (*ParseProcessor, error) {
	pp := &ParseProcessor{
		name:    name,
		config:  config,
		parsers: make([]LogParser, 0),
	}

	settings := config.Settings

	// Initialize parsers based on configuration
	if parsers, ok := settings["parsers"].([]interface{}); ok {
		for _, parserConfig := range parsers {
			if parserMap, ok := parserConfig.(map[string]interface{}); ok {
				parser, err := pp.createParser(parserMap)
				if err != nil {
					continue // Skip invalid parsers
				}
				pp.parsers = append(pp.parsers, parser)
			}
		}
	}

	// Add default JSON parser if no parsers configured
	if len(pp.parsers) == 0 {
		pp.parsers = append(pp.parsers, NewJSONLogParser())
	}

	return pp, nil
}

func (pp *ParseProcessor) createParser(config map[string]interface{}) (LogParser, error) {
	parserType, ok := config["type"].(string)
	if !ok {
		return nil, fmt.Errorf("parser type not specified")
	}

	switch parserType {
	case "json":
		return NewJSONLogParser(), nil
	case "syslog":
		return NewSyslogParser(), nil
	case "common_log":
		return NewCommonLogParser(), nil
	case "custom":
		return NewCustomPatternParser(config)
	case "grok":
		return NewGrokParser(config)
	case "multiline":
		return NewMultilineParser(config)
	default:
		return nil, fmt.Errorf("unsupported parser type: %s", parserType)
	}
}

func (pp *ParseProcessor) Process(entry *LogEntry) (*LogEntry, error) {
	// Try each parser in priority order
	originalMessage := entry.Message
	messageBytes := []byte(originalMessage)

	for _, parser := range pp.parsers {
		if parser.CanParse(messageBytes) {
			parsedEntry, err := parser.Parse(messageBytes)
			if err != nil {
				continue // Try next parser
			}

			// Merge parsed data with original entry
			return pp.mergeEntries(entry, parsedEntry), nil
		}
	}

	// If no parser worked, return original entry
	return entry, nil
}

func (pp *ParseProcessor) mergeEntries(original, parsed *LogEntry) *LogEntry {
	merged := &LogEntry{
		Timestamp: original.Timestamp,
		Level:     original.Level,
		Message:   original.Message,
		Logger:    original.Logger,
		Fields:    make(map[string]interface{}),
		Labels:    make(map[string]string),
		TraceID:   original.TraceID,
		SpanID:    original.SpanID,
		Source:    original.Source,
	}

	// Copy original fields
	for k, v := range original.Fields {
		merged.Fields[k] = v
	}

	// Copy original labels
	for k, v := range original.Labels {
		merged.Labels[k] = v
	}

	// Add parsed fields (overwrite originals)
	if parsed.Fields != nil {
		for k, v := range parsed.Fields {
			merged.Fields[k] = v
		}
	}

	// Add parsed labels
	if parsed.Labels != nil {
		for k, v := range parsed.Labels {
			merged.Labels[k] = v
		}
	}

	// Use parsed values if they exist
	if !parsed.Timestamp.IsZero() {
		merged.Timestamp = parsed.Timestamp
	}
	if parsed.Level != "" {
		merged.Level = parsed.Level
	}
	if parsed.Logger != "" {
		merged.Logger = parsed.Logger
	}
	if parsed.TraceID != "" {
		merged.TraceID = parsed.TraceID
	}
	if parsed.SpanID != "" {
		merged.SpanID = parsed.SpanID
	}

	return merged
}

func (pp *ParseProcessor) Name() string {
	return pp.name
}

// NewJSONLogParser creates a JSON log parser
func NewJSONLogParser() *JSONLogParser {
	return &JSONLogParser{name: "json"}
}

func (jp *JSONLogParser) Parse(data []byte) (*LogEntry, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	entry := &LogEntry{
		Fields: make(map[string]interface{}),
		Labels: make(map[string]string),
	}

	// Extract standard fields
	if timestamp, ok := parsed["timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
			entry.Timestamp = t
		}
	}
	if level, ok := parsed["level"].(string); ok {
		entry.Level = level
	}
	if message, ok := parsed["message"].(string); ok {
		entry.Message = message
	}
	if logger, ok := parsed["logger"].(string); ok {
		entry.Logger = logger
	}
	if traceID, ok := parsed["trace_id"].(string); ok {
		entry.TraceID = traceID
	}
	if spanID, ok := parsed["span_id"].(string); ok {
		entry.SpanID = spanID
	}

	// Extract fields object
	if fields, ok := parsed["fields"].(map[string]interface{}); ok {
		entry.Fields = fields
	}

	// Extract labels object
	if labels, ok := parsed["labels"].(map[string]interface{}); ok {
		entry.Labels = make(map[string]string)
		for k, v := range labels {
			if str, ok := v.(string); ok {
				entry.Labels[k] = str
			}
		}
	}

	// Add any remaining fields to Fields map
	for k, v := range parsed {
		if k != "timestamp" && k != "level" && k != "message" && k != "logger" &&
			k != "trace_id" && k != "span_id" && k != "fields" && k != "labels" {
			entry.Fields[k] = v
		}
	}

	return entry, nil
}

func (jp *JSONLogParser) CanParse(data []byte) bool {
	data = bytes.TrimSpace(data)
	return len(data) > 0 && data[0] == '{'
}

func (jp *JSONLogParser) Name() string {
	return jp.name
}

// NewSyslogParser creates a syslog parser
func NewSyslogParser() *SyslogParser {
	// RFC3164 pattern: <PRI>MMM DD HH:MM:SS HOSTNAME TAG: MESSAGE
	rfc3164 := regexp.MustCompile(`^<(\d+)>(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+(\S+)\s+([^:]+):\s*(.*)$`)

	// RFC5424 pattern: <PRI>VERSION TIMESTAMP HOSTNAME APP-NAME PROCID MSGID SD MESSAGE
	rfc5424 := regexp.MustCompile(`^<(\d+)>(\d+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s*(.*)$`)

	return &SyslogParser{
		name:    "syslog",
		rfc3164: rfc3164,
		rfc5424: rfc5424,
		timeFormats: []string{
			"Jan 2 15:04:05",
			"Jan  2 15:04:05",
			time.RFC3339,
			"2006-01-02T15:04:05.000000+00:00",
		},
	}
}

func (sp *SyslogParser) Parse(data []byte) (*LogEntry, error) {
	line := string(bytes.TrimSpace(data))

	// Try RFC5424 first
	if matches := sp.rfc5424.FindStringSubmatch(line); matches != nil {
		return sp.parseRFC5424(matches)
	}

	// Try RFC3164
	if matches := sp.rfc3164.FindStringSubmatch(line); matches != nil {
		return sp.parseRFC3164(matches)
	}

	return nil, fmt.Errorf("line does not match syslog format")
}

func (sp *SyslogParser) parseRFC3164(matches []string) (*LogEntry, error) {
	entry := &LogEntry{
		Fields: make(map[string]interface{}),
		Labels: make(map[string]string),
	}

	// Parse priority
	if priority, err := strconv.Atoi(matches[1]); err == nil {
		facility := priority >> 3
		severity := priority & 7
		entry.Fields["syslog_facility"] = facility
		entry.Fields["syslog_severity"] = severity
		entry.Level = sp.severityToLevel(severity)
	}

	// Parse timestamp
	timeStr := matches[2]
	for _, format := range sp.timeFormats {
		if t, err := time.Parse(format, timeStr); err == nil {
			// For formats without year, use current year
			if format == "Jan 2 15:04:05" || format == "Jan  2 15:04:05" {
				now := time.Now()
				t = t.AddDate(now.Year(), 0, 0)
			}
			entry.Timestamp = t
			break
		}
	}

	// Parse hostname, tag, message
	entry.Fields["syslog_hostname"] = matches[3]
	entry.Logger = matches[4]
	entry.Message = matches[5]

	return entry, nil
}

func (sp *SyslogParser) parseRFC5424(matches []string) (*LogEntry, error) {
	entry := &LogEntry{
		Fields: make(map[string]interface{}),
		Labels: make(map[string]string),
	}

	// Parse priority
	if priority, err := strconv.Atoi(matches[1]); err == nil {
		facility := priority >> 3
		severity := priority & 7
		entry.Fields["syslog_facility"] = facility
		entry.Fields["syslog_severity"] = severity
		entry.Level = sp.severityToLevel(severity)
	}

	// Parse version
	entry.Fields["syslog_version"] = matches[2]

	// Parse timestamp
	if t, err := time.Parse(time.RFC3339, matches[3]); err == nil {
		entry.Timestamp = t
	}

	// Parse structured data
	entry.Fields["syslog_hostname"] = matches[4]
	entry.Fields["syslog_app_name"] = matches[5]
	entry.Fields["syslog_proc_id"] = matches[6]
	entry.Fields["syslog_msg_id"] = matches[7]
	entry.Fields["syslog_structured_data"] = matches[8]
	entry.Logger = matches[5] // Use app-name as logger
	entry.Message = matches[9]

	return entry, nil
}

func (sp *SyslogParser) severityToLevel(severity int) string {
	switch severity {
	case 0, 1, 2: // Emergency, Alert, Critical
		return "error"
	case 3: // Error
		return "error"
	case 4: // Warning
		return "warn"
	case 5, 6: // Notice, Informational
		return "info"
	case 7: // Debug
		return "debug"
	default:
		return "info"
	}
}

func (sp *SyslogParser) CanParse(data []byte) bool {
	line := string(bytes.TrimSpace(data))
	return strings.HasPrefix(line, "<") &&
		(sp.rfc3164.MatchString(line) || sp.rfc5424.MatchString(line))
}

func (sp *SyslogParser) Name() string {
	return sp.name
}

// NewCommonLogParser creates a common log format parser
func NewCommonLogParser() *CommonLogParser {
	// Combined log format: IP - - [timestamp] "method URL version" status size "referer" "user-agent"
	pattern := regexp.MustCompile(`^(\S+) (\S+) (\S+) \[([^\]]+)\] "([^"]*)" (\d+) (\S+)(?: "([^"]*)" "([^"]*)")?`)

	return &CommonLogParser{
		name:    "common_log",
		pattern: pattern,
	}
}

func (clp *CommonLogParser) Parse(data []byte) (*LogEntry, error) {
	line := string(bytes.TrimSpace(data))
	matches := clp.pattern.FindStringSubmatch(line)
	if matches == nil {
		return nil, fmt.Errorf("line does not match common log format")
	}

	entry := &LogEntry{
		Fields: make(map[string]interface{}),
		Labels: make(map[string]string),
		Level:  "info",
		Logger: "access_log",
	}

	// Parse fields
	entry.Fields["remote_addr"] = matches[1]
	entry.Fields["remote_user"] = matches[3]

	// Parse timestamp
	if t, err := time.Parse("02/Jan/2006:15:04:05 -0700", matches[4]); err == nil {
		entry.Timestamp = t
	}

	// Parse request
	if len(matches[5]) > 0 {
		requestParts := strings.SplitN(matches[5], " ", 3)
		if len(requestParts) >= 2 {
			entry.Fields["method"] = requestParts[0]
			entry.Fields["uri"] = requestParts[1]
			if len(requestParts) >= 3 {
				entry.Fields["protocol"] = requestParts[2]
			}
		}
	}

	// Parse status and size
	if status, err := strconv.Atoi(matches[6]); err == nil {
		entry.Fields["status"] = status

		// Set log level based on status code
		if status >= 400 && status < 500 {
			entry.Level = "warn"
		} else if status >= 500 {
			entry.Level = "error"
		}
	}

	if matches[7] != "-" {
		if size, err := strconv.Atoi(matches[7]); err == nil {
			entry.Fields["bytes_sent"] = size
		}
	}

	// Parse optional referer and user-agent
	if len(matches) > 8 && matches[8] != "" {
		entry.Fields["referer"] = matches[8]
	}
	if len(matches) > 9 && matches[9] != "" {
		entry.Fields["user_agent"] = matches[9]
	}

	// Create message
	entry.Message = fmt.Sprintf("%s %s %s %s",
		entry.Fields["method"],
		entry.Fields["uri"],
		entry.Fields["status"],
		entry.Fields["remote_addr"])

	return entry, nil
}

func (clp *CommonLogParser) CanParse(data []byte) bool {
	line := string(bytes.TrimSpace(data))
	return clp.pattern.MatchString(line)
}

func (clp *CommonLogParser) Name() string {
	return clp.name
}

// NewCustomPatternParser creates a custom pattern parser
func NewCustomPatternParser(config map[string]interface{}) (*CustomPatternParser, error) {
	cpp := &CustomPatternParser{
		name:     "custom",
		patterns: make([]*LogPattern, 0),
	}

	if patterns, ok := config["patterns"].([]interface{}); ok {
		for _, patternConfig := range patterns {
			if patternMap, ok := patternConfig.(map[string]interface{}); ok {
				pattern, err := cpp.createPattern(patternMap)
				if err != nil {
					continue
				}
				cpp.patterns = append(cpp.patterns, pattern)
			}
		}
	}

	return cpp, nil
}

func (cpp *CustomPatternParser) createPattern(config map[string]interface{}) (*LogPattern, error) {
	pattern := &LogPattern{
		Fields: make(map[string]string),
	}

	if name, ok := config["name"].(string); ok {
		pattern.Name = name
	}

	patternStr, ok := config["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("pattern string required")
	}
	pattern.Pattern = patternStr

	regex, err := regexp.Compile(patternStr)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}
	pattern.Regex = regex

	if fields, ok := config["fields"].(map[string]interface{}); ok {
		for k, v := range fields {
			if str, ok := v.(string); ok {
				pattern.Fields[k] = str
			}
		}
	}

	if timeFormat, ok := config["time_format"].(string); ok {
		pattern.TimeFormat = timeFormat
	}

	return pattern, nil
}

func (cpp *CustomPatternParser) Parse(data []byte) (*LogEntry, error) {
	line := string(bytes.TrimSpace(data))

	for _, pattern := range cpp.patterns {
		if matches := pattern.Regex.FindStringSubmatch(line); matches != nil {
			return cpp.parseWithPattern(pattern, matches)
		}
	}

	return nil, fmt.Errorf("no pattern matched the input")
}

func (cpp *CustomPatternParser) parseWithPattern(pattern *LogPattern, matches []string) (*LogEntry, error) {
	entry := &LogEntry{
		Fields: make(map[string]interface{}),
		Labels: make(map[string]string),
		Level:  "info",
		Logger: "custom",
	}

	// Map captured groups to fields
	for i, match := range matches[1:] { // Skip first match (full string)
		fieldIndex := fmt.Sprintf("field_%d", i+1)
		if fieldName, exists := pattern.Fields[fieldIndex]; exists {
			entry.Fields[fieldName] = match

			// Handle special fields
			switch fieldName {
			case "timestamp":
				if pattern.TimeFormat != "" {
					if t, err := time.Parse(pattern.TimeFormat, match); err == nil {
						entry.Timestamp = t
					}
				}
			case "level":
				entry.Level = strings.ToLower(match)
			case "logger":
				entry.Logger = match
			case "message":
				entry.Message = match
			}
		} else {
			// Add as numbered field if no mapping exists
			entry.Fields[fieldIndex] = match
		}
	}

	// Set message if not already set
	if entry.Message == "" {
		entry.Message = matches[0]
	}

	return entry, nil
}

func (cpp *CustomPatternParser) CanParse(data []byte) bool {
	line := string(bytes.TrimSpace(data))
	for _, pattern := range cpp.patterns {
		if pattern.Regex.MatchString(line) {
			return true
		}
	}
	return false
}

func (cpp *CustomPatternParser) Name() string {
	return cpp.name
}

// NewGrokParser creates a grok pattern parser
func NewGrokParser(config map[string]interface{}) (*GrokParser, error) {
	gp := &GrokParser{
		name:         "grok",
		patterns:     make(map[string]*regexp.Regexp),
		grokPatterns: getBuiltinGrokPatterns(),
	}

	// Add custom patterns from config
	if patterns, ok := config["patterns"].(map[string]interface{}); ok {
		for name, pattern := range patterns {
			if patternStr, ok := pattern.(string); ok {
				if regex, err := regexp.Compile(patternStr); err == nil {
					gp.patterns[name] = regex
				}
			}
		}
	}

	// Add built-in patterns
	for name, pattern := range gp.grokPatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			gp.patterns[name] = regex
		}
	}

	return gp, nil
}

func (gp *GrokParser) Parse(data []byte) (*LogEntry, error) {
	line := string(bytes.TrimSpace(data))

	// Try each pattern
	for name, pattern := range gp.patterns {
		if matches := pattern.FindStringSubmatch(line); matches != nil {
			return gp.parseWithGrok(name, pattern, matches)
		}
	}

	return nil, fmt.Errorf("no grok pattern matched")
}

func (gp *GrokParser) parseWithGrok(patternName string, pattern *regexp.Regexp, matches []string) (*LogEntry, error) {
	entry := &LogEntry{
		Fields: make(map[string]interface{}),
		Labels: make(map[string]string),
		Level:  "info",
		Logger: "grok",
	}

	// Get subexpression names
	names := pattern.SubexpNames()

	for i, match := range matches[1:] {
		if i+1 < len(names) && names[i+1] != "" {
			fieldName := names[i+1]
			entry.Fields[fieldName] = match

			// Handle special fields
			switch fieldName {
			case "timestamp":
				// Try common timestamp formats
				timeFormats := []string{
					time.RFC3339,
					"2006-01-02 15:04:05",
					"Jan 2 15:04:05",
					"02/Jan/2006:15:04:05 -0700",
				}
				for _, format := range timeFormats {
					if t, err := time.Parse(format, match); err == nil {
						entry.Timestamp = t
						break
					}
				}
			case "level", "loglevel":
				entry.Level = strings.ToLower(match)
			case "logger", "program":
				entry.Logger = match
			case "message", "msg":
				entry.Message = match
			}
		}
	}

	if entry.Message == "" {
		entry.Message = matches[0]
	}

	entry.Labels["grok_pattern"] = patternName

	return entry, nil
}

func (gp *GrokParser) CanParse(data []byte) bool {
	line := string(bytes.TrimSpace(data))
	for _, pattern := range gp.patterns {
		if pattern.MatchString(line) {
			return true
		}
	}
	return false
}

func (gp *GrokParser) Name() string {
	return gp.name
}

// NewMultilineParser creates a multiline parser
func NewMultilineParser(config map[string]interface{}) (*MultilineParser, error) {
	mp := &MultilineParser{
		name:    "multiline",
		timeout: 5 * time.Second,
	}

	if startPattern, ok := config["start_pattern"].(string); ok {
		if regex, err := regexp.Compile(startPattern); err == nil {
			mp.startPattern = regex
		}
	}

	if continuePattern, ok := config["continue_pattern"].(string); ok {
		if regex, err := regexp.Compile(continuePattern); err == nil {
			mp.continuePattern = regex
		}
	}

	if endPattern, ok := config["end_pattern"].(string); ok {
		if regex, err := regexp.Compile(endPattern); err == nil {
			mp.endPattern = regex
		}
	}

	if timeout, ok := config["timeout"].(string); ok {
		if duration, err := time.ParseDuration(timeout); err == nil {
			mp.timeout = duration
		}
	}

	return mp, nil
}

func (mp *MultilineParser) Parse(data []byte) (*LogEntry, error) {
	line := string(bytes.TrimSpace(data))

	// Check if this starts a new multi-line entry
	if mp.startPattern != nil && mp.startPattern.MatchString(line) {
		// Return any buffered entry first
		if mp.buffer.Len() > 0 {
			bufferedContent := mp.buffer.String()
			mp.buffer.Reset()
			mp.buffer.WriteString(line)

			// Create entry from buffered content
			return &LogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Logger:    "multiline",
				Message:   bufferedContent,
				Fields:    map[string]interface{}{"multiline": true},
			}, nil
		} else {
			mp.buffer.WriteString(line)
			return nil, nil // Wait for more lines
		}
	}

	// Check if this continues a multi-line entry
	if mp.continuePattern != nil && mp.continuePattern.MatchString(line) {
		if mp.buffer.Len() > 0 {
			mp.buffer.WriteString("\n" + line)
			return nil, nil // Wait for more lines
		}
	}

	// Check if this ends a multi-line entry
	if mp.endPattern != nil && mp.endPattern.MatchString(line) {
		if mp.buffer.Len() > 0 {
			mp.buffer.WriteString("\n" + line)
			content := mp.buffer.String()
			mp.buffer.Reset()

			return &LogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Logger:    "multiline",
				Message:   content,
				Fields:    map[string]interface{}{"multiline": true},
			}, nil
		}
	}

	// If we have buffered content but this line doesn't match patterns,
	// return the buffered content and start fresh
	if mp.buffer.Len() > 0 {
		content := mp.buffer.String()
		mp.buffer.Reset()
		mp.buffer.WriteString(line)

		return &LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Logger:    "multiline",
			Message:   content,
			Fields:    map[string]interface{}{"multiline": true},
		}, nil
	}

	// Single line entry
	return &LogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Logger:    "single",
		Message:   line,
		Fields:    make(map[string]interface{}),
	}, nil
}

func (mp *MultilineParser) CanParse(data []byte) bool {
	return true // Multiline parser can handle any input
}

func (mp *MultilineParser) Name() string {
	return mp.name
}

// getBuiltinGrokPatterns returns common grok patterns
func getBuiltinGrokPatterns() map[string]string {
	return map[string]string{
		"COMMONAPACHELOG": `(?P<clientip>\S+) \S+ (?P<auth>\S+) \[(?P<timestamp>[^\]]+)\] "(?P<verb>\S+) (?P<request>\S+) HTTP/(?P<httpversion>[^"]*)" (?P<response>\d+) (?P<bytes>\d+)`,

		"COMBINEDAPACHELOG": `(?P<clientip>\S+) \S+ (?P<auth>\S+) \[(?P<timestamp>[^\]]+)\] "(?P<verb>\S+) (?P<request>\S+) HTTP/(?P<httpversion>[^"]*)" (?P<response>\d+) (?P<bytes>\d+) "(?P<referrer>[^"]*)" "(?P<agent>[^"]*)"`,

		"HTTPD20_ERRORLOG": `\[(?P<timestamp>[^\]]+)\] \[(?P<loglevel>\S+)\] (?P<message>.*)`,

		"HTTPD24_ERRORLOG": `\[(?P<timestamp>[^\]]+)\] \[(?P<module>\S+):(?P<loglevel>\S+)\] \[pid (?P<pid>\d+)\] (?P<message>.*)`,

		"NGINXACCESS": `(?P<clientip>\S+) - (?P<auth>\S+) \[(?P<timestamp>[^\]]+)\] "(?P<verb>\S+) (?P<request>\S+) HTTP/(?P<httpversion>[^"]*)" (?P<response>\d+) (?P<bytes>\d+) "(?P<referrer>[^"]*)" "(?P<agent>[^"]*)"`,

		"SYSLOGBASE": `(?P<timestamp>\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}) (?P<logsource>\S+) (?P<program>\w+)(?:\[(?P<pid>\d+)\])?: (?P<message>.*)`,

		"JAVALOGBACK": `(?P<timestamp>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}.\d{3}) \[(?P<thread>[^\]]+)\] (?P<loglevel>\S+)\s+(?P<logger>\S+) - (?P<message>.*)`,

		"GOLOGFMT": `time="(?P<timestamp>[^"]+)" level=(?P<loglevel>\S+) msg="(?P<message>[^"]*)"`,
	}
}
