package monitoring

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ElasticsearchShipper ships logs to Elasticsearch
type ElasticsearchShipper struct {
	name       string
	config     *ShipperConfig
	client     *http.Client
	endpoint   string
	index      string
	docType    string
	buffer     []*LogEntry
	bufferSize int
	mutex      sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// LokiShipper ships logs to Grafana Loki
type LokiShipper struct {
	name       string
	config     *ShipperConfig
	client     *http.Client
	endpoint   string
	labels     map[string]string
	buffer     []*LogEntry
	bufferSize int
	mutex      sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// FluentdShipper ships logs to Fluentd
type FluentdShipper struct {
	name       string
	config     *ShipperConfig
	client     *http.Client
	endpoint   string
	tag        string
	buffer     []*LogEntry
	bufferSize int
	mutex      sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// HTTPShipper ships logs to generic HTTP endpoints
type HTTPShipper struct {
	name       string
	config     *ShipperConfig
	client     *http.Client
	endpoint   string
	method     string
	headers    map[string]string
	buffer     []*LogEntry
	bufferSize int
	batchSize  int
	mutex      sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewElasticsearchShipper creates a new Elasticsearch shipper
func NewElasticsearchShipper(name string, config *ShipperConfig) (*ElasticsearchShipper, error) {
	settings := config.Settings

	index := "gzh-manager-logs"
	if i, ok := settings["index"].(string); ok {
		index = i
	}

	docType := "_doc"
	if dt, ok := settings["doc_type"].(string); ok {
		docType = dt
	}

	bufferSize := 100
	if bs, ok := settings["buffer_size"].(int); ok {
		bufferSize = bs
	}

	timeout := 30 * time.Second
	if t, ok := settings["timeout"].(string); ok {
		if duration, err := time.ParseDuration(t); err == nil {
			timeout = duration
		}
	}

	client := &http.Client{
		Timeout: timeout,
	}

	return &ElasticsearchShipper{
		name:       name,
		config:     config,
		client:     client,
		endpoint:   config.Endpoint,
		index:      index,
		docType:    docType,
		buffer:     make([]*LogEntry, 0, bufferSize),
		bufferSize: bufferSize,
	}, nil
}

func (es *ElasticsearchShipper) Start(ctx context.Context) error {
	es.ctx, es.cancel = context.WithCancel(ctx)

	// Start batch shipping routine
	es.wg.Add(1)
	go es.batchShippingRoutine()

	return nil
}

func (es *ElasticsearchShipper) Ship(entries []*LogEntry) error {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	es.buffer = append(es.buffer, entries...)

	// Ship if buffer is full
	if len(es.buffer) >= es.bufferSize {
		return es.shipBuffer()
	}

	return nil
}

func (es *ElasticsearchShipper) batchShippingRoutine() {
	defer es.wg.Done()

	ticker := time.NewTicker(10 * time.Second) // Ship every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-es.ctx.Done():
			es.flushBuffer()
			return
		case <-ticker.C:
			es.mutex.Lock()
			if len(es.buffer) > 0 {
				es.shipBuffer()
			}
			es.mutex.Unlock()
		}
	}
}

func (es *ElasticsearchShipper) shipBuffer() error {
	if len(es.buffer) == 0 {
		return nil
	}

	// Create bulk request
	var bulkBody strings.Builder
	for _, entry := range es.buffer {
		// Index action
		indexAction := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": es.index,
				"_type":  es.docType,
			},
		}
		actionBytes, _ := json.Marshal(indexAction)
		bulkBody.Write(actionBytes)
		bulkBody.WriteString("\n")

		// Document
		docBytes, _ := json.Marshal(entry)
		bulkBody.Write(docBytes)
		bulkBody.WriteString("\n")
	}

	// Send bulk request
	req, err := http.NewRequestWithContext(es.ctx, "POST", es.endpoint+"/_bulk", strings.NewReader(bulkBody.String()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := es.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Elasticsearch request failed with status: %d", resp.StatusCode)
	}

	// Clear buffer on success
	es.buffer = es.buffer[:0]

	return nil
}

func (es *ElasticsearchShipper) flushBuffer() error {
	es.mutex.Lock()
	defer es.mutex.Unlock()
	return es.shipBuffer()
}

func (es *ElasticsearchShipper) Stop() error {
	if es.cancel != nil {
		es.cancel()
	}
	es.wg.Wait()
	return es.flushBuffer()
}

func (es *ElasticsearchShipper) Name() string {
	return es.name
}

// NewLokiShipper creates a new Loki shipper
func NewLokiShipper(name string, config *ShipperConfig) (*LokiShipper, error) {
	settings := config.Settings

	labels := map[string]string{
		"service": "gzh-manager",
		"env":     getShipperEnvironment(),
	}
	if l, ok := settings["labels"].(map[string]interface{}); ok {
		for k, v := range l {
			if str, ok := v.(string); ok {
				labels[k] = str
			}
		}
	}

	bufferSize := 100
	if bs, ok := settings["buffer_size"].(int); ok {
		bufferSize = bs
	}

	timeout := 30 * time.Second
	if t, ok := settings["timeout"].(string); ok {
		if duration, err := time.ParseDuration(t); err == nil {
			timeout = duration
		}
	}

	client := &http.Client{
		Timeout: timeout,
	}

	return &LokiShipper{
		name:       name,
		config:     config,
		client:     client,
		endpoint:   config.Endpoint,
		labels:     labels,
		buffer:     make([]*LogEntry, 0, bufferSize),
		bufferSize: bufferSize,
	}, nil
}

func (ls *LokiShipper) Start(ctx context.Context) error {
	ls.ctx, ls.cancel = context.WithCancel(ctx)

	// Start batch shipping routine
	ls.wg.Add(1)
	go ls.batchShippingRoutine()

	return nil
}

func (ls *LokiShipper) Ship(entries []*LogEntry) error {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()

	ls.buffer = append(ls.buffer, entries...)

	// Ship if buffer is full
	if len(ls.buffer) >= ls.bufferSize {
		return ls.shipBuffer()
	}

	return nil
}

func (ls *LokiShipper) batchShippingRoutine() {
	defer ls.wg.Done()

	ticker := time.NewTicker(5 * time.Second) // Ship every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ls.ctx.Done():
			ls.flushBuffer()
			return
		case <-ticker.C:
			ls.mutex.Lock()
			if len(ls.buffer) > 0 {
				ls.shipBuffer()
			}
			ls.mutex.Unlock()
		}
	}
}

