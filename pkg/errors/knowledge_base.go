package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// KnowledgeBase represents an external knowledge source
type KnowledgeBase interface {
	Search(ctx context.Context, query string) ([]KnowledgeArticle, error)
	GetArticle(ctx context.Context, id string) (*KnowledgeArticle, error)
	GetInfo() KnowledgeBaseInfo
}

// KnowledgeBaseInfo describes a knowledge base
type KnowledgeBaseInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	BaseURL     string `json:"base_url"`
	Priority    int    `json:"priority"`
	Enabled     bool   `json:"enabled"`
}

// KnowledgeArticle represents a knowledge base article
type KnowledgeArticle struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	Summary     string            `json:"summary"`
	URL         string            `json:"url"`
	Tags        []string          `json:"tags"`
	Category    string            `json:"category"`
	Score       float64           `json:"score"` // Relevance score
	LastUpdated time.Time         `json:"last_updated"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// FAQEntry represents a frequently asked question
type FAQEntry struct {
	ID       string   `json:"id"`
	Question string   `json:"question"`
	Answer   string   `json:"answer"`
	Tags     []string `json:"tags"`
	Category string   `json:"category"`
	Votes    int      `json:"votes"`
	Views    int      `json:"views"`
}

// GitHubKnowledgeBase implements KnowledgeBase for GitHub issues/discussions
type GitHubKnowledgeBase struct {
	client   *http.Client
	baseURL  string
	token    string
	repo     string
	priority int
}

// LocalKnowledgeBase implements KnowledgeBase for local markdown files
type LocalKnowledgeBase struct {
	basePath string
	articles map[string]KnowledgeArticle
	mu       sync.RWMutex
	priority int
}

// FAQDatabase manages frequently asked questions
type FAQDatabase struct {
	entries   map[string][]FAQEntry // Keyed by category
	searchIdx map[string][]string   // Keyword to FAQ IDs mapping
	mu        sync.RWMutex
}

// KnowledgeManager coordinates multiple knowledge sources
type KnowledgeManager struct {
	knowledgeBases []KnowledgeBase
	faqDB          *FAQDatabase
	cache          *KnowledgeCache
	mu             sync.RWMutex
}

// KnowledgeCache caches search results
type KnowledgeCache struct {
	articles map[string]CacheEntry
	mu       sync.RWMutex
	ttl      time.Duration
}

// CacheEntry represents a cached knowledge article
type CacheEntry struct {
	Article   KnowledgeArticle `json:"article"`
	Timestamp time.Time        `json:"timestamp"`
}

// NewGitHubKnowledgeBase creates a GitHub-based knowledge base
func NewGitHubKnowledgeBase(repo, token string) *GitHubKnowledgeBase {
	return &GitHubKnowledgeBase{
		client:   &http.Client{Timeout: 30 * time.Second},
		baseURL:  "https://api.github.com",
		token:    token,
		repo:     repo,
		priority: 8,
	}
}

// NewLocalKnowledgeBase creates a local file-based knowledge base
func NewLocalKnowledgeBase(basePath string) *LocalKnowledgeBase {
	return &LocalKnowledgeBase{
		basePath: basePath,
		articles: make(map[string]KnowledgeArticle),
		priority: 5,
	}
}

// NewFAQDatabase creates a new FAQ database
func NewFAQDatabase() *FAQDatabase {
	return &FAQDatabase{
		entries:   make(map[string][]FAQEntry),
		searchIdx: make(map[string][]string),
	}
}

// NewKnowledgeManager creates a new knowledge manager
func NewKnowledgeManager() *KnowledgeManager {
	return &KnowledgeManager{
		knowledgeBases: make([]KnowledgeBase, 0),
		faqDB:          NewFAQDatabase(),
		cache: &KnowledgeCache{
			articles: make(map[string]CacheEntry),
			ttl:      30 * time.Minute,
		},
	}
}

// GitHub Knowledge Base Implementation

func (kb *GitHubKnowledgeBase) Search(ctx context.Context, query string) ([]KnowledgeArticle, error) {
	searchURL := fmt.Sprintf("%s/search/issues?q=%s+repo:%s",
		kb.baseURL,
		url.QueryEscape(query),
		kb.repo)

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	if kb.token != "" {
		req.Header.Set("Authorization", "token "+kb.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := kb.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var searchResult struct {
		Items []struct {
			ID     int    `json:"id"`
			Title  string `json:"title"`
			Body   string `json:"body"`
			URL    string `json:"html_url"`
			State  string `json:"state"`
			Labels []struct {
				Name string `json:"name"`
			} `json:"labels"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, err
	}

	var articles []KnowledgeArticle
	for _, item := range searchResult.Items {
		tags := make([]string, len(item.Labels))
		for i, label := range item.Labels {
			tags[i] = label.Name
		}

		article := KnowledgeArticle{
			ID:          fmt.Sprintf("github-%d", item.ID),
			Title:       item.Title,
			Content:     item.Body,
			Summary:     truncateString(item.Body, 200),
			URL:         item.URL,
			Tags:        tags,
			Category:    "github-issue",
			Score:       calculateRelevanceScore(item.Title, item.Body, query),
			LastUpdated: item.UpdatedAt,
			Metadata: map[string]string{
				"state": item.State,
				"repo":  kb.repo,
			},
		}
		articles = append(articles, article)
	}

	return articles, nil
}

