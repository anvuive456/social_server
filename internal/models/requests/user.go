package requests

type SendFriendRequestRequest struct {
	ReceiverID uint   `json:"receiver_id" binding:"required"`
	Message    string `json:"message,omitempty"`
}

type RespondFriendRequestRequest struct {
	Status string `json:"status" binding:"required,oneof=accepted rejected"`
}

type UserSearchRequest struct {
	Search string `form:"search,required" json:"search,required"`
	Before string `form:"before,omitempty" json:"before,omitempty"`
	After  string `form:"after,omitempty" json:"after,omitempty"`
	Limit  int    `form:"limit,required" json:"limit,required"`
}

type UpdateUserSettingsRequest struct {
	PrivacyProfileVisibility    *string `json:"privacy_profile_visibility,omitempty"`
	PrivacyShowOnlineStatus     *bool   `json:"privacy_show_online_status,omitempty"`
	PrivacyAllowFriendRequests  *bool   `json:"privacy_allow_friend_requests,omitempty"`
	NotificationsEmail          *bool   `json:"notifications_email,omitempty"`
	NotificationsPush           *bool   `json:"notifications_push,omitempty"`
	NotificationsFriendRequests *bool   `json:"notifications_friend_requests,omitempty"`
	NotificationsMessages       *bool   `json:"notifications_messages,omitempty"`
	NotificationsPosts          *bool   `json:"notifications_posts,omitempty"`
}

type BlockUserRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Reason string `json:"reason,omitempty"`
}

type ReportUserRequest struct {
	UserID  uint   `json:"user_id" binding:"required"`
	Reason  string `json:"reason" binding:"required"`
	Details string `json:"details,omitempty"`
}

type UpdateOnlineStatusRequest struct {
	Status     string `json:"status" binding:"required,oneof=online away busy invisible"`
	DeviceInfo string `json:"device_info,omitempty"`
}

type GetUserListRequest struct {
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
	SortBy   string `json:"sort_by,omitempty"`
	Order    string `json:"order,omitempty"`
}

type ChangeEmailRequest struct {
	NewEmail        string `json:"new_email" binding:"required,email"`
	CurrentPassword string `json:"current_password" binding:"required"`
}

type DeactivateAccountRequest struct {
	Password string `json:"password" binding:"required"`
	Reason   string `json:"reason,omitempty"`
}

type DeleteAccountRequest struct {
	Password     string `json:"password" binding:"required"`
	Confirmation string `json:"confirmation" binding:"required"`
	Reason       string `json:"reason,omitempty"`
}
