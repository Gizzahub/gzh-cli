package feedback

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// FeedbackType represents different types of user feedback
type FeedbackType string

const (
	FeedbackTypeBug           FeedbackType = "bug"
	FeedbackTypeFeature       FeedbackType = "feature"
	FeedbackTypeImprovement   FeedbackType = "improvement"
	FeedbackTypeUsability     FeedbackType = "usability"
	FeedbackTypePerformance   FeedbackType = "performance"
	FeedbackTypeDocumentation FeedbackType = "documentation"
	FeedbackTypeGeneral       FeedbackType = "general"
)

// FeedbackSeverity represents the severity level of feedback
type FeedbackSeverity string

const (
	FeedbackSeverityCritical FeedbackSeverity = "critical"
	FeedbackSeverityHigh     FeedbackSeverity = "high"
	FeedbackSeverityMedium   FeedbackSeverity = "medium"
	FeedbackSeverityLow      FeedbackSeverity = "low"
)

// FeedbackStatus represents the current status of feedback
type FeedbackStatus string

const (
	FeedbackStatusOpen       FeedbackStatus = "open"
	FeedbackStatusInProgress FeedbackStatus = "in_progress"
	FeedbackStatusResolved   FeedbackStatus = "resolved"
	FeedbackStatusClosed     FeedbackStatus = "closed"
	FeedbackStatusDuplicate  FeedbackStatus = "duplicate"
)

// UserFeedback represents a single piece of user feedback
type UserFeedback struct {
	ID          string            `json:"id"`
	Type        FeedbackType      `json:"type"`
	Severity    FeedbackSeverity  `json:"severity"`
	Status      FeedbackStatus    `json:"status"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Component   string            `json:"component,omitempty"`
	Version     string            `json:"version,omitempty"`
	Environment string            `json:"environment,omitempty"`
	UserID      string            `json:"user_id,omitempty"`
	UserEmail   string            `json:"user_email,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	ResolvedAt  *time.Time        `json:"resolved_at,omitempty"`
	AssignedTo  string            `json:"assigned_to,omitempty"`
	Priority    int               `json:"priority"`
	Votes       int               `json:"votes"`
	Comments    []FeedbackComment `json:"comments,omitempty"`
}

