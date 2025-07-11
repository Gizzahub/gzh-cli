package monitoring

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// LogIndexer interface for search indexing
type LogIndexer interface {
	Index(entry *LogEntry) error
	Search(query *SearchQuery) (*SearchResult, error)
	CreateIndex(name string, config *IndexConfig) error
	DeleteIndex(name string) error
	GetStats() *IndexStats
	Close() error
}

// SearchQuery represents a search query
type SearchQuery struct {
	Query        string                  `json:"query"`
	Filters      map[string]interface{}  `json:"filters"`
	TimeRange    *SearchTimeRange        `json:"time_range"`
	Fields       []string                `json:"fields"`
	Sort         []SortField             `json:"sort"`
	Limit        int                     `json:"limit"`
	Offset       int                     `json:"offset"`
	Aggregations map[string]*Aggregation `json:"aggregations"`
	Highlight    *HighlightConfig        `json:"highlight"`
}

// SearchResult represents search results
type SearchResult struct {
	Total        int64                         `json:"total"`
	TookMs       int64                         `json:"took_ms"`
	Hits         []*SearchHit                  `json:"hits"`
	Aggregations map[string]*AggregationResult `json:"aggregations,omitempty"`
	Facets       map[string]*FacetResult       `json:"facets,omitempty"`
}

// SearchHit represents a single search result
type SearchHit struct {
	ID        string              `json:"id"`
	Score     float64             `json:"score"`
	Source    *LogEntry           `json:"source"`
	Highlight map[string][]string `json:"highlight,omitempty"`
}

// SearchTimeRange represents a time range filter for search
type SearchTimeRange struct {
	From *time.Time `json:"from,omitempty"`
	To   *time.Time `json:"to,omitempty"`
}

// SortField represents sorting configuration
type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"` // "asc" or "desc"
}

// Aggregation represents aggregation configuration
type Aggregation struct {
	Type     string                  `json:"type"` // "terms", "date_histogram", "range", "stats"
	Field    string                  `json:"field"`
	Size     int                     `json:"size"`
	Interval string                  `json:"interval,omitempty"` // for date_histogram
	Ranges   []AggregationRange      `json:"ranges,omitempty"`   // for range aggregation
	SubAggs  map[string]*Aggregation `json:"sub_aggs,omitempty"`
}

// AggregationRange represents a range for range aggregations
type AggregationRange struct {
	From *float64 `json:"from,omitempty"`
	To   *float64 `json:"to,omitempty"`
	Key  string   `json:"key,omitempty"`
}

// AggregationResult represents aggregation results
type AggregationResult struct {
	Buckets []AggregationBucket     `json:"buckets,omitempty"`
	Stats   *StatsAggregationResult `json:"stats,omitempty"`
}

// AggregationBucket represents a single aggregation bucket
type AggregationBucket struct {
	Key      interface{}                   `json:"key"`
	DocCount int64                         `json:"doc_count"`
	SubAggs  map[string]*AggregationResult `json:"sub_aggs,omitempty"`
}

