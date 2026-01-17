package service

import "errors"

var (
	// application
	ErrApplicationAlreadyExists = errors.New("application already exists")
	ErrApplicationInvalidName   = errors.New("application name is invalid")
	ErrApplicationNotFound      = errors.New("application not found")

	// device
	ErrDeviceNotFound = errors.New("device not found")

	// login
	ErrInvalidWechatCode     = errors.New("invalid wechat code")
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrUserNameAlreadyExists = errors.New("user name already exists")
	ErrWechatInvalidScope    = errors.New("wechat access_token scope not found or invalid")

	// rbac
	ErrRoleNotFound       = errors.New("role not found")
	ErrRoleAlreadyExists  = errors.New("role already exists")
	ErrScopeAlreadyExists = errors.New("scope already exists")

	// organization
	ErrOrganizationAlreadyExists = errors.New("organization already exists")

	// binding verify
	ErrBindingVerifyAlreadyExists = errors.New("binding verify already exists")
)
