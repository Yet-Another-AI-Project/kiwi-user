package admin

import (
	"kiwi-user/internal/facade/dto"
	"regexp"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/gin-gonic/gin"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// CreateRoleUser godoc
// @Summary CreateRoleUser
// @Tags Admin
// @Description CreateRoleUser
// @Accept  json
// @Produce  json
// @Param  request body dto.UpdateUserRoleRequest true "create role request"
// @Success 200 {object}  facade.BaseResponse{data=dto.OperationResponse}
//
// @Router /admin/role/user [post]
func (c *Controller) CreateUserRole(ctx *gin.Context) (*dto.OperationResponse, *facade.Error) {

	request := dto.UpdateUserRoleRequest{}

	if err := ctx.BindJSON(request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	if err := c.userApplication.UpdateUserRole(ctx, request); err != nil {
		return nil, err
	}

	return &dto.OperationResponse{
		Success: true,
	}, nil
}

// CreateUserWithPassword godoc
// @Summary CreateUserWithPassword
// @Tags Admin
// @Description CreateUserWithPassword
// @Accept  json
// @Produce  json
// @Param  request body dto.CreateUserWithPasswordRequest true "create user with password request"
// @Success 200 {object}  facade.BaseResponse{data=dto.UserInfo}
//
// @Router /admin/user/password [post]
func (c *Controller) CreateUserWithPassword(ctx *gin.Context) (*dto.UserInfo, *facade.Error) {
	request := dto.CreateUserWithPasswordRequest{}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	// validate request name
	if !usernameRegex.MatchString(request.Name) {
		return nil, facade.ErrBadRequest.Facade("invalid username")
	}

	userInfo, err := c.userApplication.CreateUserWithPassword(ctx, request)

	if err != nil {
		return nil, err
	}

	return userInfo, nil
}
