package handlers

import (
	"net/http"
	"social_server/internal/middleware"
	"social_server/internal/models/postgres"
	"social_server/internal/models/requests"
	"social_server/internal/models/responses"

	"social_server/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)

// Dummy for swaggo
var _ = postgres.Post{}
var _ = responses.PostResponse{}

type PostHandler struct {
	postService *services.PostService
}

func NewPostHandler(postService *services.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

// CreatePost creates a new post
// @Summary Create a new post
// @Description Create a new post with text, images, or videos
// @Tags Posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param post body requests.CreatePostRequest true "Post data"
// @Success 201 {object} postgres.Post "Created post"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	var req requests.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	post, err := h.postService.CreatePost(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "create_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    post,
		"message": "Post created successfully",
	})
}

// GetPost retrieves a specific post by ID
// @Summary Get post by ID
// @Description Retrieve a specific post by its ID
// @Tags Posts
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} postgres.Post "Post data"
// @Failure 400 {object} map[string]interface{} "Invalid post ID"
// @Failure 404 {object} map[string]interface{} "Post not found"
// @Router /posts/{id} [get]
func (h *PostHandler) GetPost(c *gin.Context) {
	postIDParam := c.Param("id")
	postID, err := strconv.ParseUint(postIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_post_id",
			"message": "Invalid post ID format",
		})
		return
	}

	// Get current user ID if authenticated
	userID, _ := middleware.GetUserID(c)

	post, err := h.postService.GetPost(uint(postID), &userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "post_not_found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": post,
	})
}

// UpdatePost updates an existing post
// @Summary Update post
// @Description Update an existing post
// @Tags Posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Param post body requests.UpdatePostRequest true "Post update data"
// @Success 200 {object} postgres.Post "Updated post"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Post not found"
// @Router /posts/{id} [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	postIDParam := c.Param("id")
	postID, err := strconv.ParseUint(postIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_post_id",
			"message": "Invalid post ID format",
		})
		return
	}

	var req requests.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	post, err := h.postService.UpdatePost(uint(postID), userID, &req)
	if err != nil {
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "post_not_found",
				"message": err.Error(),
			})
		} else if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "You can only update your own posts",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "update_failed",
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    post,
		"message": "Post updated successfully",
	})
}

// DeletePost deletes a post
// @Summary Delete post
// @Description Delete a post
// @Tags Posts
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 200 {object} map[string]interface{} "Success message"
// @Failure 400 {object} map[string]interface{} "Invalid post ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Post not found"
// @Router /posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	postIDParam := c.Param("id")
	postID, err := strconv.ParseUint(postIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_post_id",
			"message": "Invalid post ID format",
		})
		return
	}

	err = h.postService.DeletePost(uint(postID), userID)
	if err != nil {
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "post_not_found",
				"message": err.Error(),
			})
		} else if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "You can only delete your own posts",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "delete_failed",
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post deleted successfully",
	})
}

// GetUserPosts retrieves posts by a specific user
// @Summary Get user posts
// @Description Get posts by a specific user with cursor-based pagination
// @Tags Posts
// @Produce json
// @Param user_id path int true "User ID"
// @Param limit query int false "Number of posts to return" default(20)
// @Param cursor query string false "Cursor for pagination"
// @Success 200 {object} responses.PostResponse "Posts data"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Router /posts/user/{user_id} [get]
func (h *PostHandler) GetPosts(c *gin.Context) {
	var req requests.GetPostsRequest

	if err := c.BindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request",
		})
		return
	}

	cursor := paginator.Cursor{
		Before: &req.Before,
		After:  &req.After,
	}

	// Get current user ID if authenticated
	currentUserID, _ := middleware.GetUserID(c)

	// Parse limit
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}

	response, err := h.postService.GetUserPosts(uint(req.UserID), &currentUserID, req.Limit, cursor)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "get_posts_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// GetFeed retrieves the user's personalized feed
