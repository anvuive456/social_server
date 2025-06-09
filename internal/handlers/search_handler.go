package handlers

import (
	"net/http"
	"social_server/internal/middleware"
	"social_server/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchService *services.SearchService
}

func NewSearchHandler(searchService *services.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// Search performs global search across all content types
// @Summary Global search
// @Description Search across posts, users, and comments with flexible filtering
// @Tags Search
// @Produce json
// @Param q query string false "Search query"
// @Param type query string false "Content type filter (post, user, comment)"
// @Param author_id query string false "Filter by author ID"
// @Param privacy query string false "Privacy filter (public, friends, private)"
// @Param sort_by query string false "Sort order (relevance, date, popularity)" default(relevance)
// @Param from query int false "Pagination offset" default(0)
// @Param size query int false "Number of results (1-100)" default(20)
// @Param highlight query bool false "Enable result highlighting" default(false)
// @Param fuzzy query bool false "Enable fuzzy search" default(false)
// @Param tags query []string false "Filter by tags"
// @Param facets query []string false "Enable faceted search"
// @Param from_date query string false "Start date filter (RFC3339 format)"
// @Param to_date query string false "End date filter (RFC3339 format)"
// @Success 200 {object} services.SearchResponse "Search results"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Search failed"
// @Router /search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	var req services.SearchRequest

	// Parse query parameters
	req.Query = c.Query("q")
	req.Type = c.Query("type")
	req.AuthorID = c.Query("author_id")
	req.Privacy = c.Query("privacy")
	req.SortBy = c.DefaultQuery("sort_by", "relevance")

	// Parse pagination
	from, err := strconv.Atoi(c.DefaultQuery("from", "0"))
	if err != nil || from < 0 {
		from = 0
	}
	req.From = from

	size, err := strconv.Atoi(c.DefaultQuery("size", "20"))
	if err != nil || size < 1 || size > 100 {
		size = 20
	}
	req.Size = size

	// Parse boolean flags
	req.Highlight = c.Query("highlight") == "true"
	req.Fuzzy = c.Query("fuzzy") == "true"

	// Parse tags
	if tags := c.QueryArray("tags"); len(tags) > 0 {
		req.Tags = tags
	}

	// Parse facets
	if facets := c.QueryArray("facets"); len(facets) > 0 {
		req.Facets = facets
	}

	// Parse date range
	if fromDate := c.Query("from_date"); fromDate != "" {
		if t, err := time.Parse(time.RFC3339, fromDate); err == nil {
			if req.DateRange == nil {
				req.DateRange = &services.DateRange{}
			}
			req.DateRange.From = &t
		}
	}

	if toDate := c.Query("to_date"); toDate != "" {
		if t, err := time.Parse(time.RFC3339, toDate); err == nil {
			if req.DateRange == nil {
				req.DateRange = &services.DateRange{}
			}
			req.DateRange.To = &t
		}
	}

	// Parse additional filters
	req.Filters = make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 && key != "q" && key != "type" && key != "author_id" && 
		   key != "privacy" && key != "sort_by" && key != "from" && key != "size" && 
		   key != "highlight" && key != "fuzzy" && key != "tags" && key != "facets" &&
		   key != "from_date" && key != "to_date" {
			req.Filters[key] = values[0]
		}
	}

	// Validate required parameters
	if req.Query == "" && req.Type == "" && len(req.Tags) == 0 && req.AuthorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "At least one search parameter is required (q, type, tags, or author_id)",
		})
		return
	}

	// Execute search
	response, err := h.searchService.Search(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "search_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AdvancedSearch performs advanced search with complex filters
// @Summary Advanced search
// @Description Perform advanced search with complex JSON-based filtering
// @Tags Search
// @Accept json
// @Produce json
// @Param search body services.SearchRequest true "Advanced search parameters"
// @Success 200 {object} services.SearchResponse "Search results"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Search failed"
// @Router /search/advanced [post]
func (h *SearchHandler) AdvancedSearch(c *gin.Context) {
	var req services.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// Set defaults
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}
	if req.From < 0 {
		req.From = 0
	}

	// Execute search
	response, err := h.searchService.Search(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "search_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AutoComplete provides search suggestions
// @Summary Auto-complete search
// @Description Get search suggestions for partial queries
// @Tags Search
// @Produce json
// @Param q query string true "Partial search query"
// @Param type query string false "Content type filter (post, user, comment)"
// @Param limit query int false "Number of suggestions (1-50)" default(10)
// @Success 200 {object} services.AutoCompleteResponse "Search suggestions"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Autocomplete failed"
// @Router /search/autocomplete [get]
func (h *SearchHandler) AutoComplete(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Query parameter 'q' is required",
		})
		return
	}

	docType := c.Query("type")
	
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	req := &services.AutoCompleteRequest{
		Query: query,
		Type:  docType,
		Limit: limit,
	}

	response, err := h.searchService.AutoComplete(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "autocomplete_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchPosts searches specifically for posts
// @Summary Search posts
// @Description Search for posts with specific filters
// @Tags Search
// @Produce json
// @Param q query string false "Search query"
// @Param from query int false "Pagination offset" default(0)
// @Param size query int false "Number of results (1-100)" default(20)
// @Param author_id query string false "Filter by author ID"
// @Param privacy query string false "Privacy filter (public, friends, private)"
// @Param tags query []string false "Filter by tags"
// @Param highlight query bool false "Enable result highlighting" default(false)
// @Param sort_by query string false "Sort order (relevance, date, popularity)" default(relevance)
// @Success 200 {object} services.SearchResponse "Post search results"
// @Failure 500 {object} map[string]interface{} "Search failed"
// @Router /search/posts [get]
func (h *SearchHandler) SearchPosts(c *gin.Context) {
	req := services.SearchRequest{
		Query: c.Query("q"),
		Type:  "post",
		From:  0,
		Size:  20,
	}

	// Parse pagination
	if from, err := strconv.Atoi(c.Query("from")); err == nil && from >= 0 {
		req.From = from
	}
	if size, err := strconv.Atoi(c.Query("size")); err == nil && size > 0 && size <= 100 {
		req.Size = size
	}

	// Parse filters
	if authorID := c.Query("author_id"); authorID != "" {
		req.AuthorID = authorID
	}
	if privacy := c.Query("privacy"); privacy != "" {
		req.Privacy = privacy
	}
	if tags := c.QueryArray("tags"); len(tags) > 0 {
		req.Tags = tags
	}

	req.Highlight = c.Query("highlight") == "true"
	req.SortBy = c.DefaultQuery("sort_by", "relevance")

	response, err := h.searchService.Search(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "search_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchUsers searches specifically for users
// @Summary Search users
// @Description Search for users by username, display name, or bio
// @Tags Search
// @Produce json
// @Param q query string false "Search query"
// @Param from query int false "Pagination offset" default(0)
// @Param size query int false "Number of results (1-100)" default(20)
// @Param highlight query bool false "Enable result highlighting" default(false)
// @Param sort_by query string false "Sort order (relevance, date)" default(relevance)
// @Success 200 {object} services.SearchResponse "User search results"
// @Failure 500 {object} map[string]interface{} "Search failed"
// @Router /search/users [get]
func (h *SearchHandler) SearchUsers(c *gin.Context) {
	req := services.SearchRequest{
		Query: c.Query("q"),
		Type:  "user",
		From:  0,
		Size:  20,
	}

	// Parse pagination
	if from, err := strconv.Atoi(c.Query("from")); err == nil && from >= 0 {
		req.From = from
	}
	if size, err := strconv.Atoi(c.Query("size")); err == nil && size > 0 && size <= 100 {
		req.Size = size
	}

	req.Highlight = c.Query("highlight") == "true"
	req.SortBy = c.DefaultQuery("sort_by", "relevance")

	response, err := h.searchService.Search(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "search_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchComments searches specifically for comments
// @Summary Search comments
// @Description Search for comments with filtering options
// @Tags Search
// @Produce json
// @Param q query string false "Search query"
// @Param from query int false "Pagination offset" default(0)
// @Param size query int false "Number of results (1-100)" default(20)
// @Param author_id query string false "Filter by author ID"
// @Param highlight query bool false "Enable result highlighting" default(false)
// @Param sort_by query string false "Sort order (relevance, date)" default(relevance)
// @Success 200 {object} services.SearchResponse "Comment search results"
// @Failure 500 {object} map[string]interface{} "Search failed"
// @Router /search/comments [get]
func (h *SearchHandler) SearchComments(c *gin.Context) {
	req := services.SearchRequest{
		Query: c.Query("q"),
		Type:  "comment",
		From:  0,
		Size:  20,
	}

	// Parse pagination
	if from, err := strconv.Atoi(c.Query("from")); err == nil && from >= 0 {
		req.From = from
	}
	if size, err := strconv.Atoi(c.Query("size")); err == nil && size > 0 && size <= 100 {
		req.Size = size
	}

	// Parse filters
	if authorID := c.Query("author_id"); authorID != "" {
		req.AuthorID = authorID
	}

	req.Highlight = c.Query("highlight") == "true"
	req.SortBy = c.DefaultQuery("sort_by", "relevance")

	response, err := h.searchService.Search(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "search_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetSearchStats retrieves search index statistics
// @Summary Get search statistics
// @Description Get statistics about the search index (admin only)
// @Tags Search
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Search statistics"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Stats failed"
// @Router /search/admin/stats [get]
func (h *SearchHandler) GetSearchStats(c *gin.Context) {
	stats, err := h.searchService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "stats_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// ReindexAll rebuilds the entire search index
// @Summary Reindex all content
// @Description Rebuild the entire search index (admin only)
// @Tags Search
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Reindex successful"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Reindex failed"
// @Router /search/admin/reindex [post]
func (h *SearchHandler) ReindexAll(c *gin.Context) {
	// Only allow admin users to reindex
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Authentication required",
		})
		return
	}

	// TODO: Add admin check
	_ = userID

	err := h.searchService.ReindexAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "reindex_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Reindex completed successfully",
	})
}

// DeleteDocument removes a document from search index
// @Summary Delete search document
// @Description Remove a specific document from the search index (admin only)
// @Tags Search
// @Produce json
// @Security BearerAuth
// @Param id path string true "Document ID"
// @Success 200 {object} map[string]interface{} "Document deleted"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Delete failed"
// @Router /search/admin/documents/{id} [delete]
func (h *SearchHandler) DeleteDocument(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Authentication required",
		})
		return
	}

	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Document ID is required",
		})
		return
	}

	// TODO: Add authorization check to ensure user can delete this document
	_ = userID

	err := h.searchService.DeleteDocument(documentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "delete_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Document deleted from search index",
	})
}

// GetTrendingTags retrieves trending hashtags
// @Summary Get trending tags
// @Description Get currently trending hashtags based on post frequency
// @Tags Search
// @Produce json
// @Param limit query int false "Number of tags to return (1-50)" default(10)
// @Success 200 {object} map[string]interface{} "Trending tags"
// @Failure 500 {object} map[string]interface{} "Trends failed"
// @Router /search/trending-tags [get]
func (h *SearchHandler) GetTrendingTags(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	// Use a match-all query with tags facet to get trending tags
	req := &services.SearchRequest{
		Query:  "*",
		Type:   "post",
		From:   0,
		Size:   1, // We don't need the documents, just the facets
		Facets: []string{"tags"},
	}

	response, err := h.searchService.Search(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "trends_failed",
			"message": err.Error(),
		})
		return
	}

	// Extract trending tags from facets
	trendingTags := make([]map[string]interface{}, 0)
	if response.Facets != nil {
		if tagsFacet, ok := response.Facets["tags"]; ok {
			// Process facet results to extract trending tags
			// The exact structure depends on Bleve's facet response format
			trendingTags = append(trendingTags, map[string]interface{}{
				"facets": tagsFacet,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"trending_tags": trendingTags,
		"limit":         limit,
	})
}