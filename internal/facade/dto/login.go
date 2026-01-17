package dto

import (
	"kiwi-user/internal/domain/model/enum"
	"strings"
)

type WechatWebLoginRequest struct {
	ApplicationName string           `json:"application_name"`
	ReferralChannel *ReferralChannel `json:"referral_channel"`
	Device          *Device          `json:"device"`
	Code            string           `json:"code"`
	Platform        string           `json:"platform"`
}

type WechatMiniProgramLoginRequest struct {
	ApplicationName      string           `json:"application_name" binding:"required"`
	Code                 string           `json:"code" binding:"required"`
	ReferralChannel      *ReferralChannel `json:"referral_channel"`
	Device               *Device          `json:"device" binding:"required"`
	MiniProgramPhoneCode string           `json:"mini_program_phone_code"` // 非必填，微信小程序获取手机号的code，当传入时会通过微信接口获取手机号并执行绑定处理
}

type QyWechatLoginRequest struct {
	ApplicationName string           `json:"application_name" binding:"required"`
	Code            string           `json:"code" binding:"required"`
	ReferralChannel *ReferralChannel `json:"referral_channel"`
	Device          *Device          `json:"device" binding:"required"`
}

type PasswordLoginRequest struct {
	ApplicationName string  `json:"application_name" binding:"required"`
	Name            string  `json:"name" binding:"required"`
	Password        string  `json:"password" binding:"required"`
	Device          *Device `json:"device" binding:"required"`
}

type OrganizationLoginRequest struct {
	OrganizationID string  `json:"organization_id" binding:"required"`
	UserID         string  `json:"user_id" binding:"required"`
	RefreshToken   string  `json:"refresh_token" binding:"required"`
	Device         *Device `json:"device" binding:"required"`
}

type Device struct {
	DeviceType string `json:"device_type" binding:"required"`
	DeviceID   string `json:"device_id" binding:"required"`
}

type ReferralChannel struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LoginResponse struct {
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresAt int64  `json:"refresh_token_expires_at"`
	AccessToken           string `json:"access_token"`
	AccessTokenExpiresAt  int64  `json:"access_token_expires_at"`
	Type                  string `json:"type"`
	DeviceType            string `json:"device_type"`
	DeviceID              string `json:"device_id"`
	UserID                string `json:"user_id"`
}

type PhoneLoginRequest struct {
	ApplicationName string  `json:"application_name" binding:"required"`
	Phone           string  `json:"phone" binding:"required"`
	VerifyCode      string  `json:"verify_code" binding:"required"`
	Device          *Device `json:"device" binding:"required"`
}

type SendEmailVerificationCodeRequest struct {
	Email    string                     `json:"email" binding:"required,email"`
	CodeType enum.VertificationCodeType `json:"code_type" binding:"required"`
}

type SendEmailVerifyCodeWithCaptchaRequest struct {
	Email              string                     `json:"email" binding:"required,email"`
	CodeType           enum.VertificationCodeType `json:"code_type" binding:"required"`
	CaptchaVerifyParam string                     `json:"captcha_verify_param" binding:"required"`
}

type EmailLoginRequest struct {
	ApplicationName string  `json:"application_name" binding:"required"`
	Email           string  `json:"email" binding:"required,email"`
	VerifyCode      string  `json:"verify_code" binding:"required"`
	Device          *Device `json:"device" binding:"required"`
}

type GoogleWebLoginRequest struct {
	ApplicationName string           `json:"application_name" binding:"required"`
	Code            string           `json:"code" binding:"required"`
	RedirectURI     string           `json:"redirect_uri" binding:"required"`
	ReferralChannel *ReferralChannel `json:"referral_channel"`
	Device          *Device          `json:"device" binding:"required"`
}

type SendEmailVerificationCodeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type HiddenLoginTypeResponse struct {
	HiddenType []string `json:"hidden_type"`
}

type SetHiddenLoginTypeRequest struct {
	HiddenType []string `json:"hidden_type" binding:"required"`
}

// IsBase64Image 判断avatar是否为base64图片
func (r *UpdateUserInfoRequest) IsBase64Image() bool {
	if r.Avatar == "" {
		return false
	}
	// 检查是否以 "data:image" 开头
	return strings.HasPrefix(r.Avatar, "data:image")
}

// GetBase64Content 从base64字符串中提取实际的base64内容
func (r *UpdateUserInfoRequest) GetBase64Content() string {
	if !r.IsBase64Image() {
		return ""
	}
	// 移除 "data:image/xxx;base64," 前缀
	parts := strings.Split(r.Avatar, ",")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}
