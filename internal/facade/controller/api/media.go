package api

import (
	"kiwi-user/internal/facade/dto"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/gin-gonic/gin"
)

// GetWechatMedia godoc
// @Summary Get WeChat Media Resource
// @Tags Media
// @Description Get media resource from WeChat
// @Accept  json
// @Produce  json
// @Param  resource_id path string true "WeChat media resource ID"
// @Success 200 {object}  facade.BaseResponse{data=dto.WechatMediaResponse}
// @Router /v1/media/wechat/{resource_id} [get]
func (c *Controller) GetWechatMedia(ctx *gin.Context) (*dto.WechatMediaResponse, *facade.Error) {
	resourceID := ctx.Param("resource_id")
	if resourceID == "" {
		return nil, facade.ErrBadRequest.Facade("resource_id is required")
	}
	c.logger.Infof(ctx, "input resource_id %s", resourceID)
	response, err := c.mediaApplication.GetWechatMedia(ctx, resourceID)
	if err != nil {
		return nil, err
	}

	return response, nil
}
