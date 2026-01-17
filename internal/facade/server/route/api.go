package route

import (
	"github.com/gin-gonic/gin"

	"kiwi-user/internal/facade/server/middleware"
)

// RegisterApiV1 register route for api server
func (route *Route) RegisterApiV1(gin *gin.Engine) {

	userAuth := middleware.NewKiwiUserAuth(
		"",
		"",
		route.jwtHepler)

	gin.GET("/ping", NormalHandler(route.apiController.Ping))

	v1 := gin.Group("/v1")

	login := v1.Group("/login")
	{
		// login.POST("/wechat/miniprogram", NormalHandler(route.apiController.WechatMiniProgramLogin))
		login.POST("/wechat/web", NormalHandler(route.apiController.WechatWebLogin))
		// login.POST("/qy_wechat", NormalHandler(route.apiController.QyWechatLogin))
		login.POST("/password", NormalHandler(route.apiController.PasswordLogin))
		// login.POST("/organization", NormalHandler(route.apiController.OrganizationLogin))
		// login.POST("/phone", NormalHandler(route.apiController.PhoneLogin))
		// login.POST("/phone/verify_code", NormalHandler(route.apiController.SendPhoneVerifyCode))
		// login.POST("/phone/captcha/verify_code", NormalHandler(route.apiController.SendPhoneVerifyCodeWithCaptcha))
		// login.POST("/email", NormalHandler(route.apiController.EmailLogin))
		// login.POST("/email/verify_code", NormalHandler(route.apiController.SendEmailVerificationCode))
		// login.POST("/email/captcha/verify_code", NormalHandler(route.apiController.SendEmailVerificationCodeWithCaptcha))
		login.POST("/google/web", NormalHandler(route.apiController.GoogleWebLogin))
	}

	token := v1.Group("/token")
	{
		token.POST("/verify", NormalHandler(route.apiController.VerifyAccessToken))
		token.GET("/publickey", NormalHandler(route.apiController.GetPublickKey))
		token.POST("/refresh", NormalHandler(route.apiController.RefreshAccessToken))
	}

	user := v1.Group("/user")
	{
		user.GET("/info", userAuth, RequireUserIDHandler(route.apiController.GetUserInfo))
		user.PUT("/info", userAuth, RequireUserIDHandler(route.apiController.UpdateUserInfo))
		// user.POST("/password", userAuth, RequireUserIDHandler(route.apiController.ChangePassword))
		// user.POST("/binding/phone", userAuth, RequireUserIDHandler(route.apiController.BindingPhoneWithMiniProgramCode))
		// user.POST("/binding/phone/verify_code", userAuth, RequireUserIDHandler(route.apiController.BindingPhoneWithVerifyCode))
		// organization application
		user.GET("/organization_application/infos", userAuth, RequireUserIDHandler(route.apiController.GetOrganizationApplicationInfos))
		user.POST("/organization_application/request", userAuth, RequireUserIDHandler(route.apiController.CreateOrganizationApplication))
		user.POST("/logout", userAuth, RequireUserIDHandler(route.apiController.Logout))
	}

	// payment := v1.Group("/payments")
	// {
	// 	payment.POST("", NormalHandler(route.apiController.CreatePayment))
	// 	payment.GET("/:out-trade-no/status", NormalHandler(route.apiController.QueryPaymentStatus))
	// 	payment.POST("/wechat/notify", NormalHandler(route.apiController.WechatPaymentCallback))
	// }

	// internal apis
	internal := gin.Group("/internal")
	{
		internal.POST("/user/infos", NormalHandler(route.apiController.GetPublicUserInfos))
		internal.POST("/organization/infos", NormalHandler(route.apiController.GetOrganizationInfos))
		internal.GET("/getCurrentInfos", userAuth, NormalHandler(route.apiController.GetCurrentInfos))
	}
}
