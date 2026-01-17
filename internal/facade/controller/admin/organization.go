package admin

import (
	"kiwi-user/internal/facade/dto"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (c *Controller) ListOrganizations(ctx *gin.Context) {
	panic("implement me")
}

func (c *Controller) ApproveOrganizationRequest(ctx *gin.Context) {
	panic("implement me")
}

func (c *Controller) CreateOrganizationRequest(ctx *gin.Context) {
	panic("implement me")
}

// CreateOrganization godoc
// @Summary CreateOrganization
// @Tags Admin
// @Description CreateOrganization
// @Accept  json
// @Produce  json
// @Param  request body dto.CreateOrganizationRequst true "create organization request"
// @Success 200 {object}  facade.BaseResponse{data=dto.OrganizationV2}
//
// @Router /admin/organization [post]
func (c *Controller) CreateOrganization(ctx *gin.Context) (*dto.OrganizationV2, *facade.Error) {
	request := &dto.CreateOrganizationRequst{}

	if err := ctx.BindJSON(request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	organization, err := c.organizationApplication.CreateOrganization(ctx, request)

	if err != nil {
		return nil, err
	}

	return convertOrganizationAggregateToDTO(organization), nil
}

// CreateOrganizationUser godoc
// @Summary CreateOrganizationUser
// @Tags Admin
// @Description CreateOrganizationUser
// @Accept  json
// @Produce  json
// @Param  request body dto.CreateOrganizationUserRequest true "create organization user request"
// @Success 200 {object}  facade.BaseResponse{data=dto.OperationResponse}
//
// @Router /admin/organization/user [post]
func (c *Controller) CreateOrganizationUser(ctx *gin.Context) (*dto.OperationResponse, *facade.Error) {
	request := &dto.CreateOrganizationUserRequest{}
	if err := ctx.BindJSON(request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	if err := c.organizationApplication.CreateOrganizationUser(ctx, request); err != nil {
		return nil, err
	}

	return &dto.OperationResponse{
		Success: true,
	}, nil
}

// GetOrganizationUserInfos godoc
// @Summary GetOrganizationUserInfos
// @Tags Admin
// @Description GetOrganizationUserInfos
// @Accept  json
// @Produce  json
// @Param  org_id query string true "organization id"
// @Success 200 {object}  facade.BaseResponse{data=[]dto.PublicUserInfo}
//
// @Router /admin/organization/user/infos [get]
func (c *Controller) GetOrganizationUserInfos(ctx *gin.Context) ([]*dto.PublicUserInfo, *facade.Error) {
	organizationIDStr := ctx.Query("org_id")

	organizationID, err := uuid.Parse(organizationIDStr)
	if err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	userInfos, ferr := c.userApplication.GetOrganizationUserInfos(ctx, organizationID)
	if ferr != nil {
		return nil, ferr
	}

	return userInfos, nil
}

// DeleteOrganizationUser godoc
// @Summary DeleteOrganizationUser
// @Tags Admin
// @Description DeleteOrganizationUser
// @Accept  json
// @Produce  json
// @Param  request body dto.DeleteOrganizationUserRequest true "delete organization user request"
// @Success 200 {object}  facade.BaseResponse{data=dto.OperationResponse}
//
// @Router /admin/organization/user [delete]
func (c *Controller) DeleteOrganizationUser(ctx *gin.Context) (*dto.OperationResponse, *facade.Error) {
	var request dto.DeleteOrganizationUserRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	organizationID, err := uuid.Parse(request.OrganizationID)
	if err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	if err := c.userApplication.DeleteOrganizationUser(ctx, request.UserID, organizationID); err != nil {
		return nil, err
	}

	return &dto.OperationResponse{Success: true}, nil
}

// GetOrganizationInfos godoc
// @Summary GetOrganizationInfos
// @Tags Admin
// @Description GetOrganizationInfos
// @Accept  json
// @Produce  json
// @Param page_num query int false "当前页"
// @Param page_size query int false "页大小"
// @Success 200 {object}  facade.BaseResponse{data=[]dto.OrganizationV2}
//
// @Router /admin/organization/infos [get]
func (c *Controller) GetOrganizationInfos(ctx *gin.Context) (*facade.PageResponse[*dto.OrganizationV2], *facade.Error) {
	request := &dto.PageOrganizationInfosRequest{}
	if err := ctx.BindQuery(request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	if request.PageNum == 0 {
		request.PageNum = 1
	}

	if request.PageSize == 0 {
		request.PageSize = 10
	}

	organizationInfos, total, ferr := c.organizationApplication.PageOrganizationInfos(ctx, request.PageNum, request.PageSize)
	if ferr != nil {
		return nil, ferr
	}

	result := make([]*dto.OrganizationV2, 0)

	for _, organizationInfo := range organizationInfos {
		result = append(result, convertOrganizationAggregateToDTO(organizationInfo))
	}

	return &facade.PageResponse[*dto.OrganizationV2]{
		Total:    total,
		List:     result,
		PageNum:  request.PageNum,
		PageSize: request.PageSize,
	}, nil
}

// UpdateOrganization godoc
// @Summary UpdateOrganization
// @Tags Admin
// @Description UpdateOrganization
// @Accept  json
// @Produce  json
// @Param  request body dto.UpdateOrganizationRequest true "update organization request"
// @Success 200 {object}  facade.BaseResponse{data=dto.OrganizationV2}
//
// @Router /admin/organization [put]
func (c *Controller) UpdateOrganization(ctx *gin.Context) (*dto.OrganizationV2, *facade.Error) {
	request := &dto.UpdateOrganizationRequest{}
	if err := ctx.BindJSON(request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	organization, err := c.organizationApplication.UpdateOrganization(ctx, request)

	if err != nil {
		return nil, err
	}

	return convertOrganizationAggregateToDTO(organization), nil
}
