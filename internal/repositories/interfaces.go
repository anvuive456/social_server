package repositories

import (
	"context"
	"social_server/internal/models/postgres"
	"social_server/internal/models/requests"
	"social_server/internal/models/responses"
	"time"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)

type UserRepository interface {
	Create(user *postgres.User) (createdUser *postgres.User, err error)
	GetByID(id uint) (*postgres.User, error)
	GetByEmail(email string) (*postgres.User, error)
	GetByUsername(username string) (*postgres.User, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error
	UpdateOnlineStatus(id uint, isOnline bool) error
	SearchUsers(currentUserID uint, query string, cursor paginator.Cursor, limit int) ([]responses.UserSearchSimpleResponse, paginator.Cursor, error)
	GetUserProfile(id uint) (*responses.UserProfile, error)
	GetBlockedUsers(userID uint, cursor paginator.Cursor, limit int) ([]responses.UserProfile, paginator.Cursor, error)
	RemoveFriend(userID, friendID uint) error

	BlockUser(userID, targetID uint) error
	UnblockUser(userID, targetID uint) error
	IsFriend(userID, targetID uint) (bool, error)
	AreFriends(userID, targetID uint) (bool, error)
	IsBlocked(userID, targetID uint) (bool, error)
}

type FriendRepository interface {
	GetFriends(userID uint, cursor paginator.Cursor, limit int, search string) ([]postgres.UserFriend, paginator.Cursor, uint, error)

	SendFriendRequest(fromID, toID uint, message string) error
	HasAlreadyFriendRequest(userID, targetID uint) (bool, error)
	GetFriendRequests(userID uint, requestType requests.GetFriendRequestType, cursor paginator.Cursor, limit int) ([]postgres.FriendRequest, paginator.Cursor, uint, error)
	GetFriendRequestStats(userID uint) (uint, uint, error)
	AcceptFriendRequest(userID, fromID uint) error
	DeclineFriendRequest(userID, fromID uint) error
	AddFriend(userID, friendID uint) error
}

type ProfileRepository interface {
	Create(profile *postgres.Profile) (*postgres.Profile, error)
	GetByID(id uint) (*postgres.Profile, error)
	GetByUserID(userID uint) (*postgres.Profile, error)
	Update(profile *postgres.Profile) (*postgres.Profile, error)
	UpdatePartial(id uint, updates map[string]interface{}) (*postgres.Profile, error)
	CreateOrUpdate(profile *postgres.Profile) (*postgres.Profile, error)
	Delete(id uint) error
}

type PostRepository interface {
	Create(post *postgres.Post) error
	GetByID(id uint) (*postgres.Post, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error
	GetUserPosts(targetUserID uint, privacy postgres.PostPrivacy) ([]postgres.Post, error)
	GetFeed(userID uint, cursor paginator.Cursor, limit int) ([]postgres.Post, paginator.Cursor, error)
	GetPublicFeed(cursor paginator.Cursor, limit int) ([]postgres.Post, paginator.Cursor, error)
	IncrementLikeCount(postID uint) error
	DecrementLikeCount(postID uint) error
	IncrementCommentCount(postID uint) error
	DecrementCommentCount(postID uint) error
	IncrementShareCount(postID uint) error
	SearchPosts(query string, cursor paginator.Cursor, limit int) ([]postgres.Post, paginator.Cursor, error)
	GetPostsByTag(tag string, cursor paginator.Cursor, limit int) ([]postgres.Post, paginator.Cursor, error)
	CreateMedia(media *postgres.PostMedia) error
	IsPostSaved(postID, userID uint) (bool, error)
	SavePost(savedPost *postgres.SavedPost) error
	UnsavePost(postID, userID uint) error
	GetSavedPosts(userID uint) ([]postgres.Post, error)
	CreateReport(report *postgres.PostReport) error
	GetUserStats(userID uint) (*responses.PostStats, error)
	RecordView(view *postgres.PostView) error
	IncrementViewCount(postID uint) error
	Search(query string, userID *uint, limit int) ([]postgres.Post, error)
	GetTrending(limit int) ([]postgres.Post, error)
}

type CommentRepository interface {
	Create(comment *postgres.Comment) error
	GetByID(id uint) (*postgres.Comment, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error
	GetByPostID(postID uint, cursor paginator.Cursor, limit int) ([]postgres.Comment, paginator.Cursor, error)
	GetReplies(parentID uint, cursor paginator.Cursor, limit int) ([]postgres.Comment, paginator.Cursor, error)
	IncrementLikeCount(commentID uint) error
	DecrementLikeCount(commentID uint) error
	GetCommentCount(postID uint) (int64, error)
}

type LikeRepository interface {
	Create(like *postgres.Like) error
	Delete(userID, postID uint) error
	GetByUserAndTarget(userID, targetID uint, targetType string) (*postgres.Like, error)
	GetLikeCount(targetID uint, targetType string) (int64, error)
	GetUserLikes(userID uint, targetType string, cursor paginator.Cursor, limit int) ([]postgres.Like, paginator.Cursor, error)
	HasUserLiked(userID, postID uint) (bool, error)
}

type ShareRepository interface {
	Create(share *postgres.Share) error
	GetByID(id uint) (*postgres.Share, error)
	GetByPostID(postID uint) ([]postgres.Share, error)
	Delete(id uint) error
	GetUserShares(userID uint, cursor paginator.Cursor, limit int) ([]postgres.Share, paginator.Cursor, error)
	GetPostShares(postID uint, cursor paginator.Cursor, limit int) ([]postgres.Share, paginator.Cursor, error)
}

type ChatRoomRepository interface {
	Create(room *postgres.ChatRoom) error
	GetByID(id uint) (*postgres.ChatRoom, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error
	GetUserRooms(userID uint) ([]responses.ChatRoomSummary, error)
	GetPrivateRoom(userID1, userID2 uint) (*postgres.ChatRoom, error)
	AddParticipant(participant *postgres.Participant) error
	RemoveParticipant(roomID, userID uint) error
	GetParticipants(roomID uint) ([]postgres.User, error)
	UpdateLastActivity(roomID uint, lastActivity time.Time) error
	ArchiveRoom(roomID uint) error
	SearchRooms(userID uint, query string, limit int) ([]responses.ChatRoomSummary, error)
}

type MessageRepository interface {
	Create(message *postgres.Message) error
	GetByID(id uint) (*postgres.Message, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error
	GetRoomMessages(roomID uint, cursor paginator.Cursor, limit int) ([]postgres.Message, paginator.Cursor, error)
	MarkAsRead(messageID, userID uint) error
	MarkAsDelivered(messageID, userID uint) error
	GetUnreadCount(roomID, userID uint) (int, error)
	SearchMessages(roomID uint, query string, limit int) ([]responses.MessageSearchResult, error)
	GetMessagesByType(roomID uint, messageType postgres.MessageType, cursor paginator.Cursor, limit int) ([]postgres.Message, paginator.Cursor, error)
	GetRecentMedia(roomID uint, mediaType postgres.MessageType, limit int) ([]postgres.Message, error)
	AddReaction(messageID, userID uint, emoji string) error
	RemoveReaction(messageID, userID uint, emoji string) error
}

type ParticipantRepository interface {
	Create(participant *postgres.Participant) error
	GetByID(id uint) (*postgres.Participant, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error
	GetByRoomAndUser(roomID, userID uint) (*postgres.Participant, error)
	GetRoomParticipants(roomID uint) ([]postgres.Participant, error)
	GetUserParticipations(userID uint) ([]postgres.Participant, error)
	UpdateRole(roomID, userID uint, role postgres.ParticipantRole) error
	UpdateLastRead(roomID, userID uint) error
	MuteParticipant(roomID, userID uint, isMuted bool) error
	BlockParticipant(roomID, userID uint, isBlocked bool) error
}

type TypingIndicatorRepository interface {
	SetTyping(roomID, userID uint, isTyping bool) error
	GetTypingUsers(roomID uint) ([]uint, error)
	ClearTyping(roomID, userID uint) error
	ClearExpiredTyping(ctx context.Context) error
}

type OnlineStatusRepository interface {
	SetOnline(userID uint, isOnline bool) error
	GetStatus(userID uint) (*postgres.OnlineStatus, error)
	GetMultipleStatus(userIDs []uint) ([]postgres.OnlineStatus, error)
	UpdateLastSeen(userID uint) error
	SetCustomStatus(userID uint, status string) error
}

type ChatInviteRepository interface {
	Create(invite *postgres.ChatInvite) error
	GetByID(id uint) (*postgres.ChatInvite, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error
	GetUserInvites(userID uint, status string, cursor paginator.Cursor, limit int) ([]postgres.ChatInvite, paginator.Cursor, error)
	GetRoomInvites(roomID uint, cursor paginator.Cursor, limit int) ([]postgres.ChatInvite, paginator.Cursor, error)
	AcceptInvite(inviteID uint) error
	DeclineInvite(inviteID uint) error
	ExpireOldInvites(ctx context.Context) error
}

type ChatNotificationRepository interface {
	Create(notification *postgres.ChatNotification) error
	GetByID(id uint) (*postgres.ChatNotification, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error
	GetUserNotifications(userID uint, cursor paginator.Cursor, limit int) ([]postgres.ChatNotification, paginator.Cursor, error)
	MarkAsRead(notificationID uint) error
	MarkAllAsRead(userID uint) error
	GetUnreadCount(userID uint) (int64, error)
	DeleteOldNotifications(olderThan time.Time) error
}

type AuthRepository interface {
	CreateRefreshToken(token *postgres.RefreshToken) error
	GetRefreshToken(token string) (*postgres.RefreshToken, error)
	RevokeRefreshToken(token string) error
	CleanupExpiredTokens(ctx context.Context) error
	CreatePasswordReset(reset *postgres.PasswordReset) error
	GetPasswordReset(token string) (*postgres.PasswordReset, error)
	UsePasswordReset(token string) error
	CreateEmailVerification(verification *postgres.EmailVerification) error
	GetEmailVerification(token string) (*postgres.EmailVerification, error)
	VerifyEmail(token string) error
	LogLoginAttempt(attempt *postgres.LoginAttempt) error
	GetLoginAttempts(email, ipAddress string, since time.Time) ([]postgres.LoginAttempt, error)
	CreateSecurityEvent(event *postgres.SecurityEvent) error
	CheckRateLimit(key string, windowStart time.Time, limit int) (bool, error)
	IncrementRateLimit(key string, windowStart, expiresAt time.Time) error
}

type CallRepository interface {
	Create(call *postgres.Call) error
	GetByID(id uint) (*postgres.Call, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error
	GetUserCalls(userID uint, cursor paginator.Cursor, limit int) ([]postgres.Call, paginator.Cursor, error)
	GetUserCallHistory(userID uint, cursor paginator.Cursor, limit int) ([]postgres.Call, paginator.Cursor, error)
	GetActiveCall(userID uint) (*postgres.Call, error)
	EndCall(callID uint) error
	JoinCall(callID, userID uint) error
	LeaveCall(callID, userID uint) error
	GetCallParticipants(callID uint) ([]postgres.CallParticipant, error)
}

type SessionRepository interface {
	// Session management
	CreateSession(session *postgres.Session) error
	GetSessionByTokenID(tokenID string) (*postgres.Session, error)
	GetSessionBySessionID(sessionID string) (*postgres.Session, error)
	GetUserSessions(userID uint) ([]*postgres.Session, error)
	UpdateSession(session *postgres.Session) error
	DeactivateSession(tokenID string) error
	DeactivateUserSessions(userID uint) error
	DeleteExpiredSessions() error

	// Token blacklist management
	BlacklistToken(tokenID string, userID uint, reason string, expiresAt time.Time) error
	IsTokenBlacklisted(tokenID string) (bool, error)
	CleanupExpiredBlacklist() error

	// Security logging
	LogSecurityEvent(log *postgres.SecurityLog) error
	GetUserSecurityLogs(userID uint, limit int) ([]*postgres.SecurityLog, error)
	GetHighRiskEvents(since time.Time) ([]*postgres.SecurityLog, error)

	// Security analysis
	GetSuspiciousActivities(userID uint, since time.Time) ([]*postgres.SecurityLog, error)
	CountFailedLogins(userID uint, since time.Time) (int64, error)
	GetActiveSessionsCount(userID uint) (int64, error)
}

type Repositories struct {
	User             UserRepository
	Friend           FriendRepository
	Profile          ProfileRepository
	Post             PostRepository
	Comment          CommentRepository
	Like             LikeRepository
	Share            ShareRepository
	ChatRoom         ChatRoomRepository
	Message          MessageRepository
	Participant      ParticipantRepository
	TypingIndicator  TypingIndicatorRepository
	OnlineStatus     OnlineStatusRepository
	ChatInvite       ChatInviteRepository
	ChatNotification ChatNotificationRepository
	Auth             AuthRepository
	Call             CallRepository
	Session          SessionRepository
}
