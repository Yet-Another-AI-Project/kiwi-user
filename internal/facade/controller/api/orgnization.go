package api

import (
	"kiwi-user/internal/facade/dto"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetOrganizationInfos godoc
// @Summary GetOrganizationInfos
// @Tags User
// @Description GetOrganizationInfos
// @Accept  json
// @Produce  json
// @Param  request body dto.GetOrganizationInfosRequest true "get organization infos request"
// @Success 200 {object}  facade.BaseResponse{data=[]dto.Organization}
//
// @Router /internal/organization/infos [post]
func (c *Controller) GetOrganizationInfos(ctx *gin.Context) ([]*dto.Organization, *facade.Error) {
	var request dto.GetOrganizationInfosRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	organizationInfos, ferr := c.organizationApplication.GetOrganizationInfos(ctx, request)
	if ferr != nil {
		return nil, ferr
	}

	result := make([]*dto.Organization, 0)

	for _, organizationInfo := range organizationInfos {
		result = append(result, convertOrganizationAggregateToDTO(organizationInfo))
	}

	return result, nil
}

func (c *Controller) CreateOrganizationRequest(ctx *gin.Context, userID uuid.UUID) {
	panic("implement me")
}

func (c *Controller) ApproveOrganizationRequest(ctx *gin.Context) {
	panic("implement me")
}

func (c *Controller) ListOrganizationUsers(ctx *gin.Context) {
	panic("implement me")
}

func (c *Controller) DeleteOrignizationUser(ctx *gin.Context) {
	panic("implement me")
}
