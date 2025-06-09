package requests

type LoginRequest struct {
	EmailOrUsername string `json:"email" binding:"required"`
	Password        string `json:"password" binding:"required"`
	RememberMe      bool   `json:"remember_me"`
	DeviceInfo      string `json:"device_info,omitempty"`
	Platform        string `json:"platform,omitempty"`
}

type RegisterRequest struct {
	Email         string `json:"email" binding:"required,email"`
	Password      string `json:"password" binding:"required,min=6"`
	AcceptTerms   bool   `json:"accept_terms" binding:"required"`
	AcceptPrivacy bool   `json:"accept_privacy" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}
