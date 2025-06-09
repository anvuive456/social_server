package constants

type PostType string

const (
	PostTypeText  PostType = "text"
	PostTypeImage PostType = "image"
	PostTypeVideo PostType = "video"
	PostTypeLink  PostType = "link"
	PostTypeAudio PostType = "audio"
)

type PostPrivacy string

const (
	PostPrivacyPublic  PostPrivacy = "public"
	PostPrivacyFriends PostPrivacy = "friends"
	PostPrivacyPrivate PostPrivacy = "private"
)