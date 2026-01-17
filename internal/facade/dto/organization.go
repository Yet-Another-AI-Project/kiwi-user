package dto

type GetOrganizationInfosRequest struct {
	OrgIDs []string `json:"org_ids"`
}

type PageOrganizationInfosRequest struct {
	PageNum  int `form:"page_num"`
	PageSize int `form:"page_size"`
}

type CreateOrganizationRequst struct {
	Application string `json:"application"`
	Name        string `json:"name"`
	ExpiresAt   int64  `json:"expires_at"`
	Version     string `json:"version"`
	LogoImage   string `json:"logo_image"`
}

type Organization struct {
	ID             string `json:"id"`
	Application    string `json:"application"`
	Name           string `json:"name"`
	Status         string `json:"status"`
	PermissionCode string `json:"permission_code"`
	RefreshAt      int64  `json:"refresh_at"`
	ExpiresAt      int64  `json:"expires_at"`
	LogoImageURL   string `json:"logo_image_url"`
}

type OrganizationV2 struct {
	ID             string `json:"id"`
	Application    string `json:"application"`
	Name           string `json:"name"`
	Version        string `json:"version"`
	PermissionCode string `json:"permission_code"`
	RefreshAt      int64  `json:"refresh_at"`
	ExpiresAt      int64  `json:"expires_at"`
	LogoImageURL   string `json:"logo_image_url"`
}

type CreateOrganizationUserRequest struct {
	UserID         string `json:"user_id"`
	Role           string `json:"role"`
	OrganizationID string `json:"organization_id"`
}

type UpdateOrganizationRequest struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Version        string `json:"version"`
	ExpiresAt      int64  `json:"expires_at"`
	PermissionCode string `json:"permission_code"`
	LogoImage      string `json:"logo_image"`
}
