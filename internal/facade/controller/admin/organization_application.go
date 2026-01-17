package admin

import (
	"kiwi-user/internal/facade/dto"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/Yet-Another-AI-Project/kiwi-lib/server/gin/utils"
	"github.com/gin-gonic/gin"
)

// PageOrganizationApplication godoc
// @Summary PageOrganizationApplication
// @Tags Admin
// @Description 分页获取试用申请信息
// @Accept  json
// @Produce  json
// @Param page_num query int false "当前页"
// @Param page_size query int false "页大小"
// @Param name query string false "企业名称"
// @Success 200 {object}  facade.BaseResponse{data=[]dto.OrganizationApplication}
// @Router /admin/organization_application/infos [get]
func (c *Controller) PageOrganizationApplication(ctx *gin.Context) (*facade.PageResponse[*dto.OrganizationApplication], *facade.Error) {
	pageNum, pageSize := utils.GetPageNumAndSize(ctx)
	name := ctx.Query("name")

	organizationApplicationInfos, total, ferr := c.organizationApplicationApplication.PageOrganizationApplicationInfos(ctx, name, pageNum, pageSize)
	if ferr != nil {
		return nil, ferr
	}

	result := make([]*dto.OrganizationApplication, 0)

	for _, organizationApplication := range organizationApplicationInfos {
		result = append(result, convertOrganizationApplicationAggregateToDTO(organizationApplication))
	}

	return &facade.PageResponse[*dto.OrganizationApplication]{
		Total:    total,
		List:     result,
		PageNum:  pageNum,
		PageSize: pageSize,
	}, nil
}

// ReviewOrganizationApplication godoc
// @Summary ReviewOrganizationApplication
// @Tags Admin
// @Description 审核试用申请
// @Accept  json
// @Produce  json
// @Param  request body dto.UpdateOrganizationApplicationRequest true "审核试用申请请求"
// @Success 200 {object}  facade.BaseResponse{data=dto.OperationResponse}
// @Router /admin/organization_application/audit [put]
func (c *Controller) ReviewOrganizationApplication(ctx *gin.Context) (*dto.OperationResponse, *facade.Error) {
	request := &dto.UpdateOrganizationApplicationRequest{}
	if err := ctx.BindJSON(request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	err := c.organizationApplicationApplication.ReviewOrganizationApplication(ctx, request)

	if err != nil {
		return nil, err
	}

	return &dto.OperationResponse{Success: true}, nil
}
