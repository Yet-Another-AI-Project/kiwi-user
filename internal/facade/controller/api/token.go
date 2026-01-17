package api

import (
	"kiwi-user/internal/facade/dto"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/gin-gonic/gin"
)

// GetPublickKey godoc
// @Summary GetPublickKey
// @Tags Token
// @Description GetPublickKey
// @Accept  json
// @Produce  json
// @Success 200 {object}  facade.BaseResponse{data=dto.GetPublicKeyResponse}
//
// @Router /v1/token/publickey [get]
func (c *Controller) GetPublickKey(ctx *gin.Context) (*dto.GetPublicKeyResponse, *facade.Error) {
	pbkey, err := c.tokenApplication.GetPublicKey(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.GetPublicKeyResponse{
		PublicKey: pbkey,
	}, nil
}

// RefreshAccessToken godoc
// @Summary RefreshAccessToken
// @Tags Token
// @Description RefreshAccessToken
// @Accept  json
// @Produce  json
// @Param  request body dto.RefreshAccessTokenRequest true "refresh access token request"
// @Success 200 {object}  facade.BaseResponse{data=dto.RefreshAccessTokenResponse}
//
// @Router /v1/token/refresh [post]
func (c *Controller) RefreshAccessToken(ctx *gin.Context) (*dto.RefreshAccessTokenResponse, *facade.Error) {
	var request dto.RefreshAccessTokenRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	response, err := c.tokenApplication.RefreshAccessToken(ctx, request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// VerifyAccessToken godoc
// @Summary VerifyAccessToken
// @Tags Token
// @Description VerifyAccessToken
// @Accept  json
// @Produce  json
// @Param  request body dto.VerifyAccessTokenRequest true "verify access token request"
// @Success 200 {object}  facade.BaseResponse{data=dto.VerifyAccessTokenResponse}
//
// @Router /v1/token/verify [post]
func (c *Controller) VerifyAccessToken(ctx *gin.Context) (*dto.VerifyAccessTokenResponse, *facade.Error) {
	var request dto.VerifyAccessTokenRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	userInfo, err := c.tokenApplication.VerifyAccessToken(ctx, request.AccessToken)
	if err != nil {
		if err.Code == facade.ErrForbidden.Code {
			return &dto.VerifyAccessTokenResponse{
				Success: false,
			}, nil
		}
		return nil, err
	}

	return &dto.VerifyAccessTokenResponse{
		Success:  true,
		UserInfo: userInfo,
	}, nil
}

// Logout godoc
// @Summary Logout
// @Tags Token
// @Description Logout
// @Accept  json
// @Produce  json
// @Param  request body dto.LogoutRequest true "logout request"
// @Success 200 {object}  facade.BaseResponse
//
// @Router /v1/token/logout [post]
func (c *Controller) Logout(ctx *gin.Context, userID string) (*dto.OperationResponse, *facade.Error) {
	var request dto.LogoutRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}
	request.UserID = userID

	if err := c.tokenApplication.Logout(ctx, request); err != nil {
		return nil, err
	}
	return &dto.OperationResponse{
		Success: true,
	}, nil
}