func (ls *LokiShipper) shipBuffer() error {
	if len(ls.buffer) == 0 {
		return nil
	}

	// Group entries by label set
	streamMap := make(map[string][]*LogEntry)

	for _, entry := range ls.buffer {
		// Create label string
		labels := make(map[string]string)
		for k, v := range ls.labels {
			labels[k] = v
		}
		labels["level"] = entry.Level
		labels["logger"] = entry.Logger

		// Add entry labels
		for k, v := range entry.Labels {
			labels[k] = v
		}

		labelStr := createLabelString(labels)
		streamMap[labelStr] = append(streamMap[labelStr], entry)
	}

	// Create Loki push request
	streams := make([]map[string]interface{}, 0)
	for labelStr, entries := range streamMap {
		values := make([][]string, 0)
		for _, entry := range entries {
			timestamp := fmt.Sprintf("%d", entry.Timestamp.UnixNano())
			line := entry.Message

			// Add fields to the line
			if len(entry.Fields) > 0 {
				fieldsJSON, _ := json.Marshal(entry.Fields)
				line += " " + string(fieldsJSON)
			}

			values = append(values, []string{timestamp, line})
		}

		stream := map[string]interface{}{
			"stream": parseLabels(labelStr),
			"values": values,
		}
		streams = append(streams, stream)
	}

	requestBody := map[string]interface{}{
		"streams": streams,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send to Loki
	req, err := http.NewRequestWithContext(ls.ctx, "POST", ls.endpoint+"/loki/api/v1/push", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ls.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Loki request failed with status: %d", resp.StatusCode)
	}

	// Clear buffer on success
	ls.buffer = ls.buffer[:0]

	return nil
}

func (ls *LokiShipper) flushBuffer() error {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	return ls.shipBuffer()
}

func (ls *LokiShipper) Stop() error {
	if ls.cancel != nil {
		ls.cancel()
	}
	ls.wg.Wait()
	return ls.flushBuffer()
}

func (ls *LokiShipper) Name() string {
	return ls.name
}

// NewFluentdShipper creates a new Fluentd shipper
func NewFluentdShipper(name string, config *ShipperConfig) (*FluentdShipper, error) {
	settings := config.Settings

	tag := "gzh-manager"
	if t, ok := settings["tag"].(string); ok {
		tag = t
	}

	bufferSize := 100
	if bs, ok := settings["buffer_size"].(int); ok {
		bufferSize = bs
	}

	timeout := 30 * time.Second
	if t, ok := settings["timeout"].(string); ok {
		if duration, err := time.ParseDuration(t); err == nil {
			timeout = duration
		}
	}

	client := &http.Client{
		Timeout: timeout,
	}

	return &FluentdShipper{
		name:       name,
		config:     config,
		client:     client,
		endpoint:   config.Endpoint,
		tag:        tag,
		buffer:     make([]*LogEntry, 0, bufferSize),
		bufferSize: bufferSize,
	}, nil
}

func (fs *FluentdShipper) Start(ctx context.Context) error {
	fs.ctx, fs.cancel = context.WithCancel(ctx)

	// Start batch shipping routine
	fs.wg.Add(1)
	go fs.batchShippingRoutine()

	return nil
}

func (fs *FluentdShipper) Ship(entries []*LogEntry) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.buffer = append(fs.buffer, entries...)

	// Ship if buffer is full
	if len(fs.buffer) >= fs.bufferSize {
		return fs.shipBuffer()
	}

	return nil
}

