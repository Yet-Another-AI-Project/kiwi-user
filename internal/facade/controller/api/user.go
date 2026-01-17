package api

import (
	"kiwi-user/internal/facade/dto"

	"github.com/google/uuid"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/gin-gonic/gin"
)

// GetUserInfo godoc
// @Summary GetUserInfo
// @Tags User
// @Description GetUserInfo
// @Accept  json
// @Produce  json
// @Success 200 {object}  facade.BaseResponse{data=dto.UserInfo}
//
// @Router /v1/user/info [get]
func (c *Controller) GetUserInfo(ctx *gin.Context, userID string) (*dto.UserInfo, *facade.Error) {
	organizationID := ctx.GetString("org_id")

	userInfo, err := c.userApplication.GetDetailUserInfo(ctx, userID, organizationID)
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

// GetPublicUserInfos GetUserInfos godoc
// @Summary GetUserInfos
// @Tags User
// @Description GetUserInfos
// @Accept  json
// @Produce  json
// @Param  request body dto.GetUserInfosRequest true "get user infos request"
// @Success 200 {object}  facade.BaseResponse{data=[]dto.PublicUserInfo}
//
// @Router /internal/user/infos [post]
func (c *Controller) GetPublicUserInfos(ctx *gin.Context) ([]*dto.PublicUserInfo, *facade.Error) {
	var request dto.GetUserInfosRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	userInfos, err := c.userApplication.GetPublicUserInfos(ctx, request.UserIDs)
	if err != nil {
		return nil, err
	}

	return userInfos, nil
}

// UpdateUserInfo godoc
// @Summary UpdateUserInfo
// @Tags User
// @Description UpdateUserInfo
// @Accept  json
// @Produce  json
// @Param  request body dto.UpdateUserInfoRequest true "update user info request"
// @Success 200 {object}  facade.BaseResponse{data=dto.UserInfo}
//
// @Router /v1/user/info [put]
func (c *Controller) UpdateUserInfo(ctx *gin.Context, userID string) (*dto.UserInfo, *facade.Error) {
	var request dto.UpdateUserInfoRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	userInfo, err := c.userApplication.UpdateUserInfo(ctx, userID, request)

	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

// ChangePassword godoc
// @Summary ChangePassword
// @Tags User
// @Description Change user password
// @Accept  json
// @Produce  json
// @Param  request body dto.ChangePasswordRequest true "change password request"
// @Success 200 {object}  facade.BaseResponse{data=dto.OperationResponse}
//
// @Router /v1/user/password [post]
func (c *Controller) ChangePassword(ctx *gin.Context, userID string) (*dto.OperationResponse, *facade.Error) {
	var request dto.ChangePasswordRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	return c.userApplication.ChangePassword(ctx, userID, request)
}

// VerifyPhoneCode godoc
func (c *Controller) VerifyPhoneCode(ctx *gin.Context, userID string) (*dto.OperationResponse, *facade.Error) {
	var request dto.VerifyPhoneCodeRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	if err := c.bindingApplication.VerifyPhoneCode(ctx, userID, &request); err != nil {
		return nil, err
	}

	return &dto.OperationResponse{
		Success: true,
	}, nil
}

// BindingPhoneWithMiniProgramCode BindingPhone godoc
// @Summary BindingPhone
// @Tags User
// @Description BindingPhone
// @Accept  json
// @Produce  json
// @Param  request body dto.BindingPhoneWithMiniProgramCodeRequest true "binding phone request"
// @Success 200 {object}  facade.BaseResponse{data=dto.OperationResponse}
//
// @Router /v1/user/binding/phone [post]
func (c *Controller) BindingPhoneWithMiniProgramCode(ctx *gin.Context, userID string) (*dto.OperationResponse, *facade.Error) {
	var request dto.BindingPhoneWithMiniProgramCodeRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	if err := c.userApplication.BindingPhone(ctx, userID, request.Phone, request.MiniProgramCode); err != nil {
		return nil, err
	}

	return &dto.OperationResponse{
		Success: true,
	}, nil
}

// BindingPhoneWithVerifyCode godoc
// @Summary BindingPhoneWithVerifyCode
// @Tags User
// @Description 使用短信验证码绑定用户手机
// @Accept  json
// @Produce  json
// @Param  request body dto.BindingPhoneWithVerifyCodeRequest true "binding phone request"
// @Success 200 {object}  facade.BaseResponse{data=dto.OperationResponse}
//
// @Router /v1/user/binding/phone/verify_code [post]
func (c *Controller) BindingPhoneWithVerifyCode(ctx *gin.Context, userID string) (*dto.OperationResponse, *facade.Error) {
	var request dto.BindingPhoneWithVerifyCodeRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	if err := c.userApplication.BindingPhoneWithVerifyCode(ctx, userID, request.Phone, request.VerifyCode); err != nil {
		return nil, err
	}

	return &dto.OperationResponse{
		Success: true,
	}, nil
}

// GetCurrentInfos godoc
// @Summary GetUserInfo
// @Tags User
// @Description GetUserInfo
// @Accept  json
// @Produce  json
// @Success 200 {object}  facade.BaseResponse{data=dto.BasicInfos}
//
// @Router /v1/getCurrentInfos [get]
func (c *Controller) GetCurrentInfos(ctx *gin.Context) (*dto.BasicInfos, *facade.Error) {
	organizationIDStr := ctx.GetString("org_id")
	organizationID, err := uuid.Parse(organizationIDStr)
	if err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}
	orgAgg, orgErr := c.organizationApplication.GetOrganizationAggregate(ctx, organizationID)
	if orgErr != nil {
		return nil, orgErr
	}
	return convertOrganizationAggregateToBasicInfos(orgAgg), nil
}
