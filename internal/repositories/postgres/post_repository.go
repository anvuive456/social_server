package postgres

import (
	"fmt"
	"strings"
	"time"

	"social_server/internal/models/paginators"
	"social_server/internal/models/postgres"
	"social_server/internal/models/responses"
	"social_server/internal/repositories"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"gorm.io/gorm"
)

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) repositories.PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(post *postgres.Post) error {
	return r.db.Create(post).Error
}

func (r *postRepository) GetByID(id uint) (*postgres.Post, error) {
	var post postgres.Post
	err := r.db.
		Preload("Author").
		Preload("Author.Profile").
		Preload("Media").
		Preload("Likes").
		Preload("Comments").
		Preload("Shares").
		First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&postgres.Post{}).Where("id = ?", id).Updates(updates).Error
}

func (r *postRepository) Delete(id uint) error {
	return r.db.Delete(&postgres.Post{}, id).Error
}

func (r *postRepository) GetUserPosts(currUserID uint, targetUserID *uint, privacy postgres.PostPrivacy, cursor paginator.Cursor, limit int) ([]postgres.Post, paginator.Cursor, error) {
	var posts []postgres.Post
	query := r.db.Model(&postgres.Post{})

	if targetUserID != nil {
		// Lấy posts của user cụ thể
		query = query.Where("author_id = ?", *targetUserID)
	} else {
		// Lấy posts của bạn bè và posts của chính currUserID
		query = query.
			Joins("LEFT JOIN user_friends ON user_friends.friend_id = posts.author_id").
			Where("(user_friends.user_id = ? AND user_friends.status = ?) OR posts.author_id = ?",
				currUserID, "accepted", currUserID)
	}

	query = query.Where("privacy = ?", privacy).
		Preload("Author").
		Preload("Author.Profile").
		Preload("Media")
	order := paginator.DESC
	p := paginators.CreatePostPaginator(cursor, &order, &limit)
	result, nextCursor, err := p.Paginate(query, &posts)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}
	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}
	return posts, nextCursor, err
}

func (r *postRepository) GetFeed(userID uint, cursor paginator.Cursor, limit int) ([]postgres.Post, paginator.Cursor, error) {
	var posts []postgres.Post
	query := r.db.
		Joins("LEFT JOIN friendships ON friendships.friend_id = posts.author_id").
		Where("(friendships.user_id = ? AND friendships.status = ?) OR posts.privacy = ? OR posts.author_id = ?",
			userID, "accepted", "public", userID).
		Preload("Author").
		Preload("Media")

	order := paginator.DESC
	p := paginators.CreatePostPaginator(cursor, &order, &limit)
	result, nextCursor, err := p.Paginate(query, &posts)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}

	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return posts, nextCursor, nil
}

func (r *postRepository) GetPublicFeed(cursor paginator.Cursor, limit int) ([]postgres.Post, paginator.Cursor, error) {
	var posts []postgres.Post
	query := r.db.
		Where("privacy = ?", "public").
		Preload("Author").
		Preload("Media")
	order := paginator.DESC
	p := paginators.CreatePostPaginator(cursor, &order, &limit)
	result, nextCursor, err := p.Paginate(query, &posts)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}

	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return posts, nextCursor, nil
}

func (r *postRepository) IncrementLikeCount(postID uint) error {
	return r.db.Model(&postgres.Post{}).
		Where("id = ?", postID).
		UpdateColumn("likes_count", gorm.Expr("likes_count + 1")).Error
}

func (r *postRepository) DecrementLikeCount(postID uint) error {
	return r.db.Model(&postgres.Post{}).
		Where("id = ?", postID).
		UpdateColumn("likes_count", gorm.Expr("likes_count - 1")).Error
}

func (r *postRepository) IncrementCommentCount(postID uint) error {
	return r.db.Model(&postgres.Post{}).
		Where("id = ?", postID).
		UpdateColumn("comments_count", gorm.Expr("comments_count + 1")).Error
}

