package api

import (
	"kiwi-user/internal/facade/dto"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/gin-gonic/gin"
)

func (c *Controller) WxOpenerConfig(ctx *gin.Context, _ string) (*dto.WxAppConfigResponse, *facade.Error) {
	var request dto.WxAppConfigRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}
	res, err := c.configApplication.GetWxOpenConfig(ctx, &request)
	if err != nil {
		c.logger.Errorf(ctx, "GetWxOpenConfig Error %v, request %v", err, request)
	}
	return res, err
}

func (c *Controller) WxOpenerConfigExternal(ctx *gin.Context) (*dto.WxAppConfigResponse, *facade.Error) {
	var request dto.WxAppConfigRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}
	return c.configApplication.GetWxOpenConfig(ctx, &request)
}
