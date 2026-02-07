package api

import (
	"encoding/json"
	"kiwi-user/internal/facade/dto"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/gin-gonic/gin"
)

// WechatWebLogin godoc
// @Summary WechatWebLogin
// @Tags Login
// @Description WechatWebLogin
// @Accept  json
// @Produce  json
// @Param  request body dto.WechatWebLoginRequest true "wechat web request"
// @Success 200 {object}  facade.BaseResponse{data=dto.LoginResponse}
//
// @Router /v1/login/wechat/web [post]
func (c *Controller) WechatWebLogin(ctx *gin.Context) (*dto.LoginResponse, *facade.Error) {
	stateStr := ctx.Query("state")
	code := ctx.Request.FormValue("code")

	c.logger.Debugf(ctx, "get state from request %s", stateStr)
	c.logger.Debugf(ctx, "get code from request %s", code)

	if code == "" || stateStr == "" {
		return nil, facade.ErrBadRequest.Facade("code or state is empty")
	}

	state := dto.WechatWebLoginRequest{}
	if err := json.Unmarshal([]byte(stateStr), &state); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	state.Code = code

	response, ferr := c.loginApplication.WechatWebLogin(ctx, state)
	if ferr != nil {
		return nil, ferr
	}

	return response, nil
}

// WechatMiniProgramLogin godoc
// @Summary LoginWechatMiniprogram
// @Tags Login
// @Description LoginWechatMiniprogram
// @Accept  json
// @Produce  json
// @Param  request body dto.WechatMiniProgramLoginRequest true "wechat miniprogram request"
// @Success 200 {object}  facade.BaseResponse{data=dto.LoginResponse}
//
// @Router /v1/login/wechat/miniprogram [post]
func (c *Controller) WechatMiniProgramLogin(ctx *gin.Context) (*dto.LoginResponse, *facade.Error) {
	var request dto.WechatMiniProgramLoginRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	response, err := c.loginApplication.WechatMiniProgramLogin(ctx, request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// PasswordLogin godoc
// @Summary PasswordLogin
// @Tags Login
// @Description PasswordLogin
// @Accept  json
// @Produce  json
// @Param  request body dto.PasswordLoginRequest true "password login request"
// @Success 200 {object}  facade.BaseResponse{data=dto.LoginResponse}
//
// @Router /v1/login/password [post]
func (c *Controller) PasswordLogin(ctx *gin.Context) (*dto.LoginResponse, *facade.Error) {
	var request dto.PasswordLoginRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	response, err := c.loginApplication.PasswordLogin(ctx, request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// OrganizationLogin godoc
// @Summary LoginOrganization
// @Tags Login
// @Description LoginOrganization
// @Accept  json
// @Produce  json
// @Param  request body dto.OrganizationLoginRequest true "login organization request"
// @Success 200 {object}  facade.BaseResponse{data=dto.LoginResponse}
//
// @Router /v1/login/organization [post]
func (c *Controller) OrganizationLogin(ctx *gin.Context) (*dto.LoginResponse, *facade.Error) {
	var request dto.OrganizationLoginRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	response, err := c.loginApplication.OrganizationLogin(ctx, request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// SendPhoneVerifyCode godoc
// @Summary SendSmsVerifyCode
// @Tags Login
// @Description 发送短信验证码
// @Accept  json
// @Produce  json
// @Param  request body dto.SendVerifyCodeRequest true "send sms verify code request"
// @Success 200 {object}  facade.BaseResponse{data=dto.OperationResponse}
//
// @Router /v1/login/phone/verify_code [post]
func (c *Controller) SendPhoneVerifyCode(ctx *gin.Context) (*dto.OperationResponse, *facade.Error) {
	var request dto.SendVerifyCodeRequest
	if err := ctx.BindJSON(&request); err != nil {
		c.logger.Errorf(ctx, "SendPhoneVerifyCode Request Error %v", err)
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	if err := c.loginApplication.SendPhoneVerifyCode(ctx, request.Phone); err != nil {
		c.logger.Errorf(ctx, "SendPhoneVerifyCode Api Error %v", err)
		return nil, err
	}
	return &dto.OperationResponse{
		Success: true,
	}, nil
}

// PhoneLogin godoc
// @Summary PhoneLogin
// @Tags Login
// @Description Phone verification code login and registration
// @Accept  json
// @Produce  json
// @Param  request body dto.PhoneLoginRequest true "phone login request"
// @Success 200 {object}  facade.BaseResponse{data=dto.LoginResponse}
//
// @Router /v1/login/phone [post]
func (c *Controller) PhoneLogin(ctx *gin.Context) (*dto.LoginResponse, *facade.Error) {
	var request dto.PhoneLoginRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	response, err := c.loginApplication.PhoneLogin(ctx, request)
	if err != nil {
		c.logger.Errorf(ctx, "PhoneLogin output %v, request %v, error %v", response, request, err)
		return nil, err
	}

	return response, nil
}

// SendEmailVerificationCode godoc
// @Summary SendEmailVerificationCode
// @Tags Login
// @Description Send email verification code for login
// @Accept  json
// @Produce  json
// @Param  request body dto.SendEmailVerificationCodeRequest true "send email verification code request"
// @Success 200 {object}  facade.BaseResponse{data=dto.SendEmailVerificationCodeResponse}
// @Router /v1/login/email/verify_code [post]
func (c *Controller) SendEmailVerificationCode(ctx *gin.Context) (*dto.SendEmailVerificationCodeResponse, *facade.Error) {
	var request dto.SendEmailVerificationCodeRequest
	if err := ctx.BindJSON(&request); err != nil {
		return &dto.SendEmailVerificationCodeResponse{
			Success: false,
			Message: "Invalid request format",
		}, facade.ErrBadRequest.Wrap(err)
	}

	err := c.loginApplication.SendEmailVerificationCode(ctx, request)
	if err != nil {
		return &dto.SendEmailVerificationCodeResponse{
			Success: false,
			Message: "Failed to send verification code",
		}, err
	}

	return &dto.SendEmailVerificationCodeResponse{
		Success: true,
	}, nil
}

// EmailLogin godoc
// @Summary EmailLogin
// @Tags Login
// @Description Email verification code login and registration
// @Accept  json
// @Produce  json
// @Param  request body dto.EmailLoginRequest true "email login request"
// @Success 200 {object}  facade.BaseResponse{data=dto.LoginResponse}
//
// @Router /v1/login/email [post]
func (c *Controller) EmailLogin(ctx *gin.Context) (*dto.LoginResponse, *facade.Error) {
	var request dto.EmailLoginRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	response, err := c.loginApplication.EmailLogin(ctx, request)
	c.logger.Debugf(ctx, "EmailLogin output %v, request %v, error %v", response, request, err)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GoogleWebLogin godoc
// @Summary GoogleWebLogin
// @Tags Login
// @Description Google OAuth web login using ID token
// @Accept  json
// @Produce  json
// @Param  request body dto.GoogleWebLoginRequest true "google web login request"
// @Success 200 {object}  facade.BaseResponse{data=dto.LoginResponse}
//
// @Router /v1/login/google/web [post]
func (c *Controller) GoogleWebLogin(ctx *gin.Context) (*dto.LoginResponse, *facade.Error) {
	var request dto.GoogleWebLoginRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	response, err := c.loginApplication.GoogleWebLogin(ctx, request)
	if err != nil {
		c.logger.Errorf(ctx, "GoogleWebLogin error: %v", err)
		return nil, err
	}

	return response, nil
}
