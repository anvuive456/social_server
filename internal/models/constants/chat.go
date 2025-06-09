package constants

type ChatRoomType string

const (
	ChatRoomTypePrivate ChatRoomType = "private"
	ChatRoomTypeGroup   ChatRoomType = "group"
)

type ParticipantRole string

const (
	ParticipantRoleMember ParticipantRole = "member"
	ParticipantRoleAdmin  ParticipantRole = "admin"
	ParticipantRoleOwner  ParticipantRole = "owner"
)

type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeVideo    MessageType = "video"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeFile     MessageType = "file"
	MessageTypeSystem   MessageType = "system"
	MessageTypeLocation MessageType = "location"
)

type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

type WebSocketEventType string

const (
	WSEventNewMessage       WebSocketEventType = "new_message"
	WSEventMessageRead      WebSocketEventType = "message_read"
	WSEventMessageDelivered WebSocketEventType = "message_delivered"
	WSEventTypingStart      WebSocketEventType = "typing_start"
	WSEventTypingStop       WebSocketEventType = "typing_stop"
	WSEventUserOnline       WebSocketEventType = "user_online"
	WSEventUserOffline      WebSocketEventType = "user_offline"
	WSEventRoomUpdate       WebSocketEventType = "room_update"
	WSEventParticipantJoin  WebSocketEventType = "participant_join"
	WSEventParticipantLeave WebSocketEventType = "participant_leave"
	WSEventMessageReaction  WebSocketEventType = "message_reaction"
)