func (fs *FluentdShipper) batchShippingRoutine() {
	defer fs.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-fs.ctx.Done():
			fs.flushBuffer()
			return
		case <-ticker.C:
			fs.mutex.Lock()
			if len(fs.buffer) > 0 {
				fs.shipBuffer()
			}
			fs.mutex.Unlock()
		}
	}
}

func (fs *FluentdShipper) shipBuffer() error {
	if len(fs.buffer) == 0 {
		return nil
	}

	// Convert to Fluentd format
	records := make([][]interface{}, 0)
	for _, entry := range fs.buffer {
		record := []interface{}{
			fs.tag,
			entry.Timestamp.Unix(),
			map[string]interface{}{
				"level":   entry.Level,
				"message": entry.Message,
				"logger":  entry.Logger,
				"fields":  entry.Fields,
				"labels":  entry.Labels,
			},
		}
		records = append(records, record)
	}

	jsonData, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("failed to marshal records: %w", err)
	}

	// Send to Fluentd
	req, err := http.NewRequestWithContext(fs.ctx, "POST", fs.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := fs.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Fluentd request failed with status: %d", resp.StatusCode)
	}

	// Clear buffer on success
	fs.buffer = fs.buffer[:0]

	return nil
}

func (fs *FluentdShipper) flushBuffer() error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	return fs.shipBuffer()
}

func (fs *FluentdShipper) Stop() error {
	if fs.cancel != nil {
		fs.cancel()
	}
	fs.wg.Wait()
	return fs.flushBuffer()
}

func (fs *FluentdShipper) Name() string {
	return fs.name
}

