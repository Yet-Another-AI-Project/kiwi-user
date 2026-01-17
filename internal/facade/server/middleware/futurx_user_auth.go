package middleware

import (
	"kiwi-user/internal/infrastructure/jwt"
	"net/http"
	"strings"
	"time"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/Yet-Another-AI-Project/kiwi-lib/server/gin/utils"
	"github.com/futurxlab/golanggraph/xerror"
	"github.com/gin-gonic/gin"
)

func getAccessToken(r *http.Request) (string, *facade.Error) {

	if cookie, err := r.Cookie("access_token"); err == nil {
		jwt := cookie.Value
		return jwt, nil
	} else if err != http.ErrNoCookie {
		return "", facade.ErrServerInternal.Wrap(err)
	}

	auth := strings.Fields(r.Header.Get("Authorization"))

	if len(auth) != 2 {
		return "", facade.ErrUnauthorized
	}

	if strings.ToLower(auth[0]) != "jwt" && strings.ToLower(auth[0]) != "bearer" {
		return "", facade.ErrUnauthorized
	}

	return auth[1], nil
}

func NewKiwiUserAuth(application, role string, jwtHelper *jwt.JWTHelper) func(*gin.Context) {

	return func(c *gin.Context) {

		token, ferr := getAccessToken(c.Request)
		if ferr != nil {
			utils.ResponseError(c, ferr)
			return
		}

		jwtToken, err := jwtHelper.VerifyRS256JWT(token)
		if xerror.Is(err, jwt.ErrInvalidJWTToken) {
			utils.ResponseError(c, facade.ErrUnauthorized)
			return
		}

		payload, err := jwtHelper.DecodeAccessPayload(jwtToken.Payload)
		if err != nil {
			utils.ResponseError(c, facade.ErrUnauthorized.Wrap(err))
			return
		}

		expire := time.Unix(payload.Expire, 0)
		if expire.Before(time.Now()) {
			utils.ResponseError(c, facade.ErrUnauthorized)
			return
		}

		// check issuer and role
		if application != "" && role != "'" {
			if payload.Application != application || payload.PersonalRole != role {
				utils.ResponseError(c, facade.ErrUnauthorized)
				return
			}
		}

		c.Set("user_id", payload.UserID)
		c.Set("org_id", payload.OrganizationID)

		c.Next()
	}
}
