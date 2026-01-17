package route

import (
	"kiwi-user/internal/constants"
	"kiwi-user/internal/facade/server/middleware"

	"github.com/gin-gonic/gin"
)

func (route *Route) RegisterAdmin(gin *gin.Engine) {

	// auth middlewares
	adminAuth := middleware.NewKiwiUserAuth(
		constants.AdminApplicationName,
		constants.AdminPersonalRoleName,
		route.jwtHepler)

	// admin apis
	admin := gin.Group("/admin", adminAuth)
	{
		admin.GET("/rbac/application", RequireUserIDHandler(route.adminController.GetApplication))
		admin.POST("/rbac/application", RequireUserIDHandler(route.adminController.CreateApplication))
		admin.PUT("/rbac/application/default-role", RequireUserIDHandler(route.adminController.UpdateApplicationDefaultRole))

		admin.POST("/rbac/role", RequireUserIDHandler(route.adminController.CreateRole))
		admin.POST("/rbac/scope", RequireUserIDHandler(route.adminController.CreateScope))

		// organization
		admin.POST("/organization", NormalHandler(route.adminController.CreateOrganization))
		admin.PUT("/organization", NormalHandler(route.adminController.UpdateOrganization))
		admin.GET("/organization/infos", NormalHandler(route.adminController.GetOrganizationInfos))
		admin.POST("/organization/user", NormalHandler(route.adminController.CreateOrganizationUser))
		admin.DELETE("/organization/user", NormalHandler(route.adminController.DeleteOrganizationUser))
		admin.GET("/organization/user/infos", NormalHandler(route.adminController.GetOrganizationUserInfos))

		admin.POST("/user/role", NormalHandler(route.adminController.CreateUserRole))
		admin.POST("/user/password", NormalHandler(route.adminController.CreateUserWithPassword))

		// organization application
		admin.GET("/organization_application/infos", NormalHandler(route.adminController.PageOrganizationApplication))
		admin.PUT("/organization_application/audit", NormalHandler(route.adminController.ReviewOrganizationApplication))
	}
}