func (r *postRepository) DecrementCommentCount(postID uint) error {
	return r.db.Model(&postgres.Post{}).
		Where("id = ?", postID).
		UpdateColumn("comments_count", gorm.Expr("comments_count - 1")).Error
}

func (r *postRepository) IncrementShareCount(postID uint) error {
	return r.db.Model(&postgres.Post{}).
		Where("id = ?", postID).
		UpdateColumn("shares_count", gorm.Expr("shares_count + 1")).Error
}

func (r *postRepository) SearchPosts(query string, cursor paginator.Cursor, limit int) ([]postgres.Post, paginator.Cursor, error) {
	var posts []postgres.Post
	searchQuery := "%" + strings.ToLower(query) + "%"

	dbQuery := r.db.
		Where("LOWER(content) LIKE ? OR LOWER(tags) LIKE ?", searchQuery, searchQuery).
		Where("privacy = ?", "public").
		Preload("Author").
		Preload("Media")

	order := paginator.DESC
	p := paginators.CreatePostPaginator(cursor, &order, &limit)
	result, nextCursor, err := p.Paginate(dbQuery, &posts)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}

	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return posts, nextCursor, nil
}

func (r *postRepository) GetPostsByTag(tag string, cursor paginator.Cursor, limit int) ([]postgres.Post, paginator.Cursor, error) {
	var posts []postgres.Post
	tagQuery := "%" + strings.ToLower(tag) + "%"

	dbQuery := r.db.
		Where("LOWER(tags) LIKE ?", tagQuery).
		Where("privacy = ?", "public").
		Preload("Author").
		Preload("Media")

	order := paginator.DESC
	p := paginators.CreatePostPaginator(cursor, &order, &limit)
	result, nextCursor, err := p.Paginate(dbQuery, &posts)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}

	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return posts, nextCursor, nil
}

func (r *postRepository) CreateMedia(media *postgres.PostMedia) error {
	return r.db.Create(media).Error
}

func (r *postRepository) IsPostSaved(postID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&postgres.SavedPost{}).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *postRepository) SavePost(savedPost *postgres.SavedPost) error {
	return r.db.Create(savedPost).Error
}

func (r *postRepository) UnsavePost(postID, userID uint) error {
	return r.db.
		Where("post_id = ? AND user_id = ?", postID, userID).
		Delete(&postgres.SavedPost{}).Error
}

func (r *postRepository) GetSavedPosts(userID uint) ([]postgres.Post, error) {
	var posts []postgres.Post
	err := r.db.
		Joins("JOIN saved_posts ON saved_posts.post_id = posts.id").
		Where("saved_posts.user_id = ?", userID).
		Preload("Author").
		Preload("Media").
		Order("saved_posts.created_at DESC").
		Find(&posts).Error
	return posts, err
}

func (r *postRepository) CreateReport(report *postgres.PostReport) error {
	return r.db.Create(report).Error
}

func (r *postRepository) GetUserStats(userID uint) (*responses.PostStats, error) {
	var stats responses.PostStats

	// Count total posts
	err := r.db.Model(&postgres.Post{}).
		Where("author_id = ?", userID).
		Count((*int64)(&stats.TotalPosts)).Error
	if err != nil {
		return nil, err
	}

	// Count total likes received
	err = r.db.Model(&postgres.Like{}).
		Joins("JOIN posts ON posts.id = likes.post_id").
		Where("posts.author_id = ? AND likes.target_type = ?", userID, "post").
		Count((*int64)(&stats.TotalLikes)).Error
	if err != nil {
		return nil, err
	}

	// Count total comments received
	err = r.db.Model(&postgres.Comment{}).
		Joins("JOIN posts ON posts.id = comments.post_id").
		Where("posts.author_id = ?", userID).
		Count((*int64)(&stats.TotalComments)).Error
	if err != nil {
		return nil, err
	}

	// Count total shares received
	err = r.db.Model(&postgres.Share{}).
		Joins("JOIN posts ON posts.id = shares.post_id").
		Where("posts.author_id = ?", userID).
		Count((*int64)(&stats.TotalShares)).Error
	if err != nil {
		return nil, err
	}

	// Count total views received
	err = r.db.Model(&postgres.PostView{}).
		Joins("JOIN posts ON posts.id = post_views.post_id").
		Where("posts.author_id = ?", userID).
		Count((*int64)(&stats.TotalViews)).Error
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *postRepository) RecordView(view *postgres.PostView) error {
	return r.db.Create(view).Error
}

