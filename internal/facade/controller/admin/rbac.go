package admin

import (
	"kiwi-user/internal/facade/dto"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/gin-gonic/gin"
)

// CreateApplication godoc
// @Summary CreateApplication
// @Tags Admin
// @Description CreateApplication
// @Accept  json
// @Produce  json
// @Param  request body dto.CreateApplicationRequest true "create application request"
// @Success 200 {object}  facade.BaseResponse{data=dto.OperationResponse}
//
// @Router /admin/rbac/application [post]
func (c *Controller) CreateApplication(ctx *gin.Context, userID string) (*dto.OperationResponse, *facade.Error) {
	request := &dto.CreateApplicationRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	if err := c.rbacApplication.CreateApplication(ctx, request.Name); err != nil {
		return nil, err
	}

	return &dto.OperationResponse{
		Success: true,
	}, nil
}

// GetApplication godoc
// @Summary GetApplication
// @Tags Admin
// @Description GetApplication
// @Accept  json
// @Produce  json
// @Param  name query string true "application name"
// @Success 200 {object}  facade.BaseResponse{data=dto.Application}
//
// @Router /admin/rbac/application [get]
func (c *Controller) GetApplication(ctx *gin.Context, userID string) (*dto.Application, *facade.Error) {
	request := &dto.GetApplicationRequest{}
	if err := ctx.ShouldBindQuery(request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	application, roles, err := c.rbacApplication.GetApplication(ctx, request.Name)
	if err != nil {
		return nil, err
	}

	return convertApplicateionAggregateToDTO(application, roles), nil
}

// CreateRole godoc
// @Summary CreateRole
// @Tags Admin
// @Description CreateRole
// @Accept  json
// @Produce  json
// @Param  request body dto.CreateRoleRequest true "create role request"
// @Success 200 {object}  facade.BaseResponse{data=dto.OperationResponse}
//
// @Router /admin/rbac/role [post]
func (c *Controller) CreateRole(ctx *gin.Context, userID string) (*dto.OperationResponse, *facade.Error) {
	request := &dto.CreateRoleRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	if err := c.rbacApplication.CreateRole(ctx, request); err != nil {
		return nil, err
	}

	return &dto.OperationResponse{
		Success: true,
	}, nil
}

// CreateScope godoc
// @Summary CreateScope
// @Tags Admin
// @Description CreateScope
// @Accept  json
// @Produce  json
// @Param  request body dto.CreateScopeRequest true "create scope request"
// @Success 200 {object}  facade.BaseResponse{data=dto.OperationResponse}
//
// @Router /admin/rbac/scope [post]
func (c *Controller) CreateScope(ctx *gin.Context, userID string) (*dto.OperationResponse, *facade.Error) {
	request := &dto.CreateScopeRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	if err := c.rbacApplication.CreateRoleScope(ctx, request); err != nil {
		return nil, err
	}

	return &dto.OperationResponse{
		Success: true,
	}, nil
}

// UpdateApplicationDefaultRole godoc
// @Summary UpdateApplicationDefaultRole
// @Tags Admin
// @Description UpdateApplicationDefaultRole
// @Accept  json
// @Produce  json
// @Param  request body dto.UpdateDefaultRoleRequest true "set default role request"
// @Success 200 {object}  facade.BaseResponse{data=dto.OperationResponse}
//
// @Router /admin/rbac/application/default-role [put]
func (c *Controller) UpdateApplicationDefaultRole(ctx *gin.Context, userID string) (*dto.OperationResponse, *facade.Error) {
	request := &dto.UpdateDefaultRoleRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	if err := c.rbacApplication.UpdateApplicationDefaultRole(ctx, request); err != nil {
		return nil, err
	}

	return &dto.OperationResponse{
		Success: true,
	}, nil
}
