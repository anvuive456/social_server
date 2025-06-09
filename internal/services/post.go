package services

import (
	"fmt"
	"social_server/internal/models/constants"
	"social_server/internal/models/postgres"
	"social_server/internal/models/requests"
	"social_server/internal/models/responses"
	"social_server/internal/repositories"
	"time"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)

type PostService struct {
	postRepo    repositories.PostRepository
	userRepo    repositories.UserRepository
	commentRepo repositories.CommentRepository
	likeRepo    repositories.LikeRepository
	shareRepo   repositories.ShareRepository
}

func NewPostService(
	postRepo repositories.PostRepository,
	userRepo repositories.UserRepository,
	commentRepo repositories.CommentRepository,
	likeRepo repositories.LikeRepository,
	shareRepo repositories.ShareRepository,
) *PostService {
	return &PostService{
		postRepo:    postRepo,
		userRepo:    userRepo,
		commentRepo: commentRepo,
		likeRepo:    likeRepo,
		shareRepo:   shareRepo,
	}
}

func (s *PostService) CreatePost(userID uint, req *requests.CreatePostRequest) (*postgres.Post, error) {
	// Validate user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Validate content
	if req.Content == "" && len(req.Media) == 0 {
		return nil, fmt.Errorf("post must have content or media")
	}

	// Create post
	post := &postgres.Post{
		AuthorID:  userID,
		Type:      req.Type,
		Content:   req.Content,
		Privacy:   req.Privacy,
		Location:  req.Location,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set default privacy if not specified
	if post.Privacy == "" {
		post.Privacy = constants.PostPrivacyPublic
	}

	// Set default type if not specified
	if post.Type == "" {
		post.Type = constants.PostTypeText
	}

	// Create post in database
	err = s.postRepo.Create(post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Create media if provided
	for i, mediaReq := range req.Media {
		media := &postgres.PostMedia{
			PostID:    post.ID,
			Type:      mediaReq.Type,
			URL:       mediaReq.URL,
			Filename:  mediaReq.Filename,
			Size:      mediaReq.Size,
			MimeType:  mediaReq.MimeType,
			Width:     mediaReq.Width,
			Height:    mediaReq.Height,
			Duration:  mediaReq.Duration,
			Thumbnail: mediaReq.Thumbnail,
			AltText:   mediaReq.AltText,
			Order:     i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = s.postRepo.CreateMedia(media)
		if err != nil {
			return nil, fmt.Errorf("failed to create post media: %w", err)
		}
	}

	// Return created post with relations
	return s.postRepo.GetByID(post.ID)
}

func (s *PostService) GetPost(postID uint, userID *uint) (*postgres.Post, error) {
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, fmt.Errorf("post not found")
	}

	// Check visibility
	canView, err := s.CheckPostVisibility(postID, userID)
	if err != nil {
		return nil, err
	}

	if !canView {
		return nil, fmt.Errorf("post not found")
	}

	return post, nil
}

func (s *PostService) UpdatePost(postID, userID uint, req *requests.UpdatePostRequest) (*postgres.Post, error) {
	// Get existing post
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, fmt.Errorf("post not found")
	}

	// Check ownership
	if post.AuthorID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.Content != "" {
		updates["content"] = req.Content
		updates["is_edited"] = true
		updates["edited_at"] = time.Now()
	}
	if req.Privacy != "" {
		updates["privacy"] = req.Privacy
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	updates["updated_at"] = time.Now()

	err = s.postRepo.Update(postID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return s.postRepo.GetByID(postID)
}

func (s *PostService) DeletePost(postID, userID uint) error {
	// Get post to check ownership
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return fmt.Errorf("post not found")
	}

	// Check ownership
	if post.AuthorID != userID {
		return fmt.Errorf("unauthorized")
	}

	err = s.postRepo.Delete(postID)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	return nil
}

func (s *PostService) GetUserPosts(

	targetUserID uint,
	currentUserID *uint,
	limit int,
	cursor paginator.Cursor,
) (*responses.PostResponse, error) {
	// Check if current user can view target user's posts
	var privacy postgres.PostPrivacy
	if currentUserID != nil && *currentUserID != targetUserID {
		// Check if they are friends for private posts
		isFriend, err := s.userRepo.IsFriend(*currentUserID, targetUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to check friendship: %w", err)
		}

		if !isFriend {
			// Only return public posts
			privacy = postgres.PostPrivacyPublic
		}
	}

	posts, err := s.postRepo.GetUserPosts(targetUserID, privacy)
	if err != nil {
		return nil, fmt.Errorf("failed to get user posts: %w", err)
	}

	return &responses.PostResponse{
		Posts: posts,
	}, nil
}

func (s *PostService) GetFeed(userID uint, limit int, cursor string) (*responses.PostResponse, error) {
	cursorObj := paginator.Cursor{}
	if cursor != "" {
		// Parse cursor if needed
	}

	posts, next, err := s.postRepo.GetFeed(userID, cursorObj, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed: %w", err)
	}

	return &responses.PostResponse{
		Posts:      posts,
		NextCursor: &next,
	}, nil
}

func (s *PostService) GetPublicFeed(limit int, cursor string) (*responses.PostResponse, error) {
	cursorObj := paginator.Cursor{}
	if cursor != "" {
		// Parse cursor if needed
	}

	result, next, err := s.postRepo.GetPublicFeed(cursorObj, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get public feed: %w", err)
	}

	return &responses.PostResponse{
		Posts:      result,
		NextCursor: &next,
	}, nil
}

func (s *PostService) SearchPosts(query string, userID *uint, limit int, cursor string) (*responses.PostResponse, error) {
	cursorObj := paginator.Cursor{}
	if cursor != "" {
		// Parse cursor if needed
	}

	result, next, err := s.postRepo.SearchPosts(query, cursorObj, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts: %w", err)
	}

	return &responses.PostResponse{
		Posts:      result,
		NextCursor: &next,
	}, nil
}

func (s *PostService) ToggleLike(postID, userID uint, likeType string) (bool, error) {
	// Check if post exists
	_, err := s.postRepo.GetByID(postID)
	if err != nil {
		return false, fmt.Errorf("post not found")
	}

	// Check if user already liked
	hasLiked, err := s.likeRepo.HasUserLiked(userID, postID)
	if err != nil {
		return false, fmt.Errorf("failed to check like status: %w", err)
	}

	if hasLiked {
		// Unlike
		err = s.likeRepo.Delete(userID, postID)
		if err != nil {
			return false, fmt.Errorf("failed to delete like: %w", err)
		}

		// Decrement like count
		err = s.postRepo.DecrementLikeCount(postID)
		if err != nil {
			return false, fmt.Errorf("failed to decrement like count: %w", err)
		}

		return false, nil
	} else {
		// Like
		if likeType == "" {
			likeType = "like"
		}

		like := &postgres.Like{
			PostID:    postID,
			UserID:    userID,
			Type:      likeType,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = s.likeRepo.Create(like)
		if err != nil {
			return false, fmt.Errorf("failed to create like: %w", err)
		}

		// Increment like count
		err = s.postRepo.IncrementLikeCount(postID)
		if err != nil {
			return false, fmt.Errorf("failed to increment like count: %w", err)
		}

		return true, nil
	}
}

func (s *PostService) CreateComment(postID, userID uint, req *requests.CreateCommentRequest) (*postgres.Comment, error) {
	// Check if post exists
	_, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, fmt.Errorf("post not found")
	}

	// Create comment
	comment := &postgres.Comment{
		PostID:    postID,
		UserID:    userID,
		ParentID:  req.ParentID,
		Content:   req.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.commentRepo.Create(comment)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// Increment comment count
	err = s.postRepo.IncrementCommentCount(postID)
	if err != nil {
		return nil, fmt.Errorf("failed to increment comment count: %w", err)
	}

	return s.commentRepo.GetByID(comment.ID)
}

func (s *PostService) GetComments(postID uint, req *requests.GetPostCommentsRequest) (*responses.CommentResponse, error) {
	// Check if post exists
	_, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, fmt.Errorf("post not found")
	}

	cursor := paginator.Cursor{
		Before: &req.Before,
		After:  &req.After,
	}

	comments, next, err := s.commentRepo.GetByPostID(postID, cursor, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	return &responses.CommentResponse{
		Comments:   comments,
		NextCursor: &next,
	}, nil
}

func (s *PostService) SharePost(postID, userID uint, req *requests.SharePostRequest) (*postgres.Share, error) {
	// Check if post exists
	_, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, fmt.Errorf("post not found")
	}

	// Create share
	share := &postgres.Share{
		PostID:    postID,
		UserID:    userID,
		Content:   req.Content,
		Privacy:   req.Privacy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set default privacy
	if share.Privacy == "" {
		share.Privacy = postgres.PostPrivacyPublic
	}

	err = s.shareRepo.Create(share)
	if err != nil {
		return nil, fmt.Errorf("failed to create share: %w", err)
	}

	// Increment share count
	err = s.postRepo.IncrementShareCount(postID)
	if err != nil {
		return nil, fmt.Errorf("failed to increment share count: %w", err)
	}

	return s.shareRepo.GetByID(share.ID)
}

func (s *PostService) UpdateComment(commentID, userID uint, req *requests.UpdateCommentRequest) (*postgres.Comment, error) {
	// Get existing comment
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return nil, fmt.Errorf("comment not found: %w", err)
	}

	// Check ownership
	if comment.UserID != userID {
		return nil, fmt.Errorf("permission denied: only author can update comment")
	}

	// Update comment
	updates := map[string]interface{}{
		"content":    req.Content,
		"is_edited":  true,
		"edited_at":  time.Now(),
		"updated_at": time.Now(),
	}

	err = s.commentRepo.Update(commentID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return s.commentRepo.GetByID(commentID)
}

func (s *PostService) DeleteComment(commentID, userID uint) error {
	// Get comment to check ownership
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return fmt.Errorf("comment not found: %w", err)
	}

	// Check ownership
	if comment.UserID != userID {
		return fmt.Errorf("permission denied: only author can delete comment")
	}

	err = s.commentRepo.Delete(commentID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}

func (s *PostService) SavePost(postID, userID uint, category string) error {
	// Check if post exists
	_, err := s.postRepo.GetByID(postID)
	if err != nil {
		return fmt.Errorf("post not found: %w", err)
	}

	// Check if already saved
	isSaved, err := s.postRepo.IsPostSaved(postID, userID)
	if err != nil {
		return fmt.Errorf("failed to check saved status: %w", err)
	}

	if isSaved {
		return fmt.Errorf("post is already saved")
	}

	// Save post
	savedPost := &postgres.SavedPost{
		PostID:    postID,
		UserID:    userID,
		Category:  category,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.postRepo.SavePost(savedPost)
	if err != nil {
		return fmt.Errorf("failed to save post: %w", err)
	}

	return nil
}

func (s *PostService) UnsavePost(postID, userID uint) error {
	err := s.postRepo.UnsavePost(postID, userID)
	if err != nil {
		return fmt.Errorf("failed to unsave post: %w", err)
	}
	return nil
}

func (s *PostService) GetSavedPosts(userID uint) ([]postgres.Post, error) {
	posts, err := s.postRepo.GetSavedPosts(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get saved posts: %w", err)
	}
	return posts, nil
}

func (s *PostService) ReportPost(postID, userID uint, reason, details string) error {
	// Check if post exists
	_, err := s.postRepo.GetByID(postID)
	if err != nil {
		return fmt.Errorf("post not found: %w", err)
	}

	// Create report
	report := &postgres.PostReport{
		PostID:    postID,
		UserID:    userID,
		Reason:    reason,
		Details:   details,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.postRepo.CreateReport(report)
	if err != nil {
		return fmt.Errorf("failed to create report: %w", err)
	}

	return nil
}

func (s *PostService) GetUserStats(userID uint) (*responses.PostStats, error) {
	stats, err := s.postRepo.GetUserStats(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}
	return stats, nil
}

func (s *PostService) RecordView(postID uint, userID *uint, ipAddress, userAgent string) error {
	view := &postgres.PostView{
		PostID:    postID,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	}

	err := s.postRepo.RecordView(view)
	if err != nil {
		return fmt.Errorf("failed to record view: %w", err)
	}

	// Increment view count
	err = s.postRepo.IncrementViewCount(postID)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	return nil
}

func (s *PostService) CheckPostVisibility(postID uint, currentUserID *uint) (bool, error) {
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return false, fmt.Errorf("post not found: %w", err)
	}

	// Public posts are visible to everyone
	if post.Privacy == postgres.PostPrivacyPublic {
		return true, nil
	}

	// Private posts are only visible to author
	if post.Privacy == postgres.PostPrivacyPrivate {
		return currentUserID != nil && *currentUserID == post.AuthorID, nil
	}

	// Friends-only posts
	if post.Privacy == postgres.PostPrivacyFriends {
		if currentUserID == nil {
			return false, nil
		}

		if *currentUserID == post.AuthorID {
			return true, nil
		}

		// Check friendship
		isFriend, err := s.userRepo.IsFriend(*currentUserID, post.AuthorID)
		if err != nil {
			return false, fmt.Errorf("failed to check friendship: %w", err)
		}
		return isFriend, nil
	}

	return false, nil
}
