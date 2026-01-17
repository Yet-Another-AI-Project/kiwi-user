package dto

type GetUserInfosRequest struct {
	UserIDs []string `json:"user_ids"`
}

type UpdateUserInfoRequest struct {
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
	Department  string `json:"department"`
}

type UserInfo struct {
	UserID         string              `json:"id"`
	Application    string              `json:"application"`
	Name           string              `json:"name"`
	Avatar         string              `json:"avatar"`
	Phone          string              `json:"phone"`
	PersonalRole   string              `json:"personal_role"`
	PersonalScopes []string            `json:"personal_scopes"`
	CurrentOrgID   string              `json:"current_org_id"`
	Orgs           []*OrganizationUser `json:"orgs"`
	Department     string              `json:"department"`
}

type PublicUserInfo struct {
	UserID      string `json:"id"`
	Application string `json:"application"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
	Username    string `json:"username"`
	Department  string `json:"department"`
}

type OrganizationUser struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Status             string   `json:"status"`
	PermissionCode     string   `json:"permission_code"`
	RefreshAt          int64    `json:"refresh_at"`
	ExpiresAt          int64    `json:"expires_at"`
	OrganizationRole   string   `json:"organization_role"`
	OrganizationScopes []string `json:"organization_scopes"`
}

type UpdateUserRoleRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type CreateUserWithPasswordRequest struct {
	Application    string `json:"application"`
	Name           string `json:"name" binding:"required,min=6,max=20"`
	Password       string `json:"password" binding:"required,min=6,max=20"`
	OrganizationID string `json:"organization_id"`
	Role           string `json:"role"`
}

type DeleteOrganizationUserRequest struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
}

type BindingPhoneWithMiniProgramCodeRequest struct {
	Phone           string `json:"phone"`
	MiniProgramCode string `json:"mini_program_code"`
}

type BindingPhoneWithVerifyCodeRequest struct {
	Phone      string `json:"phone"`
	VerifyCode string `json:"verify_code"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type SendVerifyCodeRequest struct {
	Phone string `json:"phone"`
}

type SendVerifyCodeWithCaptchaRequest struct {
	Phone              string `json:"phone"`
	CaptchaVerifyParam string `json:"captcha_verify_param" binding:"required"`
}