// StatsAggregationResult represents stats aggregation results
type StatsAggregationResult struct {
	Count int64   `json:"count"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Avg   float64 `json:"avg"`
	Sum   float64 `json:"sum"`
}

// FacetResult represents facet results
type FacetResult struct {
	Terms []FacetTerm `json:"terms"`
	Other int64       `json:"other"`
	Total int64       `json:"total"`
}

// FacetTerm represents a facet term
type FacetTerm struct {
	Term  string `json:"term"`
	Count int64  `json:"count"`
}

// HighlightConfig represents highlight configuration
type HighlightConfig struct {
	Fields   []string `json:"fields"`
	PreTag   string   `json:"pre_tag"`
	PostTag  string   `json:"post_tag"`
	FragSize int      `json:"frag_size"`
	NumFrags int      `json:"num_frags"`
}

// IndexConfig represents index configuration
type IndexConfig struct {
	Name            string            `json:"name"`
	Settings        *IndexSettings    `json:"settings"`
	Mappings        *IndexMappings    `json:"mappings"`
	Aliases         map[string]string `json:"aliases"`
	RetentionPolicy *RetentionPolicy  `json:"retention_policy"`
}

// IndexSettings represents index settings
type IndexSettings struct {
	Shards          int             `json:"shards"`
	Replicas        int             `json:"replicas"`
	RefreshInterval string          `json:"refresh_interval"`
	MaxResultWindow int             `json:"max_result_window"`
	Analysis        *AnalysisConfig `json:"analysis"`
}

// IndexMappings represents field mappings
type IndexMappings struct {
	Properties map[string]*FieldMapping `json:"properties"`
}

// FieldMapping represents field mapping configuration
type FieldMapping struct {
	Type     string                   `json:"type"` // "text", "keyword", "date", "long", "double", "boolean"
	Index    bool                     `json:"index"`
	Store    bool                     `json:"store"`
	Analyzer string                   `json:"analyzer,omitempty"`
	Format   string                   `json:"format,omitempty"`
	Fields   map[string]*FieldMapping `json:"fields,omitempty"`
}

// AnalysisConfig represents text analysis configuration
type AnalysisConfig struct {
	Analyzers  map[string]*AnalyzerConfig    `json:"analyzers"`
	Tokenizers map[string]*TokenizerConfig   `json:"tokenizers"`
	Filters    map[string]*IndexFilterConfig `json:"filters"`
}

// AnalyzerConfig represents analyzer configuration
type AnalyzerConfig struct {
	Type      string   `json:"type"`
	Tokenizer string   `json:"tokenizer"`
	Filters   []string `json:"filters"`
}

// TokenizerConfig represents tokenizer configuration
type TokenizerConfig struct {
	Type    string `json:"type"`
	Pattern string `json:"pattern,omitempty"`
}

// IndexFilterConfig represents text analysis filter configuration (different from log filters)
type IndexFilterConfig struct {
	Type      string   `json:"type"`
	Stopwords []string `json:"stopwords,omitempty"`
}

// RetentionPolicy represents data retention policy
type RetentionPolicy struct {
	MaxAge  string `json:"max_age"`  // "30d", "1y", etc.
	MaxSize string `json:"max_size"` // "10GB", "1TB", etc.
	MaxDocs int64  `json:"max_docs"`
}

// IndexStats represents index statistics
type IndexStats struct {
	Name        string    `json:"name"`
	DocCount    int64     `json:"doc_count"`
	Size        int64     `json:"size_bytes"`
	CreatedAt   time.Time `json:"created_at"`
	LastUpdated time.Time `json:"last_updated"`
	Health      string    `json:"health"`
}

// MemoryIndexer implements in-memory search indexing for development/testing
type MemoryIndexer struct {
	name      string
	documents map[string]*IndexedLogEntry
	inverted  map[string]map[string]float64 // term -> docID -> score
	fields    map[string]map[string]bool    // field -> value -> exists
	stats     *IndexStats
	mutex     sync.RWMutex
	config    *IndexConfig
}

// IndexedLogEntry represents an indexed log entry
type IndexedLogEntry struct {
	ID       string                 `json:"id"`
	Entry    *LogEntry              `json:"entry"`
	Indexed  time.Time              `json:"indexed"`
	Terms    []string               `json:"terms"`
	FieldMap map[string]interface{} `json:"field_map"`
}

// NewMemoryIndexer creates a new memory-based indexer
func NewMemoryIndexer(name string, config *IndexConfig) *MemoryIndexer {
	return &MemoryIndexer{
		name:      name,
		documents: make(map[string]*IndexedLogEntry),
		inverted:  make(map[string]map[string]float64),
		fields:    make(map[string]map[string]bool),
		config:    config,
		stats: &IndexStats{
			Name:      name,
			CreatedAt: time.Now(),
			Health:    "green",
		},
	}
}

func (mi *MemoryIndexer) Index(entry *LogEntry) error {
	mi.mutex.Lock()
	defer mi.mutex.Unlock()

	// Generate document ID
	docID := fmt.Sprintf("%d_%s_%d",
		entry.Timestamp.UnixNano(),
		entry.Logger,
		len(mi.documents))

	// Create indexed entry
	indexed := &IndexedLogEntry{
		ID:       docID,
		Entry:    entry,
		Indexed:  time.Now(),
		FieldMap: make(map[string]interface{}),
	}

	// Index full-text content
	content := mi.buildIndexableContent(entry)
	terms := mi.tokenize(content)
	indexed.Terms = terms

	// Index individual fields
	mi.indexFields(indexed, entry)

	// Add to inverted index
	for _, term := range terms {
		if mi.inverted[term] == nil {
			mi.inverted[term] = make(map[string]float64)
		}
		mi.inverted[term][docID] = mi.calculateTFIDF(term, terms, docID)
	}

	// Store document
	mi.documents[docID] = indexed

	// Update statistics
	mi.stats.DocCount++
	mi.stats.LastUpdated = time.Now()

	return nil
}

func (mi *MemoryIndexer) buildIndexableContent(entry *LogEntry) string {
	var parts []string

	parts = append(parts, entry.Message)
	parts = append(parts, entry.Logger)
	parts = append(parts, entry.Level)

	// Add field values
	for _, v := range entry.Fields {
		if str, ok := v.(string); ok {
			parts = append(parts, str)
		}
	}

	// Add label values
	for _, v := range entry.Labels {
		parts = append(parts, v)
	}

	return strings.Join(parts, " ")
}

func (mi *MemoryIndexer) tokenize(text string) []string {
	// Simple tokenization - split on whitespace and punctuation
	re := regexp.MustCompile(`[^\w]+`)
	tokens := re.Split(strings.ToLower(text), -1)

	var result []string
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if len(token) > 0 && len(token) > 1 { // Skip single characters
			result = append(result, token)
		}
	}

	return result
}

func (mi *MemoryIndexer) indexFields(indexed *IndexedLogEntry, entry *LogEntry) {
	// Index structured fields
	indexed.FieldMap["timestamp"] = entry.Timestamp
	indexed.FieldMap["level"] = entry.Level
	indexed.FieldMap["logger"] = entry.Logger
	indexed.FieldMap["message"] = entry.Message

	if entry.TraceID != "" {
		indexed.FieldMap["trace_id"] = entry.TraceID
	}
	if entry.SpanID != "" {
		indexed.FieldMap["span_id"] = entry.SpanID
	}

	// Index custom fields
	for k, v := range entry.Fields {
		indexed.FieldMap[k] = v
		mi.addFieldValue(k, v)
	}

	// Index labels
	for k, v := range entry.Labels {
		labelKey := "label_" + k
		indexed.FieldMap[labelKey] = v
		mi.addFieldValue(labelKey, v)
	}
}

func (mi *MemoryIndexer) addFieldValue(field string, value interface{}) {
	if mi.fields[field] == nil {
		mi.fields[field] = make(map[string]bool)
	}

	valueStr := fmt.Sprintf("%v", value)
	mi.fields[field][valueStr] = true
}

func (mi *MemoryIndexer) calculateTFIDF(term string, docTerms []string, docID string) float64 {
	// Simple TF-IDF calculation
	tf := mi.termFrequency(term, docTerms)
	idf := mi.inverseDocumentFrequency(term)
	return tf * idf
}

func (mi *MemoryIndexer) termFrequency(term string, docTerms []string) float64 {
	count := 0
	for _, t := range docTerms {
		if t == term {
			count++
		}
	}
	return float64(count) / float64(len(docTerms))
}

func (mi *MemoryIndexer) inverseDocumentFrequency(term string) float64 {
	docCount := len(mi.documents)
	if docCount == 0 {
		return 0
	}

	termDocs := len(mi.inverted[term])
	if termDocs == 0 {
		return 0
	}

	return 1.0 + float64(docCount)/float64(termDocs)
}

func (mi *MemoryIndexer) Search(query *SearchQuery) (*SearchResult, error) {
	mi.mutex.RLock()
	defer mi.mutex.RUnlock()

	start := time.Now()

	// Find matching documents
	candidates := mi.findCandidates(query)

	// Apply filters
	filtered := mi.applyFilters(candidates, query)

	// Apply time range filter
	if query.TimeRange != nil {
		filtered = mi.applyTimeRange(filtered, query.TimeRange)
	}

	// Score and sort results
	scored := mi.scoreResults(filtered, query)

	// Apply pagination
	total := int64(len(scored))
	offset := query.Offset
	limit := query.Limit
	if limit == 0 {
		limit = 50 // Default limit
	}

	end := offset + limit
	if end > len(scored) {
		end = len(scored)
	}

	if offset > len(scored) {
		offset = len(scored)
	}

	paginatedResults := scored[offset:end]

	// Create search hits
	hits := make([]*SearchHit, len(paginatedResults))
	for i, doc := range paginatedResults {
		hits[i] = &SearchHit{
			ID:     doc.ID,
			Score:  mi.getDocumentScore(doc.ID, query),
			Source: doc.Entry,
		}

		// Add highlighting if requested
		if query.Highlight != nil {
			hits[i].Highlight = mi.generateHighlights(doc, query)
		}
	}

	result := &SearchResult{
		Total:  total,
		TookMs: time.Since(start).Milliseconds(),
		Hits:   hits,
	}

	// Process aggregations if requested
	if len(query.Aggregations) > 0 {
		result.Aggregations = mi.processAggregations(filtered, query.Aggregations)
	}

	return result, nil
}

func (mi *MemoryIndexer) findCandidates(query *SearchQuery) []*IndexedLogEntry {
	if query.Query == "" {
		// Return all documents if no query
		result := make([]*IndexedLogEntry, 0, len(mi.documents))
		for _, doc := range mi.documents {
			result = append(result, doc)
		}
		return result
	}

	// Tokenize query
	queryTerms := mi.tokenize(query.Query)
	if len(queryTerms) == 0 {
		return nil
	}

	// Find documents containing query terms
	candidateScores := make(map[string]float64)

	for _, term := range queryTerms {
		if termDocs, exists := mi.inverted[term]; exists {
			for docID, score := range termDocs {
				candidateScores[docID] += score
			}
		}
	}

	// Convert to document list
	var candidates []*IndexedLogEntry
	for docID := range candidateScores {
		if doc, exists := mi.documents[docID]; exists {
			candidates = append(candidates, doc)
		}
	}

	return candidates
}

func (mi *MemoryIndexer) applyFilters(docs []*IndexedLogEntry, query *SearchQuery) []*IndexedLogEntry {
	if len(query.Filters) == 0 {
		return docs
	}

	var filtered []*IndexedLogEntry

	for _, doc := range docs {
		match := true

		for field, filterValue := range query.Filters {
			docValue, exists := doc.FieldMap[field]
			if !exists {
				match = false
				break
			}

			// Simple equality filter for now
			if fmt.Sprintf("%v", docValue) != fmt.Sprintf("%v", filterValue) {
				match = false
				break
			}
		}

		if match {
			filtered = append(filtered, doc)
		}
	}

	return filtered
}

func (mi *MemoryIndexer) applyTimeRange(docs []*IndexedLogEntry, timeRange *SearchTimeRange) []*IndexedLogEntry {
	var filtered []*IndexedLogEntry

	for _, doc := range docs {
		timestamp := doc.Entry.Timestamp

		if timeRange.From != nil && timestamp.Before(*timeRange.From) {
			continue
		}

		if timeRange.To != nil && timestamp.After(*timeRange.To) {
			continue
		}

		filtered = append(filtered, doc)
	}

	return filtered
}

func (mi *MemoryIndexer) scoreResults(docs []*IndexedLogEntry, query *SearchQuery) []*IndexedLogEntry {
	// Sort by timestamp descending by default
	sort.Slice(docs, func(i, j int) bool {
		if len(query.Sort) > 0 {
			for _, sortField := range query.Sort {
				valueI := mi.getFieldValue(docs[i], sortField.Field)
				valueJ := mi.getFieldValue(docs[j], sortField.Field)

				if sortField.Order == "asc" {
					return mi.compareValues(valueI, valueJ) < 0
				} else {
					return mi.compareValues(valueI, valueJ) > 0
				}
			}
		}

		// Default: newest first
		return docs[i].Entry.Timestamp.After(docs[j].Entry.Timestamp)
	})

	return docs
}

func (mi *MemoryIndexer) getFieldValue(doc *IndexedLogEntry, field string) interface{} {
	if value, exists := doc.FieldMap[field]; exists {
		return value
	}
	return nil
}

func (mi *MemoryIndexer) compareValues(a, b interface{}) int {
	// Simple comparison - could be enhanced
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	if aStr < bStr {
		return -1
	} else if aStr > bStr {
		return 1
	}
	return 0
}

func (mi *MemoryIndexer) getDocumentScore(docID string, query *SearchQuery) float64 {
	if query.Query == "" {
		return 1.0
	}

	queryTerms := mi.tokenize(query.Query)
	score := 0.0

	for _, term := range queryTerms {
		if termDocs, exists := mi.inverted[term]; exists {
			if docScore, exists := termDocs[docID]; exists {
				score += docScore
			}
		}
	}

	return score
}

func (mi *MemoryIndexer) generateHighlights(doc *IndexedLogEntry, query *SearchQuery) map[string][]string {
	highlights := make(map[string][]string)

	if query.Highlight == nil || query.Query == "" {
		return highlights
	}

	queryTerms := mi.tokenize(query.Query)
	preTag := query.Highlight.PreTag
	postTag := query.Highlight.PostTag

	if preTag == "" {
		preTag = "<em>"
	}
	if postTag == "" {
		postTag = "</em>"
	}

	// Highlight message field
	if containsString(query.Highlight.Fields, "message") || len(query.Highlight.Fields) == 0 {
		highlighted := mi.highlightText(doc.Entry.Message, queryTerms, preTag, postTag)
		if highlighted != doc.Entry.Message {
			highlights["message"] = []string{highlighted}
		}
	}

	return highlights
}

func (mi *MemoryIndexer) highlightText(text string, terms []string, preTag, postTag string) string {
	result := text

	for _, term := range terms {
		pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(term) + `\b`)
		result = pattern.ReplaceAllString(result, preTag+"$0"+postTag)
	}

	return result
}

func (mi *MemoryIndexer) processAggregations(docs []*IndexedLogEntry, aggs map[string]*Aggregation) map[string]*AggregationResult {
	results := make(map[string]*AggregationResult)

	for name, agg := range aggs {
		switch agg.Type {
		case "terms":
			results[name] = mi.processTermsAggregation(docs, agg)
		case "date_histogram":
			results[name] = mi.processDateHistogramAggregation(docs, agg)
		case "stats":
			results[name] = mi.processStatsAggregation(docs, agg)
		}
	}

	return results
}

func (mi *MemoryIndexer) processTermsAggregation(docs []*IndexedLogEntry, agg *Aggregation) *AggregationResult {
	counts := make(map[string]int64)

	for _, doc := range docs {
		value := mi.getFieldValue(doc, agg.Field)
		if value != nil {
			key := fmt.Sprintf("%v", value)
			counts[key]++
		}
	}

	// Convert to buckets and sort by count
	var buckets []AggregationBucket
	for key, count := range counts {
		buckets = append(buckets, AggregationBucket{
			Key:      key,
			DocCount: count,
		})
	}

	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].DocCount > buckets[j].DocCount
	})

	// Limit results
	size := agg.Size
	if size == 0 {
		size = 10
	}
	if len(buckets) > size {
		buckets = buckets[:size]
	}

	return &AggregationResult{Buckets: buckets}
}

func (mi *MemoryIndexer) processDateHistogramAggregation(docs []*IndexedLogEntry, agg *Aggregation) *AggregationResult {
	// Simple date histogram - group by hour
	counts := make(map[string]int64)

	for _, doc := range docs {
		timestamp := doc.Entry.Timestamp
		// Round to hour
		hour := timestamp.Truncate(time.Hour)
		key := hour.Format(time.RFC3339)
		counts[key]++
	}

	var buckets []AggregationBucket
	for key, count := range counts {
		buckets = append(buckets, AggregationBucket{
			Key:      key,
			DocCount: count,
		})
	}

	// Sort by time
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Key.(string) < buckets[j].Key.(string)
	})

	return &AggregationResult{Buckets: buckets}
}

func (mi *MemoryIndexer) processStatsAggregation(docs []*IndexedLogEntry, agg *Aggregation) *AggregationResult {
	var values []float64

	for _, doc := range docs {
		value := mi.getFieldValue(doc, agg.Field)
		if value != nil {
			if floatVal, ok := value.(float64); ok {
				values = append(values, floatVal)
			} else if intVal, ok := value.(int); ok {
				values = append(values, float64(intVal))
			}
		}
	}

	if len(values) == 0 {
		return &AggregationResult{
			Stats: &StatsAggregationResult{Count: 0},
		}
	}

	// Calculate stats
	var sum, min, max float64
	min = values[0]
	max = values[0]

	for _, v := range values {
		sum += v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	avg := sum / float64(len(values))

	return &AggregationResult{
		Stats: &StatsAggregationResult{
			Count: int64(len(values)),
			Min:   min,
			Max:   max,
			Avg:   avg,
			Sum:   sum,
		},
	}
}

func (mi *MemoryIndexer) CreateIndex(name string, config *IndexConfig) error {
	mi.mutex.Lock()
	defer mi.mutex.Unlock()

	mi.config = config
	mi.stats.Name = name

	return nil
}

func (mi *MemoryIndexer) DeleteIndex(name string) error {
	mi.mutex.Lock()
	defer mi.mutex.Unlock()

	// Clear all data
	mi.documents = make(map[string]*IndexedLogEntry)
	mi.inverted = make(map[string]map[string]float64)
	mi.fields = make(map[string]map[string]bool)
	mi.stats.DocCount = 0

	return nil
}

func (mi *MemoryIndexer) GetStats() *IndexStats {
	mi.mutex.RLock()
	defer mi.mutex.RUnlock()

	// Calculate approximate size
	var size int64
	for _, doc := range mi.documents {
		data, _ := json.Marshal(doc)
		size += int64(len(data))
	}

	stats := *mi.stats
	stats.Size = size

	return &stats
}

func (mi *MemoryIndexer) Close() error {
	mi.mutex.Lock()
	defer mi.mutex.Unlock()

	// Clear all data to free memory
	mi.documents = nil
	mi.inverted = nil
	mi.fields = nil

	return nil
}

// Helper functions

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