// FeedbackComment represents a comment on feedback
type FeedbackComment struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	UserEmail string    `json:"user_email,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	IsPublic  bool      `json:"is_public"`
}

// FeedbackAnalytics represents analytics data for feedback
type FeedbackAnalytics struct {
	TotalFeedback         int                          `json:"total_feedback"`
	FeedbackByType        map[FeedbackType]int         `json:"feedback_by_type"`
	FeedbackBySeverity    map[FeedbackSeverity]int     `json:"feedback_by_severity"`
	FeedbackByStatus      map[FeedbackStatus]int       `json:"feedback_by_status"`
	FeedbackByComponent   map[string]int               `json:"feedback_by_component"`
	AverageResolutionTime time.Duration                `json:"average_resolution_time"`
	TopRequests           []FeedbackSummary            `json:"top_requests"`
	TrendData             map[string][]TimeSeriesPoint `json:"trend_data"`
	UserSatisfaction      float64                      `json:"user_satisfaction"`
	ResponseRate          float64                      `json:"response_rate"`
}

// FeedbackSummary represents a summary of similar feedback items
type FeedbackSummary struct {
	Title       string           `json:"title"`
	Count       int              `json:"count"`
	TotalVotes  int              `json:"total_votes"`
	Type        FeedbackType     `json:"type"`
	Severity    FeedbackSeverity `json:"severity"`
	LastUpdated time.Time        `json:"last_updated"`
}

// TimeSeriesPoint represents a point in time series data
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     int       `json:"value"`
}

// FeedbackFilter represents filters for querying feedback
type FeedbackFilter struct {
	Type          []FeedbackType     `json:"type,omitempty"`
	Severity      []FeedbackSeverity `json:"severity,omitempty"`
	Status        []FeedbackStatus   `json:"status,omitempty"`
	Component     []string           `json:"component,omitempty"`
	AssignedTo    []string           `json:"assigned_to,omitempty"`
	UserID        string             `json:"user_id,omitempty"`
	CreatedAfter  *time.Time         `json:"created_after,omitempty"`
	CreatedBefore *time.Time         `json:"created_before,omitempty"`
	Tags          []string           `json:"tags,omitempty"`
	SearchTerm    string             `json:"search_term,omitempty"`
	MinPriority   int                `json:"min_priority,omitempty"`
	MinVotes      int                `json:"min_votes,omitempty"`
}

// FeedbackManager manages user feedback collection and processing
type FeedbackManager interface {
	// Submit new feedback
	SubmitFeedback(ctx context.Context, feedback *UserFeedback) error

	// Get feedback by ID
	GetFeedback(ctx context.Context, id string) (*UserFeedback, error)

	// List feedback with filters
	ListFeedback(ctx context.Context, filter *FeedbackFilter, limit, offset int) ([]*UserFeedback, int, error)

	// Update feedback status
	UpdateFeedbackStatus(ctx context.Context, id string, status FeedbackStatus, comment string) error

	// Add comment to feedback
	AddComment(ctx context.Context, feedbackID string, comment *FeedbackComment) error

	// Vote on feedback
	VoteFeedback(ctx context.Context, id string, userID string) error

	// Assign feedback to user
	AssignFeedback(ctx context.Context, id string, assigneeID string) error

	// Get feedback analytics
	GetAnalytics(ctx context.Context, filter *FeedbackFilter) (*FeedbackAnalytics, error)

	// Auto-categorize feedback using ML/AI
	AutoCategorizeFeedback(ctx context.Context, feedback *UserFeedback) error

	// Get similar feedback
	FindSimilarFeedback(ctx context.Context, feedback *UserFeedback) ([]*UserFeedback, error)

	// Export feedback data
	ExportFeedback(ctx context.Context, filter *FeedbackFilter, format string) ([]byte, error)

	// Get user satisfaction metrics
	GetUserSatisfactionMetrics(ctx context.Context) (map[string]float64, error)
}

// DefaultFeedbackManager implements FeedbackManager interface
type DefaultFeedbackManager struct {
	mu        sync.RWMutex
	feedback  map[string]*UserFeedback
	userVotes map[string]map[string]bool // feedbackID -> userID -> voted
	analytics *FeedbackAnalytics
	config    *FeedbackConfig
}

// FeedbackConfig represents configuration for feedback manager
type FeedbackConfig struct {
	AutoCategorization     bool          `json:"auto_categorization"`
	EnableVoting           bool          `json:"enable_voting"`
	EnableComments         bool          `json:"enable_comments"`
	MaxCommentsPerFeedback int           `json:"max_comments_per_feedback"`
	AutoAssignment         bool          `json:"auto_assignment"`
	NotificationEnabled    bool          `json:"notification_enabled"`
	AnalyticsEnabled       bool          `json:"analytics_enabled"`
	RetentionPeriod        time.Duration `json:"retention_period"`
}

// NewFeedbackManager creates a new feedback manager
func NewFeedbackManager(config *FeedbackConfig) FeedbackManager {
	if config == nil {
		config = &FeedbackConfig{
			AutoCategorization:     true,
			EnableVoting:           true,
			EnableComments:         true,
			MaxCommentsPerFeedback: 50,
			AutoAssignment:         false,
			NotificationEnabled:    true,
			AnalyticsEnabled:       true,
			RetentionPeriod:        365 * 24 * time.Hour, // 1 year
		}
	}

	return &DefaultFeedbackManager{
		feedback:  make(map[string]*UserFeedback),
		userVotes: make(map[string]map[string]bool),
		analytics: &FeedbackAnalytics{
			FeedbackByType:      make(map[FeedbackType]int),
			FeedbackBySeverity:  make(map[FeedbackSeverity]int),
			FeedbackByStatus:    make(map[FeedbackStatus]int),
			FeedbackByComponent: make(map[string]int),
			TrendData:           make(map[string][]TimeSeriesPoint),
		},
		config: config,
	}
}

// SubmitFeedback submits new user feedback
func (fm *DefaultFeedbackManager) SubmitFeedback(ctx context.Context, feedback *UserFeedback) error {
	if err := fm.validateFeedback(feedback); err != nil {
		return fmt.Errorf("invalid feedback: %w", err)
	}

	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Generate ID if not provided
	if feedback.ID == "" {
		feedback.ID = fmt.Sprintf("feedback_%d", time.Now().UnixNano())
	}

	// Set timestamps
	feedback.CreatedAt = time.Now()
	feedback.UpdatedAt = time.Now()

	// Set initial status
	if feedback.Status == "" {
		feedback.Status = FeedbackStatusOpen
	}

	// Auto-categorize if enabled
	if fm.config.AutoCategorization {
		fm.autoCategorize(feedback)
	}

	// Store feedback
	fm.feedback[feedback.ID] = feedback
	fm.userVotes[feedback.ID] = make(map[string]bool)

	// Update analytics
	fm.updateAnalytics(feedback)

	return nil
}

// GetFeedback retrieves feedback by ID
func (fm *DefaultFeedbackManager) GetFeedback(ctx context.Context, id string) (*UserFeedback, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	feedback, exists := fm.feedback[id]
	if !exists {
		return nil, fmt.Errorf("feedback not found: %s", id)
	}

	// Return a copy to prevent external modifications
	return fm.copyFeedback(feedback), nil
}

// ListFeedback lists feedback with filters
func (fm *DefaultFeedbackManager) ListFeedback(ctx context.Context, filter *FeedbackFilter, limit, offset int) ([]*UserFeedback, int, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	var filtered []*UserFeedback

	for _, feedback := range fm.feedback {
		if fm.matchesFilter(feedback, filter) {
			filtered = append(filtered, fm.copyFeedback(feedback))
		}
	}

	// Sort by priority and creation time
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Priority != filtered[j].Priority {
			return filtered[i].Priority > filtered[j].Priority
		}
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	total := len(filtered)

	// Apply pagination
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

// UpdateFeedbackStatus updates the status of feedback
func (fm *DefaultFeedbackManager) UpdateFeedbackStatus(ctx context.Context, id string, status FeedbackStatus, comment string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	feedback, exists := fm.feedback[id]
	if !exists {
		return fmt.Errorf("feedback not found: %s", id)
	}

	oldStatus := feedback.Status
	feedback.Status = status
	feedback.UpdatedAt = time.Now()

	// Set resolved time
	if status == FeedbackStatusResolved || status == FeedbackStatusClosed {
		now := time.Now()
		feedback.ResolvedAt = &now
	}

	// Add status change comment if provided
	if comment != "" && fm.config.EnableComments {
		statusComment := &FeedbackComment{
			ID:        fmt.Sprintf("comment_%d", time.Now().UnixNano()),
			UserID:    "system",
			Content:   fmt.Sprintf("Status changed from %s to %s: %s", oldStatus, status, comment),
			CreatedAt: time.Now(),
			IsPublic:  true,
		}
		feedback.Comments = append(feedback.Comments, *statusComment)
	}

	// Update analytics
	fm.updateStatusAnalytics(oldStatus, status)

	return nil
}

// AddComment adds a comment to feedback
func (fm *DefaultFeedbackManager) AddComment(ctx context.Context, feedbackID string, comment *FeedbackComment) error {
	if !fm.config.EnableComments {
		return fmt.Errorf("comments are disabled")
	}

	fm.mu.Lock()
	defer fm.mu.Unlock()

	feedback, exists := fm.feedback[feedbackID]
	if !exists {
		return fmt.Errorf("feedback not found: %s", feedbackID)
	}

	if len(feedback.Comments) >= fm.config.MaxCommentsPerFeedback {
		return fmt.Errorf("maximum number of comments reached")
	}

	// Generate ID if not provided
	if comment.ID == "" {
		comment.ID = fmt.Sprintf("comment_%d", time.Now().UnixNano())
	}

	comment.CreatedAt = time.Now()
	feedback.Comments = append(feedback.Comments, *comment)
	feedback.UpdatedAt = time.Now()

	return nil
}

// VoteFeedback allows users to vote on feedback
func (fm *DefaultFeedbackManager) VoteFeedback(ctx context.Context, id string, userID string) error {
	if !fm.config.EnableVoting {
		return fmt.Errorf("voting is disabled")
	}

	fm.mu.Lock()
	defer fm.mu.Unlock()

	feedback, exists := fm.feedback[id]
	if !exists {
		return fmt.Errorf("feedback not found: %s", id)
	}

	votes, exists := fm.userVotes[id]
	if !exists {
		votes = make(map[string]bool)
		fm.userVotes[id] = votes
	}

	// Check if user already voted
	if votes[userID] {
		return fmt.Errorf("user has already voted on this feedback")
	}

	// Record vote
	votes[userID] = true
	feedback.Votes++
	feedback.UpdatedAt = time.Now()

	return nil
}

// AssignFeedback assigns feedback to a user
func (fm *DefaultFeedbackManager) AssignFeedback(ctx context.Context, id string, assigneeID string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	feedback, exists := fm.feedback[id]
	if !exists {
		return fmt.Errorf("feedback not found: %s", id)
	}

	feedback.AssignedTo = assigneeID
	feedback.UpdatedAt = time.Now()

	// Add assignment comment
	if fm.config.EnableComments {
		assignmentComment := &FeedbackComment{
			ID:        fmt.Sprintf("comment_%d", time.Now().UnixNano()),
			UserID:    "system",
			Content:   fmt.Sprintf("Assigned to %s", assigneeID),
			CreatedAt: time.Now(),
			IsPublic:  true,
		}
		feedback.Comments = append(feedback.Comments, *assignmentComment)
	}

	return nil
}

// GetAnalytics returns feedback analytics
func (fm *DefaultFeedbackManager) GetAnalytics(ctx context.Context, filter *FeedbackFilter) (*FeedbackAnalytics, error) {
	if !fm.config.AnalyticsEnabled {
		return nil, fmt.Errorf("analytics are disabled")
	}

	fm.mu.RLock()
	defer fm.mu.RUnlock()

	analytics := &FeedbackAnalytics{
		FeedbackByType:      make(map[FeedbackType]int),
		FeedbackBySeverity:  make(map[FeedbackSeverity]int),
		FeedbackByStatus:    make(map[FeedbackStatus]int),
		FeedbackByComponent: make(map[string]int),
		TrendData:           make(map[string][]TimeSeriesPoint),
	}

	var totalResolutionTime time.Duration
	var resolvedCount int
	var totalSatisfaction float64
	var satisfactionCount int

	for _, feedback := range fm.feedback {
		if filter != nil && !fm.matchesFilter(feedback, filter) {
			continue
		}

		analytics.TotalFeedback++
		analytics.FeedbackByType[feedback.Type]++
		analytics.FeedbackBySeverity[feedback.Severity]++
		analytics.FeedbackByStatus[feedback.Status]++

		if feedback.Component != "" {
			analytics.FeedbackByComponent[feedback.Component]++
		}

		// Calculate resolution time
		if feedback.ResolvedAt != nil {
			resolutionTime := feedback.ResolvedAt.Sub(feedback.CreatedAt)
			totalResolutionTime += resolutionTime
			resolvedCount++
		}

		// Calculate satisfaction (simulated based on votes and status)
		if feedback.Status == FeedbackStatusResolved {
			satisfaction := float64(feedback.Votes) / 10.0 // Simplified calculation
			if satisfaction > 5.0 {
				satisfaction = 5.0
			}
			totalSatisfaction += satisfaction
			satisfactionCount++
		}
	}

	// Calculate averages
	if resolvedCount > 0 {
		analytics.AverageResolutionTime = totalResolutionTime / time.Duration(resolvedCount)
	}

	if satisfactionCount > 0 {
		analytics.UserSatisfaction = totalSatisfaction / float64(satisfactionCount)
	}

	// Calculate response rate (simplified)
	if analytics.TotalFeedback > 0 {
		responseCount := analytics.FeedbackByStatus[FeedbackStatusResolved] + analytics.FeedbackByStatus[FeedbackStatusInProgress]
		analytics.ResponseRate = float64(responseCount) / float64(analytics.TotalFeedback)
	}

	// Generate top requests
	analytics.TopRequests = fm.generateTopRequests(filter)

	return analytics, nil
}

// Helper methods

func (fm *DefaultFeedbackManager) validateFeedback(feedback *UserFeedback) error {
	if feedback == nil {
		return fmt.Errorf("feedback cannot be nil")
	}

	if feedback.Title == "" {
		return fmt.Errorf("title is required")
	}

	if feedback.Description == "" {
		return fmt.Errorf("description is required")
	}

	if feedback.Type == "" {
		feedback.Type = FeedbackTypeGeneral
	}

	if feedback.Severity == "" {
		feedback.Severity = FeedbackSeverityMedium
	}

	return nil
}

func (fm *DefaultFeedbackManager) autoCategorize(feedback *UserFeedback) {
	// Simple auto-categorization based on keywords
	content := feedback.Title + " " + feedback.Description

	if containsAny(content, []string{"bug", "error", "crash", "broken", "fail"}) {
		feedback.Type = FeedbackTypeBug
		feedback.Severity = FeedbackSeverityHigh
	} else if containsAny(content, []string{"slow", "performance", "lag", "timeout"}) {
		feedback.Type = FeedbackTypePerformance
		feedback.Severity = FeedbackSeverityMedium
	} else if containsAny(content, []string{"feature", "add", "new", "want", "need"}) {
		feedback.Type = FeedbackTypeFeature
		feedback.Severity = FeedbackSeverityLow
	} else if containsAny(content, []string{"improve", "better", "enhance", "optimize"}) {
		feedback.Type = FeedbackTypeImprovement
		feedback.Severity = FeedbackSeverityMedium
	} else if containsAny(content, []string{"confusing", "unclear", "help", "how to"}) {
		feedback.Type = FeedbackTypeUsability
		feedback.Severity = FeedbackSeverityMedium
	}

	// Auto-assign priority based on severity
	switch feedback.Severity {
	case FeedbackSeverityCritical:
		feedback.Priority = 100
	case FeedbackSeverityHigh:
		feedback.Priority = 75
	case FeedbackSeverityMedium:
		feedback.Priority = 50
	case FeedbackSeverityLow:
		feedback.Priority = 25
	}
}

func (fm *DefaultFeedbackManager) copyFeedback(feedback *UserFeedback) *UserFeedback {
	copied := *feedback
	if feedback.Tags != nil {
		copied.Tags = make([]string, len(feedback.Tags))
		copy(copied.Tags, feedback.Tags)
	}
	if feedback.Metadata != nil {
		copied.Metadata = make(map[string]string)
		for k, v := range feedback.Metadata {
			copied.Metadata[k] = v
		}
	}
	if feedback.Comments != nil {
		copied.Comments = make([]FeedbackComment, len(feedback.Comments))
		copy(copied.Comments, feedback.Comments)
	}
	return &copied
}

func (fm *DefaultFeedbackManager) matchesFilter(feedback *UserFeedback, filter *FeedbackFilter) bool {
	if filter == nil {
		return true
	}

	// Type filter
	if len(filter.Type) > 0 && !contains(filter.Type, feedback.Type) {
		return false
	}

	// Severity filter
	if len(filter.Severity) > 0 && !containsSeverity(filter.Severity, feedback.Severity) {
		return false
	}

	// Status filter
	if len(filter.Status) > 0 && !containsStatus(filter.Status, feedback.Status) {
		return false
	}

	// Component filter
	if len(filter.Component) > 0 && !containsString(filter.Component, feedback.Component) {
		return false
	}

	// User ID filter
	if filter.UserID != "" && feedback.UserID != filter.UserID {
		return false
	}

	// Date filters
	if filter.CreatedAfter != nil && feedback.CreatedAt.Before(*filter.CreatedAfter) {
		return false
	}

	if filter.CreatedBefore != nil && feedback.CreatedAt.After(*filter.CreatedBefore) {
		return false
	}

	// Priority filter
	if filter.MinPriority > 0 && feedback.Priority < filter.MinPriority {
		return false
	}

	// Votes filter
	if filter.MinVotes > 0 && feedback.Votes < filter.MinVotes {
		return false
	}

	return true
}

func (fm *DefaultFeedbackManager) updateAnalytics(feedback *UserFeedback) {
	fm.analytics.TotalFeedback++
	fm.analytics.FeedbackByType[feedback.Type]++
	fm.analytics.FeedbackBySeverity[feedback.Severity]++
	fm.analytics.FeedbackByStatus[feedback.Status]++

	if feedback.Component != "" {
		fm.analytics.FeedbackByComponent[feedback.Component]++
	}
}

func (fm *DefaultFeedbackManager) updateStatusAnalytics(oldStatus, newStatus FeedbackStatus) {
	fm.analytics.FeedbackByStatus[oldStatus]--
	fm.analytics.FeedbackByStatus[newStatus]++
}

func (fm *DefaultFeedbackManager) generateTopRequests(filter *FeedbackFilter) []FeedbackSummary {
	titleCounts := make(map[string]*FeedbackSummary)

	for _, feedback := range fm.feedback {
		if filter != nil && !fm.matchesFilter(feedback, filter) {
			continue
		}

		if summary, exists := titleCounts[feedback.Title]; exists {
			summary.Count++
			summary.TotalVotes += feedback.Votes
			if feedback.UpdatedAt.After(summary.LastUpdated) {
				summary.LastUpdated = feedback.UpdatedAt
			}
		} else {
			titleCounts[feedback.Title] = &FeedbackSummary{
				Title:       feedback.Title,
				Count:       1,
				TotalVotes:  feedback.Votes,
				Type:        feedback.Type,
				Severity:    feedback.Severity,
				LastUpdated: feedback.UpdatedAt,
			}
		}
	}

	var summaries []FeedbackSummary
	for _, summary := range titleCounts {
		summaries = append(summaries, *summary)
	}

	// Sort by votes and count
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].TotalVotes != summaries[j].TotalVotes {
			return summaries[i].TotalVotes > summaries[j].TotalVotes
		}
		return summaries[i].Count > summaries[j].Count
	})

	// Return top 10
	if len(summaries) > 10 {
		summaries = summaries[:10]
	}

	return summaries
}

// Additional interface methods with placeholder implementations

func (fm *DefaultFeedbackManager) AutoCategorizeFeedback(ctx context.Context, feedback *UserFeedback) error {
	fm.autoCategorize(feedback)
	return nil
}

func (fm *DefaultFeedbackManager) FindSimilarFeedback(ctx context.Context, feedback *UserFeedback) ([]*UserFeedback, error) {
	// Simple similarity based on title and type
	var similar []*UserFeedback

	fm.mu.RLock()
	defer fm.mu.RUnlock()

	for _, existing := range fm.feedback {
		if existing.ID != feedback.ID && existing.Type == feedback.Type {
			if similarity := calculateSimilarity(feedback.Title, existing.Title); similarity > 0.7 {
				similar = append(similar, fm.copyFeedback(existing))
			}
		}
	}

	return similar, nil
}

func (fm *DefaultFeedbackManager) ExportFeedback(ctx context.Context, filter *FeedbackFilter, format string) ([]byte, error) {
	feedbackList, _, err := fm.ListFeedback(ctx, filter, 1000, 0)
	if err != nil {
		return nil, err
	}

	switch format {
	case "json":
		return json.Marshal(feedbackList)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (fm *DefaultFeedbackManager) GetUserSatisfactionMetrics(ctx context.Context) (map[string]float64, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	metrics := make(map[string]float64)

	// Calculate overall satisfaction
	totalSatisfaction := 0.0
	satisfactionCount := 0

	for _, feedback := range fm.feedback {
		if feedback.Status == FeedbackStatusResolved {
			satisfaction := float64(feedback.Votes) / 10.0 // Simplified calculation
			if satisfaction > 5.0 {
				satisfaction = 5.0
			}
			totalSatisfaction += satisfaction
			satisfactionCount++
		}
	}

	if satisfactionCount > 0 {
		metrics["overall_satisfaction"] = totalSatisfaction / float64(satisfactionCount)
	}

	metrics["resolution_rate"] = float64(fm.analytics.FeedbackByStatus[FeedbackStatusResolved]) / float64(fm.analytics.TotalFeedback)
	metrics["response_rate"] = fm.analytics.ResponseRate

	return metrics, nil
}

// Utility functions

func contains(slice []FeedbackType, item FeedbackType) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsSeverity(slice []FeedbackSeverity, item FeedbackSeverity) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsStatus(slice []FeedbackStatus, item FeedbackStatus) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsAny(text string, keywords []string) bool {
	textLower := strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(textLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func calculateSimilarity(text1, text2 string) float64 {
	// Simple similarity calculation based on common words
	words1 := strings.Fields(strings.ToLower(text1))
	words2 := strings.Fields(strings.ToLower(text2))

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	common := 0
	for _, word1 := range words1 {
		for _, word2 := range words2 {
			if word1 == word2 {
				common++
				break
			}
		}
	}

	return float64(common) / float64(len(words1)+len(words2)-common)
}
