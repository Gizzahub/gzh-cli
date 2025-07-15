package template

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// SearchCmd represents the search command
var SearchCmd = &cobra.Command{
	Use:   "search",
	Short: "템플릿 검색",
	Long: `마켓플레이스에서 템플릿을 검색합니다.

검색 기능:
- 이름 및 설명 기반 텍스트 검색
- 카테고리별 필터링
- 타입별 필터링
- 작성자별 필터링
- 키워드 기반 검색
- 인기도 및 최신순 정렬

Examples:
  gz template search docker
  gz template search --category web --type helm
  gz template search --author myauthor --sort downloads
  gz template search nginx --limit 10`,
	Run: runSearch,
}

var (
	searchQuery    string
	searchCategory string
	searchType     string
	searchAuthor   string
	searchKeywords []string
	sortBy         string
	searchLimit    int
	showDetails    bool
	outputFormat   string
	searchServer   string
	searchAPIKey   string
	searchConfig   string
	useLocalIndex  bool
)

func init() {
	SearchCmd.Flags().StringVarP(&searchCategory, "category", "c", "", "카테고리로 필터링")
	SearchCmd.Flags().StringVarP(&searchType, "type", "t", "", "타입으로 필터링")
	SearchCmd.Flags().StringVarP(&searchAuthor, "author", "a", "", "작성자로 필터링")
	SearchCmd.Flags().StringSliceVarP(&searchKeywords, "keywords", "k", []string{}, "키워드로 검색")
	SearchCmd.Flags().StringVar(&sortBy, "sort", "relevance", "정렬 기준 (relevance, name, downloads, rating, updated)")
	SearchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 20, "결과 개수 제한")
	SearchCmd.Flags().BoolVarP(&showDetails, "details", "d", false, "상세 정보 표시")
	SearchCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "출력 형식 (table, json, yaml)")
	SearchCmd.Flags().StringVar(&searchServer, "server", "http://localhost:8080", "템플릿 서버 URL")
	SearchCmd.Flags().StringVar(&searchAPIKey, "api-key", "", "API 키")
	SearchCmd.Flags().StringVar(&searchConfig, "config", "", "클라이언트 설정 파일")
	SearchCmd.Flags().BoolVar(&useLocalIndex, "local", false, "로컬 인덱스 사용")
}

// SearchResult represents search results
type SearchResult struct {
	Query       string       `json:"query"`
	Total       int          `json:"total"`
	Limit       int          `json:"limit"`
	Templates   []Template   `json:"templates"`
	Facets      SearchFacets `json:"facets"`
	SortBy      string       `json:"sortBy"`
	Suggestions []string     `json:"suggestions,omitempty"`
}

// SearchFacets represents search facets
type SearchFacets struct {
	Categories map[string]int `json:"categories"`
	Types      map[string]int `json:"types"`
	Authors    map[string]int `json:"authors"`
	Licenses   map[string]int `json:"licenses"`
}

// SearchFilter represents search filters
type SearchFilter struct {
	Query      string   `json:"query,omitempty"`
	Category   string   `json:"category,omitempty"`
	Type       string   `json:"type,omitempty"`
	Author     string   `json:"author,omitempty"`
	Keywords   []string `json:"keywords,omitempty"`
	License    string   `json:"license,omitempty"`
	Verified   *bool    `json:"verified,omitempty"`
	Deprecated *bool    `json:"deprecated,omitempty"`
}

func runSearch(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		searchQuery = strings.Join(args, " ")
	}

	fmt.Printf("🔍 템플릿 검색\n")
	if searchQuery != "" {
		fmt.Printf("🎯 검색어: %s\n", searchQuery)
	}

	// Build search filter
	filter := SearchFilter{
		Query:    searchQuery,
		Category: searchCategory,
		Type:     searchType,
		Author:   searchAuthor,
		Keywords: searchKeywords,
	}

	// Perform search
	var result *SearchResult
	var err error

	if useLocalIndex {
		result, err = performLocalSearch(filter)
	} else {
		result, err = performRemoteSearch(filter)
	}

	if err != nil {
		fmt.Printf("❌ 검색 실패: %v\n", err)
		os.Exit(1)
	}

	// Display results
	displaySearchResults(result)
}