func (r *postRepository) IncrementViewCount(postID uint) error {
	return r.db.Model(&postgres.Post{}).
		Where("id = ?", postID).
		UpdateColumn("views_count", gorm.Expr("views_count + 1")).Error
}

func (r *postRepository) Search(query string, userID *uint, limit int) ([]postgres.Post, error) {
	var posts []postgres.Post
	searchQuery := "%" + strings.ToLower(query) + "%"

	dbQuery := r.db.
		Where("LOWER(content) LIKE ? OR LOWER(tags) LIKE ?", searchQuery, searchQuery).
		Preload("Author").
		Preload("Media").
		Order("created_at DESC").
		Limit(limit)

	if userID == nil {
		dbQuery = dbQuery.Where("privacy = ?", "public")
	} else {
		dbQuery = dbQuery.Where("privacy = ? OR author_id = ?", "public", *userID)
	}

	err := dbQuery.Find(&posts).Error
	return posts, err
}

func (r *postRepository) GetTrending(limit int) ([]postgres.Post, error) {
	var posts []postgres.Post
	err := r.db.
		Where("privacy = ?", "public").
		Where("created_at > ?", time.Now().AddDate(0, 0, -7)). // Last 7 days
		Preload("Author").
		Preload("Media").
		Order("(likes_count + comments_count + shares_count + views_count) DESC").
		Limit(limit).
		Find(&posts).Error
	return posts, err
}

// Comment Repository
type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) repositories.CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(comment *postgres.Comment) error {
	return r.db.Create(comment).Error
}

func (r *commentRepository) GetByID(id uint) (*postgres.Comment, error) {
	var comment postgres.Comment
	err := r.db.
		Preload("Author").
		Preload("Replies").
		First(&comment, id).Error
	return &comment, err
}

func (r *commentRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&postgres.Comment{}).Where("id = ?", id).Updates(updates).Error
}

func (r *commentRepository) Delete(id uint) error {
	return r.db.Delete(&postgres.Comment{}, id).Error
}

func (r *commentRepository) GetByPostID(postID uint, cursor paginator.Cursor, limit int) ([]postgres.Comment, paginator.Cursor, error) {
	var comments []postgres.Comment
	query := r.db.
		Where("post_id = ? AND parent_id IS NULL", postID).
		Preload("Author").
		Preload("Replies")

	order := paginator.DESC
	p := paginators.CreateCommentPaginator(cursor, &order, &limit)
	result, nextCursor, err := p.Paginate(query, comments)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}
	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return comments, nextCursor, nil
}

func (r *commentRepository) GetReplies(parentID uint, cursor paginator.Cursor, limit int) ([]postgres.Comment, paginator.Cursor, error) {
	var comments []postgres.Comment
	query := r.db.
		Where("parent_id = ?", parentID).
		Preload("Author")
	order := paginator.DESC
	p := paginators.CreateCommentPaginator(cursor, &order, &limit)
	result, nextCursor, err := p.Paginate(query, comments)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}
	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return comments, nextCursor, nil
}

func (r *commentRepository) IncrementLikeCount(commentID uint) error {
	return r.db.Model(&postgres.Comment{}).
		Where("id = ?", commentID).
		UpdateColumn("likes_count", gorm.Expr("likes_count + 1")).Error
}

func (r *commentRepository) DecrementLikeCount(commentID uint) error {
	return r.db.Model(&postgres.Comment{}).
		Where("id = ?", commentID).
		UpdateColumn("likes_count", gorm.Expr("likes_count - 1")).Error
}

func (r *commentRepository) GetCommentCount(postID uint) (int64, error) {
	var count int64
	err := r.db.Model(&postgres.Comment{}).
		Where("post_id = ?", postID).
		Count(&count).Error
	return count, err
}

