package dto

type RefreshAccessTokenRequest struct {
	UserID       string  `json:"user_id" binding:"required"`
	RefreshToken string  `json:"refresh_token" binding:"required"`
	Device       *Device `json:"device" binding:"required"`
}

type RefreshAccessTokenResponse struct {
	LoginResponse
}

type VerifyAccessTokenRequest struct {
	AccessToken string `json:"access_token"`
}

type VerifyAccessTokenResponse struct {
	Success  bool      `json:"success"`
	UserInfo *UserInfo `json:"user_info"`
}

type GetPublicKeyResponse struct {
	PublicKey string `json:"public_key"`
}

type LogoutRequest struct {
	UserID       string  `json:"user_id"`
	RefreshToken string  `json:"refresh_token" binding:"required"`
	Device       *Device `json:"device" binding:"required"`
}