func performRemoteSearch(filter SearchFilter) (*SearchResult, error) {
	// Setup client
	client, err := setupSearchClient()
	if err != nil {
		return nil, fmt.Errorf("클라이언트 설정 실패: %w", err)
	}

	// Calculate pagination
	page := 1
	perPage := searchLimit
	if searchLimit <= 0 {
		perPage = 20
	}

	// Perform API search
	response, err := client.SearchTemplates(filter.Query, filter.Category, filter.Type, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("API 검색 실패: %w", err)
	}

	// Convert API response to our SearchResult format
	var templates []Template
	for _, apiTemplate := range response.Templates {
		template := Template{
			Name:        apiTemplate.Name,
			Version:     apiTemplate.Version,
			Description: apiTemplate.Description,
			Author:      apiTemplate.Author,
			Category:    apiTemplate.Category,
			Type:        apiTemplate.Type,
			Keywords:    apiTemplate.Keywords,
			Downloads:   apiTemplate.Downloads,
			Rating:      apiTemplate.Rating,
			Created:     apiTemplate.Created,
			Updated:     apiTemplate.Updated,
			License:     apiTemplate.License,
			Homepage:    apiTemplate.Homepage,
			Repository:  apiTemplate.Repository,
			Verified:    apiTemplate.Verified,
			Deprecated:  apiTemplate.Deprecated,
			Tags:        apiTemplate.Tags,
		}
		templates = append(templates, template)
	}

	// Generate facets from results
	facets := generateSearchFacets(templates)

	result := &SearchResult{
		Query:     filter.Query,
		Total:     response.Total,
		Limit:     searchLimit,
		Templates: templates,
		Facets:    facets,
		SortBy:    sortBy,
	}

	return result, nil
}

func setupSearchClient() (*TemplateClient, error) {
	var config *ClientConfig

	// Try to load from config file
	if searchConfig == "" {
		searchConfig = GetDefaultConfigPath()
	}

	if _, err := os.Stat(searchConfig); err == nil {
		loadedConfig, err := LoadClientConfig(searchConfig)
		if err != nil {
			return nil, fmt.Errorf("설정 파일 로드 실패: %w", err)
		}
		config = loadedConfig
	} else {
		// Create default config
		config = &ClientConfig{
			BaseURL: "http://localhost:8080",
			Timeout: 30,
		}
	}

	// Override with command line flags
	if searchServer != "" {
		config.BaseURL = searchServer
	}
	if searchAPIKey != "" {
		config.APIKey = searchAPIKey
	}

	return NewTemplateClient(config), nil
}

func performLocalSearch(filter SearchFilter) (*SearchResult, error) {
	// Load marketplace index
	index, err := loadMarketplaceIndex()
	if err != nil {
		return nil, fmt.Errorf("마켓플레이스 인덱스 로드 실패: %w", err)
	}

	// Apply filters
	var matchedTemplates []Template
	for _, template := range index.Templates {
		if matchesFilter(template, filter) {
			matchedTemplates = append(matchedTemplates, template)
		}
	}

	// Sort results
	sortTemplates(matchedTemplates, sortBy)

	// Apply limit
	if searchLimit > 0 && len(matchedTemplates) > searchLimit {
		matchedTemplates = matchedTemplates[:searchLimit]
	}

	// Generate facets
	facets := generateSearchFacets(matchedTemplates)

	// Generate suggestions if no results
	var suggestions []string
	if len(matchedTemplates) == 0 && filter.Query != "" {
		suggestions = generateSearchSuggestions(filter.Query, index)
	}

	result := &SearchResult{
		Query:       filter.Query,
		Total:       len(matchedTemplates),
		Limit:       searchLimit,
		Templates:   matchedTemplates,
		Facets:      facets,
		SortBy:      sortBy,
		Suggestions: suggestions,
	}

	return result, nil
}