// Like Repository
type likeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) repositories.LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) Create(like *postgres.Like) error {
	return r.db.Create(like).Error
}

func (r *likeRepository) Delete(userID, postID uint) error {
	return r.db.
		Where("user_id = ? AND post_id = ?", userID, postID).
		Delete(&postgres.Like{}).Error
}

func (r *likeRepository) GetByUserAndTarget(userID, targetID uint, targetType string) (*postgres.Like, error) {
	var like postgres.Like
	var err error

	if targetType == "post" {
		err = r.db.
			Where("user_id = ? AND post_id = ?", userID, targetID).
			First(&like).Error
	} else if targetType == "comment" {
		err = r.db.
			Where("user_id = ? AND comment_id = ?", userID, targetID).
			First(&like).Error
	}

	return &like, err
}

func (r *likeRepository) GetLikeCount(targetID uint, targetType string) (int64, error) {
	var count int64

	if targetType == "post" {
		err := r.db.Model(&postgres.Like{}).
			Where("post_id = ?", targetID).
			Count(&count).Error
		return count, err
	} else if targetType == "comment" {
		err := r.db.Model(&postgres.Like{}).
			Where("comment_id = ?", targetID).
			Count(&count).Error
		return count, err
	}

	return 0, fmt.Errorf("unsupported target type: %s", targetType)
}

func (r *likeRepository) GetUserLikes(userID uint, targetType string, cursor paginator.Cursor, limit int) ([]postgres.Like, paginator.Cursor, error) {
	var likes []postgres.Like
	query := r.db.
		Where("user_id = ?", userID).
		Preload("User")
	order := paginator.DESC
	p := paginators.CreateLikePaginator(
		cursor,
		&order,
		&limit,
	)
	result, nextCursor, err := p.Paginate(query, &likes)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}

	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return likes, nextCursor, nil

}

func (r *likeRepository) HasUserLiked(userID, postID uint) (bool, error) {
	var count int64
	err := r.db.Model(&postgres.Like{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Count(&count).Error
	return count > 0, err
}

// Share Repository
type shareRepository struct {
	db *gorm.DB
}

func NewShareRepository(db *gorm.DB) repositories.ShareRepository {
	return &shareRepository{db: db}
}

func (r *shareRepository) Create(share *postgres.Share) error {
	return r.db.Create(share).Error
}

func (r *shareRepository) GetByID(id uint) (*postgres.Share, error) {
	var share postgres.Share
	err := r.db.
		Preload("User").
		Preload("Post").
		First(&share, id).Error
	return &share, err
}

func (r *shareRepository) GetByPostID(postID uint) ([]postgres.Share, error) {
	var shares []postgres.Share
	err := r.db.
		Where("post_id = ?", postID).
		Preload("User").
		Order("created_at DESC").
		Find(&shares).Error
	return shares, err
}

func (r *shareRepository) Delete(id uint) error {
	return r.db.Delete(&postgres.Share{}, id).Error
}

func (r *shareRepository) GetUserShares(userID uint, cursor paginator.Cursor, limit int) ([]postgres.Share, paginator.Cursor, error) {
	var shares []postgres.Share
	query := r.db.
		Where("user_id = ?", userID).
		Preload("Post")

	order := paginator.DESC
	p := paginators.CreateSharePaginator(
		cursor,
		&order,
		&limit,
	)

	result, nextCursor, err := p.Paginate(query, &shares)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}
	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return shares, nextCursor, nil
}

func (r *shareRepository) GetPostShares(postID uint, cursor paginator.Cursor, limit int) ([]postgres.Share, paginator.Cursor, error) {
	var shares []postgres.Share
	query := r.db.
		Where("post_id = ?", postID).
		Preload("User").
		Order("created_at DESC").
		Limit(limit + 1)
	order := paginator.DESC
	p := paginators.CreateSharePaginator(
		cursor,
		&order,
		&limit,
	)

	result, nextCursor, err := p.Paginate(query, &shares)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}
	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return shares, nextCursor, nil
}
