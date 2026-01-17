package dto

type Application struct {
	Name                         string  `json:"name"`
	Roles                        []*Role `json:"roles"`
	DefaultPersonalRole          string  `json:"default_personal_role"`
	DefaultOrganizationRole      string  `json:"default_organization_role"`
	DefaultOrganizationAdminRole string  `json:"default_organization_admin_role"`
}

type Role struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Scopes []string `json:"scopes"`
}

type GetApplicationRequest struct {
	Name string `form:"name" binding:"required"`
}

type CreateApplicationRequest struct {
	Name string `json:"name"`
}

type DeleteApplicationRequest struct {
	Name string `json:"name"`
}

type CreateRoleRequest struct {
	ApplicationName string `json:"application_name"`
	RoleType        string `json:"role_type"`
	RoleName        string `json:"role_name"`
}

type DeleteRoleRequest struct {
	ApplicationName string `json:"application_name"`
	RoleName        string `json:"role_name"`
}

type UpdateDefaultRoleRequest struct {
	Type            string `json:"type"`
	ApplicationName string `json:"application_name"`
	RoleName        string `json:"role_name"`
}

type SetUserRoleRequest struct {
	ApplicationName string `json:"application_name"`
	UserID          string `json:"user_id"`
	RoleName        string `json:"role_name"`
}

type CreateScopeRequest struct {
	ApplicationName string `json:"application_name"`
	RoleName        string `json:"role_name"`
	ScopeName       string `json:"scope_name"`
}

type DeleteScopeRequest struct {
	ApplicationName string `json:"application_name"`
	RoleName        string `json:"role_name"`
	ScopeName       string `json:"scope_name"`
}