func loadMarketplaceIndex() (*MarketplaceIndex, error) {
	// Default marketplace path
	indexPath := "./marketplace/index/index.json"

	// Try alternative paths
	alternativePaths := []string{
		indexPath,
		"./index/index.json",
		"../marketplace/index/index.json",
	}

	var data []byte
	var err error
	var foundPath string

	for _, path := range alternativePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("인덱스 파일을 찾을 수 없습니다. 마켓플레이스를 먼저 초기화하세요: %w", err)
	}

	var index MarketplaceIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("인덱스 파싱 실패 (%s): %w", foundPath, err)
	}

	return &index, nil
}

func matchesFilter(template Template, filter SearchFilter) bool {
	// Text search in name, description, keywords
	if filter.Query != "" {
		query := strings.ToLower(filter.Query)
		searchableText := strings.ToLower(fmt.Sprintf("%s %s %s",
			template.Name, template.Description, strings.Join(template.Keywords, " ")))

		if !strings.Contains(searchableText, query) {
			return false
		}
	}

	// Category filter
	if filter.Category != "" && template.Category != filter.Category {
		return false
	}

	// Type filter
	if filter.Type != "" && template.Type != filter.Type {
		return false
	}

	// Author filter
	if filter.Author != "" && !strings.EqualFold(template.Author, filter.Author) {
		return false
	}

	// Keywords filter
	if len(filter.Keywords) > 0 {
		templateKeywords := make(map[string]bool)
		for _, keyword := range template.Keywords {
			templateKeywords[strings.ToLower(keyword)] = true
		}

		for _, filterKeyword := range filter.Keywords {
			if !templateKeywords[strings.ToLower(filterKeyword)] {
				return false
			}
		}
	}

	// Deprecated filter
	if filter.Deprecated != nil && template.Deprecated != *filter.Deprecated {
		return false
	}

	// Verified filter
	if filter.Verified != nil && template.Verified != *filter.Verified {
		return false
	}

	return true
}

func sortTemplates(templates []Template, sortBy string) {
	switch sortBy {
	case "name":
		sort.Slice(templates, func(i, j int) bool {
			return strings.ToLower(templates[i].Name) < strings.ToLower(templates[j].Name)
		})
	case "downloads":
		sort.Slice(templates, func(i, j int) bool {
			return templates[i].Downloads > templates[j].Downloads
		})
	case "rating":
		sort.Slice(templates, func(i, j int) bool {
			return templates[i].Rating > templates[j].Rating
		})
	case "updated":
		sort.Slice(templates, func(i, j int) bool {
			return templates[i].Updated.After(templates[j].Updated)
		})
	case "created":
		sort.Slice(templates, func(i, j int) bool {
			return templates[i].Created.After(templates[j].Created)
		})
	default: // relevance
		// For now, sort by downloads as a proxy for relevance
		sort.Slice(templates, func(i, j int) bool {
			return templates[i].Downloads > templates[j].Downloads
		})
	}
}

func generateSearchFacets(templates []Template) SearchFacets {
	facets := SearchFacets{
		Categories: make(map[string]int),
		Types:      make(map[string]int),
		Authors:    make(map[string]int),
		Licenses:   make(map[string]int),
	}

	for _, template := range templates {
		facets.Categories[template.Category]++
		facets.Types[template.Type]++
		facets.Authors[template.Author]++
		facets.Licenses[template.License]++
	}

	return facets
}

func generateSearchSuggestions(query string, index *MarketplaceIndex) []string {
	var suggestions []string
	queryLower := strings.ToLower(query)

	// Suggest similar template names
	for _, template := range index.Templates {
		if strings.Contains(strings.ToLower(template.Name), queryLower) {
			suggestions = append(suggestions, template.Name)
		}
	}

	// Suggest categories
	for category := range index.Categories {
		if strings.Contains(strings.ToLower(category), queryLower) {
			suggestions = append(suggestions, fmt.Sprintf("category:%s", category))
		}
	}

	// Limit suggestions
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}

	return suggestions
}

