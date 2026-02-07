package api

import (
	"kiwi-user/internal/application"

	"github.com/futurxlab/golanggraph/logger"
)

// @title FuturxUser Service API
// @Version 0.0.1
// @description FuturxUser Service API
//
// @Contact.Name API Support
//
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
type Controller struct {
	loginApplication                   *application.LoginApplication
	rbacApplication                    *application.RBACApplication
	tokenApplication                   *application.TokenApplication
	userApplication                    *application.UserApplication
	organizationApplication            *application.OrganizationApplication
	bindingApplication                 *application.BindingApplication
	paymentApplication                 *application.PaymentApplication
	organizationApplicationApplication *application.OrganizationApplicationApplication
	logger                             logger.ILogger
}

func NewController(
	loginApplication *application.LoginApplication,
	rbacApplication *application.RBACApplication,
	tokenApplication *application.TokenApplication,
	userApplication *application.UserApplication,
	organizationApplication *application.OrganizationApplication,
	bindingApplication *application.BindingApplication,
	paymentApplication *application.PaymentApplication,
	organizationApplicationApplication *application.OrganizationApplicationApplication,
	logger logger.ILogger,
) (*Controller, error) {
	return &Controller{
		loginApplication:                   loginApplication,
		rbacApplication:                    rbacApplication,
		tokenApplication:                   tokenApplication,
		userApplication:                    userApplication,
		organizationApplication:            organizationApplication,
		bindingApplication:                 bindingApplication,
		paymentApplication:                 paymentApplication,
		organizationApplicationApplication: organizationApplicationApplication,
		logger:                             logger,
	}, nil
}
