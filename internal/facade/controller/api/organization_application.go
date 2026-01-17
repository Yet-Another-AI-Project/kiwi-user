package api

import (
	"kiwi-user/internal/facade/dto"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/gin-gonic/gin"
)

// GetOrganizationApplicationInfos godoc
// @Summary GetOrganizationApplicationInfos
// @Tags User
// @Description 分页查看个人申请的试用企业信息
// @Accept  json
// @Produce  json
// @Param name query string false "企业名称"
// @Success 200 {object}  facade.BaseResponse{data=[]dto.OrganizationApplicationResponse}
// @Router /v1/user/organization_application/infos [get]
func (c *Controller) GetOrganizationApplicationInfos(ctx *gin.Context, userID string) ([]*dto.OrganizationApplicationResponse, *facade.Error) {
	name := ctx.Param("name")
	if userID == "" {
		return nil, facade.ErrForbidden.Facade("user_id does not exist")
	}
	organizationApplicationAggregates, ferr := c.organizationApplicationApplication.GetUserOrganizationApplications(ctx, userID, name)
	if ferr != nil {
		return nil, ferr
	}

	result := make([]*dto.OrganizationApplicationResponse, 0)

	for _, organizationApplicationAggregate := range organizationApplicationAggregates {
		result = append(result, &dto.OrganizationApplicationResponse{
			Name:            organizationApplicationAggregate.OrganizationApplication.Name,
			BrandShortName:  organizationApplicationAggregate.OrganizationApplication.BrandShortName,
			PrimaryBusiness: organizationApplicationAggregate.OrganizationApplication.PrimaryBusiness,
			UsageScenario:   organizationApplicationAggregate.OrganizationApplication.UsageScenario,
			ReferrerName:    organizationApplicationAggregate.OrganizationApplication.ReferrerName,
			DiscoveryWay:    organizationApplicationAggregate.OrganizationApplication.DiscoveryWay,
			ReviewStatus:    organizationApplicationAggregate.OrganizationApplication.ReviewStatus,
		})
	}

	return result, nil
}

// CreateOrganizationApplication godoc
// @Summary CreateOrganizationApplication
// @Tags User
// @Description 创建企业试用申请
// @Accept  json
// @Produce  json
// @Param  request body dto.CreateOrganizationApplicationRequest true "create organization_application request"
// @Success 200 {object}  facade.BaseResponse{data=dto.OperationResponse}
// @Router /v1/user/organization_application/request [post]
func (c *Controller) CreateOrganizationApplication(ctx *gin.Context, userID string) (*dto.OperationResponse, *facade.Error) {
	request := &dto.CreateOrganizationApplicationRequest{}

	if err := ctx.BindJSON(request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}
	if userID == "" {
		return nil, facade.ErrBadRequest.Facade("user_id does not exist")
	}
	err := c.organizationApplicationApplication.CreateOrganizationApplication(ctx, userID, request)

	if err != nil {
		return nil, err
	}

	return &dto.OperationResponse{Success: true}, nil
}