// @Summary Get user feed
// @Description Get personalized feed with posts from friends
// @Tags Posts
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of posts to return" default(20)
// @Param cursor query string false "Cursor for pagination"
// @Success 200 {object} responses.PostResponse "Feed data"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /posts/feed [get]
func (h *PostHandler) GetFeed(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	// Parse limit
	limitParam := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	cursor := c.Query("cursor")

	response, err := h.postService.GetFeed(userID, limit, cursor)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "get_feed_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// GetPublicFeed retrieves public posts for non-authenticated users
// @Summary Get public feed
// @Description Get public posts for explore page
// @Tags Posts
// @Produce json
// @Param limit query int false "Number of posts to return" default(20)
// @Param cursor query string false "Cursor for pagination"
// @Success 200 {object} responses.PostResponse "Public feed data"
// @Router /posts/public [get]
func (h *PostHandler) GetPublicFeed(c *gin.Context) {
	// Parse limit
	limitParam := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	cursor := c.Query("cursor")

	// Get current user ID if authenticated
	// userID, _ := middleware.GetUserID(c)

	response, err := h.postService.GetPublicFeed(limit, cursor)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "get_public_feed_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// SearchPosts searches for posts
// @Summary Search posts
// @Description Search posts by content or tags
// @Tags Posts
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Number of posts to return" default(20)
// @Param cursor query string false "Cursor for pagination"
// @Success 200 {object} responses.PostResponse "Search results"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Router /posts/search [get]
func (h *PostHandler) SearchPosts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_query",
			"message": "Search query is required",
		})
		return
	}

	// Parse limit
	limitParam := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	cursor := c.Query("cursor")

	// Get current user ID if authenticated
	userID, _ := middleware.GetUserID(c)

	response, err := h.postService.SearchPosts(query, &userID, limit, cursor)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "search_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// LikePost likes or unlikes a post
// @Summary Like/Unlike post
// @Description Toggle like status on a post
// @Tags Posts
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Param like_type query string false "Type of like" default("like")
// @Success 200 {object} map[string]interface{} "Success message"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Post not found"
// @Router /posts/{id}/like [post]
func (h *PostHandler) LikePost(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	postIDParam := c.Param("id")
	postID, err := strconv.ParseUint(postIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_post_id",
			"message": "Invalid post ID format",
		})
		return
	}

	likeType := c.DefaultQuery("like_type", "like")

	isLiked, err := h.postService.ToggleLike(uint(postID), userID, likeType)
	if err != nil {
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "post_not_found",
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "like_failed",
				"message": err.Error(),
			})
		}
		return
	}

	message := "Post liked successfully"
	if !isLiked {
		message = "Post unliked successfully"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  message,
		"is_liked": isLiked,
	})
}

// CreateComment creates a comment on a post
// @Summary Create comment
// @Description Create a comment on a post
// @Tags Posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Param comment body requests.CreateCommentRequest true "Comment data"
// @Success 201 {object} postgres.Comment "Created comment"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Post not found"
// @Router /posts/{id}/comments [post]
func (h *PostHandler) CreateComment(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	postIDParam := c.Param("id")
	postID, err := strconv.ParseUint(postIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_post_id",
			"message": "Invalid post ID format",
		})
		return
	}

	var req requests.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	comment, err := h.postService.CreateComment(uint(postID), userID, &req)
	if err != nil {
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "post_not_found",
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "comment_failed",
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    comment,
		"message": "Comment created successfully",
	})
}

// GetComments retrieves comments for a post
// @Summary Get post comments
// @Description Get comments for a specific post with cursor-based pagination
// @Tags Posts
// @Produce json
// @Param id path int true "Post ID"
// @Param limit query int false "Number of comments to return" default(20)
// @Param cursor query string false "Cursor for pagination"
// @Success 200 {object} responses.CommentResponse "Comments data"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Post not found"
// @Router /posts/{id}/comments [get]
func (h *PostHandler) GetComments(c *gin.Context) {
	postIDParam := c.Param("id")
	postID, err := strconv.ParseUint(postIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_post_id",
			"message": "Invalid post ID format",
		})
		return
	}

	// Parse request
	var req requests.GetPostCommentsRequest
	err = c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	response, err := h.postService.GetComments(uint(postID), &req)
	if err != nil {
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "post_not_found",
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "get_comments_failed",
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// SharePost shares a post
// @Summary Share post
// @Description Share a post to user's timeline
// @Tags Posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Param share body requests.SharePostRequest true "Share data"
// @Success 201 {object} postgres.Share "Created share"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Post not found"
// @Router /posts/{id}/share [post]
func (h *PostHandler) SharePost(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	postIDParam := c.Param("id")
	postID, err := strconv.ParseUint(postIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_post_id",
			"message": "Invalid post ID format",
		})
		return
	}

	var req requests.SharePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	share, err := h.postService.SharePost(uint(postID), userID, &req)
	if err != nil {
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "post_not_found",
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "share_failed",
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    share,
		"message": "Post shared successfully",
	})
}
