package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"social_server/internal/models/postgres"
	"social_server/internal/repositories"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/standard"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search/query"
)

type SearchService struct {
	index    bleve.Index
	postRepo repositories.PostRepository
	userRepo repositories.UserRepository
	indexDir string
}

type SearchDocument struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"` // post, user, comment
	Content    string            `json:"content"`
	Title      string            `json:"title,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	AuthorID   string            `json:"author_id,omitempty"`
	AuthorName string            `json:"author_name,omitempty"`
	Privacy    string            `json:"privacy,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type SearchRequest struct {
	Query     string            `json:"query"`
	Type      string            `json:"type,omitempty"` // post, user, comment
	Tags      []string          `json:"tags,omitempty"`
	AuthorID  string            `json:"author_id,omitempty"`
	Privacy   string            `json:"privacy,omitempty"`
	From      int               `json:"from"`
	Size      int               `json:"size"`
	SortBy    string            `json:"sort_by,omitempty"` // relevance, date, popularity
	Filters   map[string]string `json:"filters,omitempty"`
	Facets    []string          `json:"facets,omitempty"`
	Highlight bool              `json:"highlight"`
	Fuzzy     bool              `json:"fuzzy"`
	DateRange *DateRange        `json:"date_range,omitempty"`
}

type DateRange struct {
	From *time.Time `json:"from,omitempty"`
	To   *time.Time `json:"to,omitempty"`
}

type SearchResponse struct {
	Results     []SearchResult         `json:"results"`
	Total       uint64                 `json:"total"`
	Took        int64                  `json:"took"` // Duration in milliseconds
	Facets      map[string]interface{} `json:"facets,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	MaxScore    float64                `json:"max_score"`
	Page        int                    `json:"page"`
	Size        int                    `json:"size"`
	HasMore     bool                   `json:"has_more"`
}

type SearchResult struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Score      float64                `json:"score"`
	Content    string                 `json:"content"`
	Title      string                 `json:"title,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
	AuthorID   string                 `json:"author_id,omitempty"`
	AuthorName string                 `json:"author_name,omitempty"`
	Privacy    string                 `json:"privacy,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	Highlights map[string][]string    `json:"highlights,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type AutoCompleteRequest struct {
	Query string `json:"query"`
	Type  string `json:"type,omitempty"`
	Limit int    `json:"limit"`
}

type AutoCompleteResponse struct {
	Suggestions []AutoCompleteSuggestion `json:"suggestions"`
}

type AutoCompleteSuggestion struct {
	Text  string  `json:"text"`
	Score float64 `json:"score"`
	Type  string  `json:"type"`
}

func NewSearchService(indexDir string, postRepo repositories.PostRepository, userRepo repositories.UserRepository) (*SearchService, error) {
	// Create index directory if it doesn't exist
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create index directory: %w", err)
	}

	indexPath := filepath.Join(indexDir, "search.bleve")

	// Create or open Bleve index
	var index bleve.Index
	var err error

	// Check if index exists
	if _, statErr := os.Stat(indexPath); os.IsNotExist(statErr) {
		// Create new index
		indexMapping := createIndexMapping()
		index, err = bleve.New(indexPath, indexMapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create search index: %w", err)
		}
		log.Println("Created new search index")
	} else {
		// Open existing index
		index, err = bleve.Open(indexPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open search index: %w", err)
		}
		log.Println("Opened existing search index")
	}

	service := &SearchService{
		index:    index,
		postRepo: postRepo,
		userRepo: userRepo,
		indexDir: indexDir,
	}

	return service, nil
}

