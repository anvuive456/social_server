package constants

type CallType string

const (
	CallTypeVideo CallType = "video"
	CallTypeAudio CallType = "audio"
)

type CallStatus string

const (
	CallStatusPending  CallStatus = "pending"
	CallStatusRinging  CallStatus = "ringing"
	CallStatusOngoing  CallStatus = "ongoing"
	CallStatusEnded    CallStatus = "ended"
	CallStatusDeclined CallStatus = "declined"
	CallStatusMissed   CallStatus = "missed"
)