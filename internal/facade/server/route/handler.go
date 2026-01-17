package route

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/Yet-Another-AI-Project/kiwi-lib/server/gin/utils"
	"github.com/gin-gonic/gin"
)

func RequireUserIDHandler[T any](f func(*gin.Context, string) (T, *facade.Error)) func(*gin.Context) {

	return func(c *gin.Context) {

		defer func() {
			if v := recover(); v != nil {
				utils.ResponseError(c, facade.ErrServerInternal)
				fmt.Println(v)
				debug.PrintStack()
			}
		}()

		userID, ok := c.Get("user_id")
		if !ok {
			utils.ResponseError(c, facade.ErrUnauthorized)
			return
		}

		data, err := f(c, userID.(string))
		if err != nil {
			utils.ResponseError(c, err)
			return
		}

		c.JSON(http.StatusOK, &facade.BaseResponse{
			Status: facade.StatusSuccess,
			Data:   data,
		})
	}
}

func NormalHandler[T any](f func(*gin.Context) (T, *facade.Error)) func(*gin.Context) {

	return func(c *gin.Context) {

		defer func() {
			if v := recover(); v != nil {
				err := facade.ErrServerInternal.Wrap(fmt.Errorf("%v\n%s", v, debug.Stack()))
				utils.ResponseError(c, err)
			}
		}()

		data, err := f(c)
		if err != nil {
			utils.ResponseError(c, err)
			return
		}

		c.JSON(http.StatusOK, &facade.BaseResponse{
			Status: "success",
			Data:   data,
		})
	}
}