func (kb *GitHubKnowledgeBase) GetArticle(ctx context.Context, id string) (*KnowledgeArticle, error) {
	// Extract issue number from ID
	var issueNum string
	if strings.HasPrefix(id, "github-") {
		issueNum = strings.TrimPrefix(id, "github-")
	} else {
		issueNum = id
	}

	issueURL := fmt.Sprintf("%s/repos/%s/issues/%s", kb.baseURL, kb.repo, issueNum)

	req, err := http.NewRequestWithContext(ctx, "GET", issueURL, nil)
	if err != nil {
		return nil, err
	}

	if kb.token != "" {
		req.Header.Set("Authorization", "token "+kb.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := kb.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issue struct {
		ID     int    `json:"id"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		URL    string `json:"html_url"`
		State  string `json:"state"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	tags := make([]string, len(issue.Labels))
	for i, label := range issue.Labels {
		tags[i] = label.Name
	}

	article := &KnowledgeArticle{
		ID:          fmt.Sprintf("github-%d", issue.ID),
		Title:       issue.Title,
		Content:     issue.Body,
		Summary:     truncateString(issue.Body, 200),
		URL:         issue.URL,
		Tags:        tags,
		Category:    "github-issue",
		LastUpdated: issue.UpdatedAt,
		Metadata: map[string]string{
			"state": issue.State,
			"repo":  kb.repo,
		},
	}

	return article, nil
}

func (kb *GitHubKnowledgeBase) GetInfo() KnowledgeBaseInfo {
	return KnowledgeBaseInfo{
		Name:        "GitHub Issues",
		Description: fmt.Sprintf("GitHub issues and discussions from %s", kb.repo),
		BaseURL:     fmt.Sprintf("https://github.com/%s", kb.repo),
		Priority:    kb.priority,
		Enabled:     true,
	}
}

// Local Knowledge Base Implementation

func (kb *LocalKnowledgeBase) Search(ctx context.Context, query string) ([]KnowledgeArticle, error) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	var results []KnowledgeArticle
	queryLower := strings.ToLower(query)

	for _, article := range kb.articles {
		score := calculateRelevanceScore(article.Title, article.Content, queryLower)
		if score > 0.1 { // Minimum relevance threshold
			article.Score = score
			results = append(results, article)
		}
	}

	// Sort by relevance score
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Score < results[j].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results, nil
}

func (kb *LocalKnowledgeBase) GetArticle(ctx context.Context, id string) (*KnowledgeArticle, error) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	if article, exists := kb.articles[id]; exists {
		return &article, nil
	}

	return nil, fmt.Errorf("article %s not found", id)
}

func (kb *LocalKnowledgeBase) GetInfo() KnowledgeBaseInfo {
	return KnowledgeBaseInfo{
		Name:        "Local Knowledge Base",
		Description: fmt.Sprintf("Local markdown files from %s", kb.basePath),
		BaseURL:     kb.basePath,
		Priority:    kb.priority,
		Enabled:     true,
	}
}

// FAQ Database Implementation

func (db *FAQDatabase) AddFAQ(entry FAQEntry) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.entries[entry.Category] == nil {
		db.entries[entry.Category] = make([]FAQEntry, 0)
	}

	db.entries[entry.Category] = append(db.entries[entry.Category], entry)

	// Update search index
	keywords := extractKeywords(entry.Question + " " + entry.Answer)
	for _, keyword := range keywords {
		db.searchIdx[keyword] = append(db.searchIdx[keyword], entry.ID)
	}
}