// NewHTTPShipper creates a new generic HTTP shipper
func NewHTTPShipper(name string, config *ShipperConfig) (*HTTPShipper, error) {
	settings := config.Settings

	method := "POST"
	if m, ok := settings["method"].(string); ok {
		method = m
	}

	headers := make(map[string]string)
	if h, ok := settings["headers"].(map[string]interface{}); ok {
		for k, v := range h {
			if str, ok := v.(string); ok {
				headers[k] = str
			}
		}
	}

	bufferSize := 100
	if bs, ok := settings["buffer_size"].(int); ok {
		bufferSize = bs
	}

	batchSize := 50
	if bs, ok := settings["batch_size"].(int); ok {
		batchSize = bs
	}

	timeout := 30 * time.Second
	if t, ok := settings["timeout"].(string); ok {
		if duration, err := time.ParseDuration(t); err == nil {
			timeout = duration
		}
	}

	client := &http.Client{
		Timeout: timeout,
	}

	return &HTTPShipper{
		name:       name,
		config:     config,
		client:     client,
		endpoint:   config.Endpoint,
		method:     method,
		headers:    headers,
		buffer:     make([]*LogEntry, 0, bufferSize),
		bufferSize: bufferSize,
		batchSize:  batchSize,
	}, nil
}

func (hs *HTTPShipper) Start(ctx context.Context) error {
	hs.ctx, hs.cancel = context.WithCancel(ctx)

	// Start batch shipping routine
	hs.wg.Add(1)
	go hs.batchShippingRoutine()

	return nil
}

func (hs *HTTPShipper) Ship(entries []*LogEntry) error {
	hs.mutex.Lock()
	defer hs.mutex.Unlock()

	hs.buffer = append(hs.buffer, entries...)

	// Ship if buffer is full
	if len(hs.buffer) >= hs.bufferSize {
		return hs.shipBuffer()
	}

	return nil
}

func (hs *HTTPShipper) batchShippingRoutine() {
	defer hs.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-hs.ctx.Done():
			hs.flushBuffer()
			return
		case <-ticker.C:
			hs.mutex.Lock()
			if len(hs.buffer) > 0 {
				hs.shipBuffer()
			}
			hs.mutex.Unlock()
		}
	}
}

func (hs *HTTPShipper) shipBuffer() error {
	if len(hs.buffer) == 0 {
		return nil
	}

	// Ship in batches
	for i := 0; i < len(hs.buffer); i += hs.batchSize {
		end := i + hs.batchSize
		if end > len(hs.buffer) {
			end = len(hs.buffer)
		}

		batch := hs.buffer[i:end]
		if err := hs.shipBatch(batch); err != nil {
			return err
		}
	}

	// Clear buffer on success
	hs.buffer = hs.buffer[:0]

	return nil
}

func (hs *HTTPShipper) shipBatch(entries []*LogEntry) error {
	requestBody := map[string]interface{}{
		"entries":    entries,
		"timestamp":  time.Now(),
		"source":     "gzh-manager",
		"batch_size": len(entries),
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal batch: %w", err)
	}

	req, err := http.NewRequestWithContext(hs.ctx, hs.method, hs.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range hs.headers {
		req.Header.Set(k, v)
	}

	resp, err := hs.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (hs *HTTPShipper) flushBuffer() error {
	hs.mutex.Lock()
	defer hs.mutex.Unlock()
	return hs.shipBuffer()
}

func (hs *HTTPShipper) Stop() error {
	if hs.cancel != nil {
		hs.cancel()
	}
	hs.wg.Wait()
	return hs.flushBuffer()
}

func (hs *HTTPShipper) Name() string {
	return hs.name
}

// Helper functions

func createLabelString(labels map[string]string) string {
	var parts []string
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, k, v))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func parseLabels(labelStr string) map[string]string {
	labels := make(map[string]string)

	// Simple parser for {key="value",key2="value2"} format
	labelStr = strings.Trim(labelStr, "{}")
	if labelStr == "" {
		return labels
	}

	pairs := strings.Split(labelStr, ",")
	for _, pair := range pairs {
		parts := strings.Split(pair, "=")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
			labels[key] = value
		}
	}

	return labels
}

// Helper function for getting environment
func getShipperEnvironment() string {
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return env
	}
	if env := os.Getenv("ENV"); env != "" {
		return env
	}
	return "development"
}