func createIndexMapping() mapping.IndexMapping {
	// Create a mapping
	indexMapping := bleve.NewIndexMapping()

	// Define analyzers
	standardAnalyzer := standard.Name

	// Create document mapping for posts
	postMapping := bleve.NewDocumentMapping()

	// Content field - full text search
	contentFieldMapping := bleve.NewTextFieldMapping()
	contentFieldMapping.Analyzer = standardAnalyzer
	contentFieldMapping.Store = true
	contentFieldMapping.Index = true
	postMapping.AddFieldMappingsAt("content", contentFieldMapping)

	// Title field
	titleFieldMapping := bleve.NewTextFieldMapping()
	titleFieldMapping.Analyzer = standardAnalyzer
	titleFieldMapping.Store = true
	titleFieldMapping.Index = true
	postMapping.AddFieldMappingsAt("title", titleFieldMapping)

	// Tags field - keyword search
	tagsFieldMapping := bleve.NewKeywordFieldMapping()
	tagsFieldMapping.Store = true
	tagsFieldMapping.Index = true
	postMapping.AddFieldMappingsAt("tags", tagsFieldMapping)

	// Type field
	typeFieldMapping := bleve.NewKeywordFieldMapping()
	typeFieldMapping.Store = true
	typeFieldMapping.Index = true
	postMapping.AddFieldMappingsAt("type", typeFieldMapping)

	// Author fields
	authorIDFieldMapping := bleve.NewKeywordFieldMapping()
	authorIDFieldMapping.Store = true
	authorIDFieldMapping.Index = true
	postMapping.AddFieldMappingsAt("author_id", authorIDFieldMapping)

	authorNameFieldMapping := bleve.NewTextFieldMapping()
	authorNameFieldMapping.Analyzer = standardAnalyzer
	authorNameFieldMapping.Store = true
	authorNameFieldMapping.Index = true
	postMapping.AddFieldMappingsAt("author_name", authorNameFieldMapping)

	// Privacy field
	privacyFieldMapping := bleve.NewKeywordFieldMapping()
	privacyFieldMapping.Store = true
	privacyFieldMapping.Index = true
	postMapping.AddFieldMappingsAt("privacy", privacyFieldMapping)

	// Date fields
	dateFieldMapping := bleve.NewDateTimeFieldMapping()
	dateFieldMapping.Store = true
	dateFieldMapping.Index = true
	postMapping.AddFieldMappingsAt("created_at", dateFieldMapping)
	postMapping.AddFieldMappingsAt("updated_at", dateFieldMapping)

	// Set default mapping
	indexMapping.DefaultMapping = postMapping

	return indexMapping
}

func (s *SearchService) IndexPost(ctx context.Context, post *postgres.Post) error {
	doc := SearchDocument{
		ID:        fmt.Sprintf("%d", post.ID),
		Type:      "post",
		Content:   post.Content,
		Tags:      []string{},
		AuthorID:  fmt.Sprintf("%d", post.AuthorID),
		Privacy:   string(post.Privacy),
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}

	// Get author info
	if post.Author.ID != 0 {
		// Use email as fallback since username is removed
		doc.AuthorName = post.Author.Email
		// TODO: Get display name from user's profiles if needed
	}

	// Add metadata
	doc.Metadata = map[string]string{
		"like_count":    fmt.Sprintf("%d", post.LikeCount),
		"comment_count": fmt.Sprintf("%d", post.CommentCount),
		"share_count":   fmt.Sprintf("%d", post.ShareCount),
	}

	return s.index.Index(doc.ID, doc)
}

func (s *SearchService) IndexUser(ctx context.Context, user *postgres.User) error {
	// Use email as searchable content since username/displayName moved to Profile
	doc := SearchDocument{
		ID:        fmt.Sprintf("%d", user.ID),
		Type:      "user",
		Content:   user.Email,
		Title:     user.Email,
		AuthorID:  fmt.Sprintf("%d", user.ID),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Metadata: map[string]string{
			"email":        user.Email,
			"is_verified":  fmt.Sprintf("%t", user.IsVerified),
			"is_active":    fmt.Sprintf("%t", user.IsActive),
		},
	}

	return s.index.Index(doc.ID, doc)
}

func (s *SearchService) IndexComment(ctx context.Context, comment *postgres.Comment) error {
	doc := SearchDocument{
		ID:        fmt.Sprintf("%d", comment.ID),
		Type:      "comment",
		Content:   comment.Content,
		AuthorID:  fmt.Sprintf("%d", comment.UserID),
		Privacy:   "public",
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
		Metadata: map[string]string{
			"post_id":    fmt.Sprintf("%d", comment.PostID),
			"like_count": fmt.Sprintf("%d", comment.LikeCount),
		},
	}

	// Get author info
	// TODO: Load comment author from UserID if needed
	doc.AuthorName = "Unknown"

	return s.index.Index(doc.ID, doc)
}