func (db *FAQDatabase) SearchFAQ(query string) []FAQEntry {
	db.mu.RLock()
	defer db.mu.RUnlock()

	keywords := extractKeywords(query)
	candidateIDs := make(map[string]int)

	// Find candidate FAQs based on keyword matches
	for _, keyword := range keywords {
		if ids, exists := db.searchIdx[keyword]; exists {
			for _, id := range ids {
				candidateIDs[id]++
			}
		}
	}

	var results []FAQEntry
	for _, entries := range db.entries {
		for _, entry := range entries {
			if score, exists := candidateIDs[entry.ID]; exists {
				if score > 0 {
					results = append(results, entry)
				}
			}
		}
	}

	return results
}

func (db *FAQDatabase) GetFAQsByCategory(category string) []FAQEntry {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if entries, exists := db.entries[category]; exists {
		return entries
	}

	return nil
}

// Knowledge Manager Implementation

func (km *KnowledgeManager) RegisterKnowledgeBase(kb KnowledgeBase) {
	km.mu.Lock()
	defer km.mu.Unlock()

	km.knowledgeBases = append(km.knowledgeBases, kb)
}

func (km *KnowledgeManager) SearchKnowledge(ctx context.Context, query string) ([]KnowledgeArticle, error) {
	// Check cache first
	if cached := km.cache.Get(query); len(cached) > 0 {
		return cached, nil
	}

	var allArticles []KnowledgeArticle

	km.mu.RLock()
	knowledgeBases := km.knowledgeBases
	km.mu.RUnlock()

	// Search all knowledge bases concurrently
	type searchResult struct {
		articles []KnowledgeArticle
		err      error
	}

	resultChan := make(chan searchResult, len(knowledgeBases))

	for _, kb := range knowledgeBases {
		go func(kb KnowledgeBase) {
			articles, err := kb.Search(ctx, query)
			resultChan <- searchResult{articles: articles, err: err}
		}(kb)
	}

	// Collect results
	for range knowledgeBases {
		result := <-resultChan
		if result.err == nil {
			allArticles = append(allArticles, result.articles...)
		}
	}

	// Sort by relevance and priority
	for i := 0; i < len(allArticles)-1; i++ {
		for j := i + 1; j < len(allArticles); j++ {
			if allArticles[i].Score < allArticles[j].Score {
				allArticles[i], allArticles[j] = allArticles[j], allArticles[i]
			}
		}
	}

	// Cache results
	km.cache.Set(query, allArticles)

	return allArticles, nil
}

func (km *KnowledgeManager) SearchFAQ(query string) []FAQEntry {
	return km.faqDB.SearchFAQ(query)
}

func (km *KnowledgeManager) GetKnowledgeBases() []KnowledgeBaseInfo {
	km.mu.RLock()
	defer km.mu.RUnlock()

	var infos []KnowledgeBaseInfo
	for _, kb := range km.knowledgeBases {
		infos = append(infos, kb.GetInfo())
	}

	return infos
}

// Cache Implementation

func (cache *KnowledgeCache) Get(query string) []KnowledgeArticle {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	if entry, exists := cache.articles[query]; exists {
		if time.Since(entry.Timestamp) < cache.ttl {
			return []KnowledgeArticle{entry.Article}
		}
		// Remove expired entry
		delete(cache.articles, query)
	}

	return nil
}

