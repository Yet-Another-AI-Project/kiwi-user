package service

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/infrastructure/utils"
	"net/url"

	"github.com/futurxlab/golanggraph/xerror"
)

type UserService struct {
	userRepository          contract.IUserRepository
	bindingVerifyRepository contract.IBindingVerifyRepository
}

func (u *UserService) Update(
	ctx context.Context,
	user *aggregate.UserAggregate) (*aggregate.UserAggregate, error) {

	user, err := u.userRepository.Update(ctx, user)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return user, nil
}

func (u *UserService) UserRegisterWithPassword(
	ctx context.Context,
	applicationAggregate *aggregate.ApplicationAggregate,
	roleAggregate *aggregate.RoleAggregate,
	name,
	password string,
	verified bool) (*aggregate.UserAggregate, error) {

	// check if name already exists
	existingUser, err := u.userRepository.FindByName(ctx, applicationAggregate.Application.Name, name)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// create the user
	salt := utils.RandomSalt(name)
	identity, err := utils.EncodePassword(password, salt)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	userAggregate := &aggregate.UserAggregate{
		User: &entity.UserEntity{
			Name: name,
		},
		Application: applicationAggregate.Application,
		Bindings: []*entity.BindingEntity{
			{
				Type:     enum.BindingTypePassword,
				Identity: identity,
				Salt:     salt,
				Verified: verified,
			},
		},
	}

	if roleAggregate != nil {
		userAggregate.PersonalRole = roleAggregate.Role
	}

	userAggregate, err = u.userRepository.Create(ctx, userAggregate)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return userAggregate, nil
}

func (u *UserService) UpdateUserRole(
	ctx context.Context,
	user *aggregate.UserAggregate,
	role *aggregate.RoleAggregate) (*aggregate.UserAggregate, error) {

	user.PersonalRole = role.Role
	user, err := u.userRepository.Update(ctx, user)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return user, nil
}

func (u *UserService) UpdateUserInfo(
	ctx context.Context,
	user *aggregate.UserAggregate,
	displayname string,
	avatar string,
	department string) (*aggregate.UserAggregate, error) {

	if err := u.userRepository.WithTransaction(ctx, func(ctx context.Context) error {
		var err error

		// update name
		if displayname != "" {
			user.User.DisplayName = displayname
		}

		// update avatar
		if avatar != "" {
			parsedURI, err := url.ParseRequestURI(avatar)
			if err != nil {
				// TODO: upload avatar to oss and get url
				return xerror.Wrap(err)
			} else {
				user.User.Avatar = parsedURI.String()
			}
		}
		user.User.Department = department

		user, err = u.userRepository.Update(ctx, user)
		if err != nil {
			return xerror.Wrap(err)
		}

		return nil
	}); err != nil {
		return nil, xerror.Wrap(err)
	}

	return user, nil
}

func (u *UserService) UpdatePassword(
	ctx context.Context,
	user *aggregate.UserAggregate,
	newPassword string,
) (*aggregate.UserAggregate, error) {

	if err := u.userRepository.WithTransaction(ctx, func(ctx context.Context) error {
		var passwordBinding *entity.BindingEntity
		for _, binding := range user.Bindings {
			if binding.Type == enum.BindingTypePassword {
				passwordBinding = binding
				break
			}
		}

		if passwordBinding == nil {
			passwordBinding = &entity.BindingEntity{
				Type:     enum.BindingTypePassword,
				Verified: true,
			}
			user.Bindings = append(user.Bindings, passwordBinding)
		}

		salt := utils.RandomSalt(user.User.Name)
		hashedPassword, err := utils.EncodePassword(newPassword, salt)
		if err != nil {
			return xerror.Wrap(err)
		}

		passwordBinding.Salt = salt
		passwordBinding.Identity = hashedPassword
		passwordBinding.Verified = true

		user, err = u.userRepository.Update(ctx, user)
		if err != nil {
			return xerror.Wrap(err)
		}

		return nil
	}); err != nil {
		return nil, xerror.Wrap(err)
	}

	return user, nil
}

func NewUserService(
	userRepository contract.IUserRepository,
	bindingVerifyRepository contract.IBindingVerifyRepository) *UserService {
	return &UserService{
		userRepository:          userRepository,
		bindingVerifyRepository: bindingVerifyRepository,
	}
}