func (s *SearchService) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	// Build query
	q := s.buildQuery(req)

	// Create search request
	searchReq := bleve.NewSearchRequestOptions(q, req.Size, req.From, false)

	// Add highlighting
	if req.Highlight {
		searchReq.Highlight = bleve.NewHighlight()
		searchReq.Highlight.AddField("content")
		searchReq.Highlight.AddField("title")
	}

	// Add facets
	if len(req.Facets) > 0 {
		searchReq.Facets = make(bleve.FacetsRequest)
		for _, facet := range req.Facets {
			searchReq.Facets[facet] = bleve.NewFacetRequest(facet, 10)
		}
	}

	// Add sorting
	if req.SortBy != "" {
		searchReq.SortBy([]string{req.SortBy})
	}

	// Execute search
	start := time.Now()
	searchResult, err := s.index.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	took := time.Since(start)

	// Convert results
	results := make([]SearchResult, 0, len(searchResult.Hits))
	for _, hit := range searchResult.Hits {
		result := SearchResult{
			ID:    hit.ID,
			Score: hit.Score,
		}

		// Parse stored fields
		if hit.Fields != nil {
			if v, ok := hit.Fields["type"]; ok {
				result.Type = fmt.Sprintf("%v", v)
			}
			if v, ok := hit.Fields["content"]; ok {
				result.Content = fmt.Sprintf("%v", v)
			}
			if v, ok := hit.Fields["title"]; ok {
				result.Title = fmt.Sprintf("%v", v)
			}
			if v, ok := hit.Fields["author_id"]; ok {
				result.AuthorID = fmt.Sprintf("%v", v)
			}
			if v, ok := hit.Fields["author_name"]; ok {
				result.AuthorName = fmt.Sprintf("%v", v)
			}
			if v, ok := hit.Fields["privacy"]; ok {
				result.Privacy = fmt.Sprintf("%v", v)
			}
			if v, ok := hit.Fields["tags"]; ok {
				if tags, ok := v.([]interface{}); ok {
					result.Tags = make([]string, len(tags))
					for i, tag := range tags {
						result.Tags[i] = fmt.Sprintf("%v", tag)
					}
				}
			}
			if v, ok := hit.Fields["created_at"]; ok {
				if createdAt, ok := v.(string); ok {
					if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
						result.CreatedAt = t
					}
				}
			}
		}

		// Add highlights
		if hit.Fragments != nil && len(hit.Fragments) > 0 {
			result.Highlights = make(map[string][]string)
			for field, fragments := range hit.Fragments {
				result.Highlights[field] = fragments
			}
		}

		results = append(results, result)
	}

	response := &SearchResponse{
		Results:  results,
		Total:    searchResult.Total,
		Took:     took.Milliseconds(),
		MaxScore: searchResult.MaxScore,
		Page:     req.From/req.Size + 1,
		Size:     req.Size,
		HasMore:  searchResult.Total > uint64(req.From+req.Size),
	}

	// Add facets
	if searchResult.Facets != nil {
		response.Facets = make(map[string]interface{})
		for name, facetResult := range searchResult.Facets {
			response.Facets[name] = facetResult
		}
	}

	return response, nil
}

func (s *SearchService) buildQuery(req *SearchRequest) query.Query {
	var queries []query.Query

	// Main query
	if req.Query != "" {
		if req.Fuzzy {
			// Fuzzy search
			fuzzyQuery := bleve.NewFuzzyQuery(req.Query)
			fuzzyQuery.SetFuzziness(2)
			queries = append(queries, fuzzyQuery)
		} else {
			// Match query for content
			matchQuery := bleve.NewMatchQuery(req.Query)
			matchQuery.SetField("content")
			queries = append(queries, matchQuery)

			// Also add query string query for multi-field search
			queryStringQuery := bleve.NewQueryStringQuery(req.Query)
			queries = append(queries, queryStringQuery)
		}
	}

	// Type filter
	if req.Type != "" {
		typeQuery := bleve.NewTermQuery(req.Type)
		typeQuery.SetField("type")
		queries = append(queries, typeQuery)
	}

	// Tags filter
	if len(req.Tags) > 0 {
		for _, tag := range req.Tags {
			tagQuery := bleve.NewTermQuery(tag)
			tagQuery.SetField("tags")
			queries = append(queries, tagQuery)
		}
	}

	// Author filter
	if req.AuthorID != "" {
		authorQuery := bleve.NewTermQuery(req.AuthorID)
		authorQuery.SetField("author_id")
		queries = append(queries, authorQuery)
	}

	// Privacy filter
	if req.Privacy != "" {
		privacyQuery := bleve.NewTermQuery(req.Privacy)
		privacyQuery.SetField("privacy")
		queries = append(queries, privacyQuery)
	}

	// Date range filter
	if req.DateRange != nil {
		if req.DateRange.From != nil && req.DateRange.To != nil {
			dateQuery := bleve.NewDateRangeQuery(*req.DateRange.From, *req.DateRange.To)
			dateQuery.SetField("created_at")
			queries = append(queries, dateQuery)
		} else if req.DateRange.From != nil {
			dateQuery := bleve.NewDateRangeQuery(*req.DateRange.From, time.Now())
			dateQuery.SetField("created_at")
			queries = append(queries, dateQuery)
		} else if req.DateRange.To != nil {
			dateQuery := bleve.NewDateRangeQuery(time.Time{}, *req.DateRange.To)
			dateQuery.SetField("created_at")
			queries = append(queries, dateQuery)
		}
	}

	// Additional filters
	for field, value := range req.Filters {
		filterQuery := bleve.NewTermQuery(value)
		filterQuery.SetField(field)
		queries = append(queries, filterQuery)
	}

	// Combine queries
	if len(queries) == 0 {
		return bleve.NewMatchAllQuery()
	} else if len(queries) == 1 {
		return queries[0]
	} else {
		conjunctionQuery := bleve.NewConjunctionQuery(queries...)
		return conjunctionQuery
	}
}