func (cache *KnowledgeCache) Set(query string, articles []KnowledgeArticle) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Store first (most relevant) article
	if len(articles) > 0 {
		cache.articles[query] = CacheEntry{
			Article:   articles[0],
			Timestamp: time.Now(),
		}
	}
}

// Helper Functions

func calculateRelevanceScore(title, content, query string) float64 {
	queryLower := strings.ToLower(query)
	titleLower := strings.ToLower(title)
	contentLower := strings.ToLower(content)

	score := 0.0

	// Title matches are weighted higher
	if strings.Contains(titleLower, queryLower) {
		score += 0.8
	}

	// Content matches
	if strings.Contains(contentLower, queryLower) {
		score += 0.3
	}

	// Word-by-word matching
	queryWords := strings.Fields(queryLower)
	titleWords := strings.Fields(titleLower)
	contentWords := strings.Fields(contentLower)

	for _, queryWord := range queryWords {
		for _, titleWord := range titleWords {
			if strings.Contains(titleWord, queryWord) {
				score += 0.2
			}
		}
		for _, contentWord := range contentWords {
			if strings.Contains(contentWord, queryWord) {
				score += 0.1
			}
		}
	}

	return score
}

func extractKeywords(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	var keywords []string

	// Filter out common stop words
	stopWords := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
		"be": true, "by": true, "for": true, "from": true, "has": true, "he": true,
		"in": true, "is": true, "it": true, "its": true, "of": true, "on": true,
		"that": true, "the": true, "to": true, "was": true, "will": true, "with": true,
	}

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:()")
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Initialize default FAQ entries
func (km *KnowledgeManager) InitializeDefaultFAQs() {
	defaultFAQs := []FAQEntry{
		{
			ID:       "github-token-setup",
			Question: "How do I set up a GitHub token for authentication?",
			Answer: `To set up a GitHub token:
1. Go to GitHub Settings > Developer settings > Personal access tokens
2. Generate a new token with 'repo' and 'admin:org' permissions
3. Set the GITHUB_TOKEN environment variable: export GITHUB_TOKEN=your_token_here
4. Test with: gz bulk-clone --dry-run`,
			Tags:     []string{"github", "authentication", "token", "setup"},
			Category: "authentication",
			Votes:    15,
			Views:    234,
		},
		{
			ID:       "network-timeout-fix",
			Question: "What should I do when I get network timeout errors?",
			Answer: `Network timeout solutions:
1. Check your internet connection
2. Use --timeout flag to increase timeout: gz bulk-clone --timeout 300s
3. Check DNS settings (try 8.8.8.8)
4. Verify firewall isn't blocking connections
5. Try again during off-peak hours`,
			Tags:     []string{"network", "timeout", "connectivity", "troubleshooting"},
			Category: "network",
			Votes:    12,
			Views:    189,
		},
		{
			ID:       "config-validation-error",
			Question: "Why am I getting configuration validation errors?",
			Answer: `Configuration validation errors usually mean:
1. Invalid YAML/JSON syntax - check indentation and commas
2. Missing required fields - see documentation for required fields
3. Incorrect field values - check the configuration schema
4. Use 'gz config validate' to check your configuration
5. Refer to examples in the samples/ directory`,
			Tags:     []string{"configuration", "validation", "yaml", "syntax"},
			Category: "configuration",
			Votes:    18,
			Views:    156,
		},
		{
			ID:       "large-organization-performance",
			Question: "How can I improve performance when cloning large organizations?",
			Answer: `For large organizations:
1. Use --workers flag to increase concurrency: gz bulk-clone --workers 10
2. Enable caching: gz bulk-clone --cache
3. Use filters to limit repositories: gz bulk-clone --filter "language:go"
4. Consider using --shallow to clone with limited history
5. Monitor memory usage and adjust --batch-size accordingly`,
			Tags:     []string{"performance", "optimization", "large-scale", "memory"},
			Category: "performance",
			Votes:    25,
			Views:    298,
		},
	}

	for _, faq := range defaultFAQs {
		km.faqDB.AddFAQ(faq)
	}
}
