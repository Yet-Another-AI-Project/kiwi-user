package dto

import (
	"kiwi-user/internal/domain/model/enum"
)

type CreateOrganizationApplicationRequest struct {
	Application     string `json:"application"`
	Name            string `json:"name"`             // 企业名称
	BrandShortName  string `json:"brand_short_name"` // 品牌简称
	PrimaryBusiness string `json:"primary_business"` // 主营业务
	UsageScenario   string `json:"usage_scenario"`   // 使用诉求
	ReferrerName    string `json:"referrer_name"`    // 推荐人，选填
	DiscoveryWay    string `json:"discovery_way"`    // 发现途径，选填
	OrgRoleName     string `json:"org_role_name"`    // 组织角色名称
}

type OrganizationApplicationRequest struct {
	Name     string                        `json:"name"`
	Industry enum.OrganizationIndustryType `json:"industry"`
}

type OrganizationApplicationResponse struct {
	Name            string                         `json:"name"`
	BrandShortName  string                         `json:"brand_short_name"` // 品牌简称
	PrimaryBusiness string                         `json:"primary_business"` // 主营业务
	UsageScenario   string                         `json:"usage_scenario"`   // 使用诉求
	ReferrerName    string                         `json:"referrer_name"`    // 推荐人，选填
	DiscoveryWay    string                         `json:"discovery_way"`    // 发现途径，选填
	ReviewStatus    enum.OrganizationRequestStatus `json:"review_status"`
	OrgRoleName     string                         `json:"org_role_name"` // 组织角色名称
}

type OrganizationApplication struct {
	ID              string                         `json:"id"`
	ApplicationID   string                         `json:"application_id"`
	Name            string                         `json:"name"`
	Status          string                         `json:"status"`
	TrailDays       int32                          `json:"trail_days"`
	BrandShortName  string                         `json:"brand_short_name"` // 品牌简称
	PrimaryBusiness string                         `json:"primary_business"` // 主营业务
	UsageScenario   string                         `json:"usage_scenario"`   // 使用诉求
	ReferrerName    string                         `json:"referrer_name"`    // 推荐人，选填
	DiscoveryWay    string                         `json:"discovery_way"`    // 发现途径，选填
	ReviewStatus    enum.OrganizationRequestStatus `json:"review_status"`
	ReviewComment   string                         `json:"review_comment"`
	UserID          string                         `json:"user_id"`
	OrgRoleName     string                         `json:"org_role_name"` // 组织角色名称
}

type UpdateOrganizationApplicationRequest struct {
	ID            string                         `json:"id" binding:"required"`
	Name          string                         `json:"name"`
	ReviewStatus  enum.OrganizationRequestStatus `json:"review_status" binding:"required"`
	ReviewComment string                         `json:"review_comment" binding:"required"`
}
