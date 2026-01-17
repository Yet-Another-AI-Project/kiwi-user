package bootstrap

import (
	"context"
	"kiwi-user/config"
	"kiwi-user/internal/constants"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/domain/service"

	"github.com/futurxlab/golanggraph/logger"
	"github.com/futurxlab/golanggraph/xerror"
)

// Do some init stuff when the application starts at the first time
type Bootstrap struct {
	logger                logger.ILogger
	config                *config.Config
	applicationRepository contract.IApplicationRepository
	userRepository        contract.IUserRepository
	roleRepository        contract.IRoleRepository

	userService *service.UserService
}

func NewBootStrap(
	logger logger.ILogger,
	config *config.Config,
	applicationRepository contract.IApplicationRepository,
	userRepository contract.IUserRepository,
	roleRepository contract.IRoleRepository,
	userService *service.UserService,
) *Bootstrap {
	return &Bootstrap{
		logger:                logger,
		config:                config,
		applicationRepository: applicationRepository,
		userRepository:        userRepository,
		roleRepository:        roleRepository,
		userService:           userService,
	}
}

func (b *Bootstrap) Init() error {

	ctx := context.Background()

	err := b.applicationRepository.WithTransaction(ctx, func(ctx context.Context) error {
		existingApplication, err := b.applicationRepository.FindByName(ctx, constants.AdminApplicationName)
		if err != nil {
			return xerror.Wrap(err)
		}

		if existingApplication != nil {
			b.logger.Infof(ctx, "admin application already exists, skip bootstrap process")
			return nil
		}

		// Create Admin Application
		applicationAggregate, err := b.applicationRepository.Create(ctx, &aggregate.ApplicationAggregate{
			Application: &entity.ApplicationEntity{
				Name: constants.AdminApplicationName,
			},
		})

		if err != nil {
			return xerror.Wrap(err)
		}

		b.logger.Infof(ctx, "admin application created %+v", applicationAggregate)

		// Create Roles
		roleAggregate, err := b.roleRepository.Create(ctx, &aggregate.RoleAggregate{
			Application: applicationAggregate.Application,
			Role: &entity.RoleEntity{
				Type: enum.RoleTypePersonal,
				Name: constants.AdminPersonalRoleName,
			},
			Scopes: []*entity.ScopeEntity{
				{
					Name: constants.AdminPersonalScopeName,
				},
			},
		})

		if err != nil {
			return xerror.Wrap(err)
		}

		b.logger.Infof(ctx, "admin role created %+v", roleAggregate)

		// Create First User
		userAggregate, err := b.userService.UserRegisterWithPassword(
			ctx,
			applicationAggregate,
			roleAggregate,
			b.config.Bootstrap.FirstUserName,
			b.config.Bootstrap.FirstUserPass,
			true)

		if err != nil {
			return xerror.Wrap(err)
		}

		b.logger.Infof(ctx, "first user created %s", userAggregate.User.Name)

		return nil
	})

	if err != nil {
		return xerror.Wrap(err)
	}

	return nil
}