func (s *SearchService) AutoComplete(ctx context.Context, req *AutoCompleteRequest) (*AutoCompleteResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Create prefix query
	prefixQuery := bleve.NewPrefixQuery(req.Query)

	// Add type filter if specified
	var finalQuery query.Query = prefixQuery
	if req.Type != "" {
		typeQuery := bleve.NewTermQuery(req.Type)
		typeQuery.SetField("type")
		finalQuery = bleve.NewConjunctionQuery(prefixQuery, typeQuery)
	}

	// Create search request
	searchReq := bleve.NewSearchRequestOptions(finalQuery, req.Limit, 0, false)
	searchReq.Fields = []string{"content", "title", "type"}

	// Execute search
	searchResult, err := s.index.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("autocomplete search failed: %w", err)
	}

	// Convert results
	suggestions := make([]AutoCompleteSuggestion, 0, len(searchResult.Hits))
	seen := make(map[string]bool)

	for _, hit := range searchResult.Hits {
		var text string
		var docType string

		if hit.Fields != nil {
			if v, ok := hit.Fields["title"]; ok && v != nil {
				text = fmt.Sprintf("%v", v)
			} else if v, ok := hit.Fields["content"]; ok && v != nil {
				content := fmt.Sprintf("%v", v)
				if len(content) > 100 {
					content = content[:100] + "..."
				}
				text = content
			}

			if v, ok := hit.Fields["type"]; ok {
				docType = fmt.Sprintf("%v", v)
			}
		}

		if text != "" && !seen[text] {
			suggestions = append(suggestions, AutoCompleteSuggestion{
				Text:  text,
				Score: hit.Score,
				Type:  docType,
			})
			seen[text] = true
		}
	}

	return &AutoCompleteResponse{
		Suggestions: suggestions,
	}, nil
}

func (s *SearchService) DeleteDocument(id string) error {
	return s.index.Delete(id)
}

func (s *SearchService) UpdateDocument(id string, doc interface{}) error {
	return s.index.Index(id, doc)
}

func (s *SearchService) GetStats() (map[string]interface{}, error) {
	docCount, err := s.index.DocCount()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"document_count": docCount,
		"index_size":     s.getIndexSize(),
	}, nil
}

func (s *SearchService) getIndexSize() int64 {
	var size int64
	filepath.Walk(s.indexDir, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			size += info.Size()
		}
		return nil
	})
	return size
}

func (s *SearchService) ReindexAll(ctx context.Context) error {
	// This is a simplified reindex - in production you'd want to do this more carefully
	log.Println("Starting reindex of all documents...")

	// Clear existing index
	batch := s.index.NewBatch()

	// You would implement actual reindexing logic here
	// For now, just return success
	err := s.index.Batch(batch)
	if err != nil {
		return fmt.Errorf("failed to clear index: %w", err)
	}

	log.Println("Reindex completed successfully")
	return nil
}

func (s *SearchService) Close() error {
	return s.index.Close()
}
