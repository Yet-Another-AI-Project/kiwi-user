package api

import (
	"kiwi-user/internal/facade/dto"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"

	"github.com/gin-gonic/gin"
)

func (c *Controller) Ping(ctx *gin.Context) (*dto.OperationResponse, *facade.Error) {

	return &dto.OperationResponse{
		Success: true,
	}, nil
}
