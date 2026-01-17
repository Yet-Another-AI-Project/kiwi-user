package application

import "go.uber.org/fx"

var Module = fx.Provide(
	NewLoginApplication,
	NewRBACApplication,
	NewTokenApplication,
	NewUserApplication,
	NewOrganizationApplication,
	NewBindingApplication,
	NewPaymentApplication,
	NewOrganizationRequestApplication,
	NewConfigApplication,
	NewMediaApplication,
)
