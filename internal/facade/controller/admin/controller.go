package admin

import (
	"kiwi-user/internal/application"
)

type Controller struct {
	rbacApplication                    *application.RBACApplication
	organizationApplication            *application.OrganizationApplication
	organizationApplicationApplication *application.OrganizationApplicationApplication
	userApplication                    *application.UserApplication
}

func NewController(
	rbacApplication *application.RBACApplication,
	organizationApplication *application.OrganizationApplication,
	organizationApplicationApplication *application.OrganizationApplicationApplication,
	userApplication *application.UserApplication,
) (*Controller, error) {
	return &Controller{
		rbacApplication:                    rbacApplication,
		organizationApplication:            organizationApplication,
		organizationApplicationApplication: organizationApplicationApplication,
		userApplication:                    userApplication,
	}, nil
}
