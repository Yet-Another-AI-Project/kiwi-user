package service

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"

	"github.com/futurxlab/golanggraph/xerror"
)

type ApplicationService struct {
	applicationRepository contract.IApplicationRepository
}

func NewApplicationService(applicationRepository contract.IApplicationRepository) *ApplicationService {
	return &ApplicationService{
		applicationRepository: applicationRepository,
	}
}

func (a *ApplicationService) CreateApplication(
	ctx context.Context,
	application *aggregate.ApplicationAggregate) (*aggregate.ApplicationAggregate, error) {

	if application.Application.Name == "" {
		return nil, xerror.Wrap(ErrApplicationInvalidName)
	}

	existingApplication, err := a.applicationRepository.FindByName(ctx, application.Application.Name)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	if existingApplication != nil {
		return nil, xerror.Wrap(ErrApplicationAlreadyExists)
	}

	application, err = a.applicationRepository.Create(ctx, application)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return application, nil
}

func (a *ApplicationService) UpdateApplication(
	ctx context.Context,
	application *aggregate.ApplicationAggregate) (*aggregate.ApplicationAggregate, error) {

	existingApplication, err := a.GetApplication(ctx, application.Application.Name)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	if existingApplication == nil {
		return nil, xerror.Wrap(ErrApplicationNotFound)
	}

	// only update default role
	existingApplication.DefaultPersonalRole = application.DefaultPersonalRole
	existingApplication.DefaultOrgRole = application.DefaultOrgRole
	existingApplication.DefaultOrgAdminRole = application.DefaultOrgAdminRole

	application, err = a.applicationRepository.Update(ctx, existingApplication)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return application, nil
}

func (a *ApplicationService) GetApplication(ctx context.Context, name string) (*aggregate.ApplicationAggregate, error) {
	if name == "" {
		return nil, xerror.Wrap(ErrApplicationInvalidName)
	}

	application, err := a.applicationRepository.FindByName(ctx, name)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	if application == nil {
		return nil, xerror.Wrap(ErrApplicationNotFound)
	}

	return application, nil
}

func (a *ApplicationService) DeleteApplication(ctx context.Context, name string) error {
	if name == "" {
		return xerror.Wrap(ErrApplicationInvalidName)
	}

	application, err := a.applicationRepository.FindByName(ctx, name)
	if err != nil {
		return xerror.Wrap(err)
	}

	if application == nil {
		return xerror.Wrap(ErrApplicationNotFound)
	}

	if err := a.applicationRepository.Delete(ctx, application); err != nil {
		return xerror.Wrap(err)
	}

	return nil
}