func displaySearchResults(result *SearchResult) {
	switch outputFormat {
	case "json":
		displayJSONResults(result)
	case "yaml":
		displayYAMLResults(result)
	default:
		displayTableResults(result)
	}
}

func displayTableResults(result *SearchResult) {
	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("📊 검색 결과\n")
	fmt.Printf(strings.Repeat("=", 80) + "\n")

	if result.Query != "" {
		fmt.Printf("🎯 검색어: %s\n", result.Query)
	}
	fmt.Printf("📄 총 %d개 결과 (최대 %d개 표시)\n", result.Total, result.Limit)
	fmt.Printf("📊 정렬: %s\n", result.SortBy)

	if len(result.Templates) == 0 {
		fmt.Printf("\n❌ 검색 결과가 없습니다.\n")

		if len(result.Suggestions) > 0 {
			fmt.Printf("\n💡 제안:\n")
			for _, suggestion := range result.Suggestions {
				fmt.Printf("  • %s\n", suggestion)
			}
		}
		return
	}

	fmt.Printf("\n📋 템플릿 목록:\n")
	fmt.Printf(strings.Repeat("-", 80) + "\n")

	for i, template := range result.Templates {
		fmt.Printf("%d. %s", i+1, template.Name)

		if template.Verified {
			fmt.Printf(" ✅")
		}
		if template.Deprecated {
			fmt.Printf(" ⚠️")
		}

		fmt.Printf("\n")
		fmt.Printf("   📦 %s | 🏷️  %s | 👤 %s | 📅 %s\n",
			template.Type, template.Category, template.Author, template.Version)

		if template.Description != "" {
			desc := template.Description
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}
			fmt.Printf("   📝 %s\n", desc)
		}

		if showDetails {
			fmt.Printf("   📊 다운로드: %d | ⭐ 평점: %.1f | 📄 라이선스: %s\n",
				template.Downloads, template.Rating, template.License)

			if len(template.Keywords) > 0 {
				fmt.Printf("   🏷️  키워드: %s\n", strings.Join(template.Keywords, ", "))
			}

			if template.Homepage != "" {
				fmt.Printf("   🌐 홈페이지: %s\n", template.Homepage)
			}
		}

		fmt.Printf("\n")
	}

	// Display facets
	if !showDetails {
		fmt.Printf("📊 카테고리별 분포:\n")
		for category, count := range result.Facets.Categories {
			fmt.Printf("  • %s: %d개\n", category, count)
		}

		fmt.Printf("\n📦 타입별 분포:\n")
		for templateType, count := range result.Facets.Types {
			fmt.Printf("  • %s: %d개\n", templateType, count)
		}
	}

	fmt.Printf(strings.Repeat("=", 80) + "\n")
	fmt.Printf("💡 사용법: gz template install <템플릿명>\n")
	fmt.Printf("💡 상세 정보: gz template search --details\n")
}

func displayJSONResults(result *SearchResult) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("❌ JSON 출력 실패: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func displayYAMLResults(result *SearchResult) {
	// For simplicity, convert to JSON first then to YAML
	// In a real implementation, use gopkg.in/yaml.v3
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("❌ YAML 출력 실패: %v\n", err)
		return
	}

	// Simple JSON to YAML conversion for demo
	yamlData := strings.ReplaceAll(string(data), "{", "")
	yamlData = strings.ReplaceAll(yamlData, "}", "")
	yamlData = strings.ReplaceAll(yamlData, "[", "")
	yamlData = strings.ReplaceAll(yamlData, "]", "")
	yamlData = strings.ReplaceAll(yamlData, ",", "")
	yamlData = strings.ReplaceAll(yamlData, "\"", "")

	fmt.Println(yamlData)
}
