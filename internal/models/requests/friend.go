package requests

type GetFriendRequestType string

const (
	GetFriendRequestTypeAll      GetFriendRequestType = "all"
	GetFriendRequestTypeSent     GetFriendRequestType = "sent"
	GetFriendRequestTypeReceived GetFriendRequestType = "received"
)

type GetFriendRequestsRequest struct {
	Type   GetFriendRequestType `form:"type" json:"type" binding:"required"`
	Before string               `form:"before" json:"before,omitempty"`
	After  string               `form:"after" json:"after,omitempty"`
	Limit  int                  `form:"limit" json:"limit" binding:"required"`
}

type GetFriendsRequest struct {
	Search string `form:"search" json:"search,omitempty"`
	Before string `form:"before" json:"before,omitempty"`
	After  string `form:"after" json:"after,omitempty"`
	Limit  int    `form:"limit" json:"limit" binding:"required"`
}

type SendFriendRequest struct {
	TargetID uint   `json:"target_id,required"`
	Message  string `json:"message,omitempty"`
}

type AcceptFriendRequest struct {
	UserID uint `form:"user_id" json:"user_id" binding:"required"`
}

type DeclineFriendRequest struct {
	UserID uint `form:"user_id" json:"user_id" binding:"required"`
}
