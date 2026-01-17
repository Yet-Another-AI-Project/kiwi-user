package domain

import (
	"kiwi-user/internal/domain/service"

	"go.uber.org/fx"
)

var Module = fx.Provide(
	service.NewApplicationService,
	service.NewDeviceService,
	service.NewLoginService,
	service.NewUserService,
	service.NewRBACService,
	service.NewOrganizationService,
	service.NewBindingService,
	service.NewPaymentService,
	service.NewOrganizationApplicationService,
	service.NewVertificationCodeService,
)
