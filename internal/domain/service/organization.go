package service

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"

	"github.com/futurxlab/golanggraph/xerror"
)

type OrganizationService struct {
	organizationRepository     contract.IOrganizationRepository
	organizationUserRepository contract.IOrganizationUserRepository
}

func (service *OrganizationService) CreateRequest(ctx context.Context, request *entity.OrganizationRequestEntity) error {
	panic("implement me")
}

func (service *OrganizationService) CreateOrganization(
	ctx context.Context,
	organizationAggregate *aggregate.OrganizationAggregate) (*aggregate.OrganizationAggregate, error) {

	organizationAggregate, err := service.organizationRepository.Create(ctx, organizationAggregate)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return organizationAggregate, nil
}

func (service *OrganizationService) UpdateOrganization(
	ctx context.Context,
	organizationAggregate *aggregate.OrganizationAggregate) (*aggregate.OrganizationAggregate, error) {

	organizationAggregate, err := service.organizationRepository.Update(ctx, organizationAggregate)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return organizationAggregate, nil
}

func (service *OrganizationService) UpsertOrganizationUser(
	ctx context.Context,
	organizationAggregate *aggregate.OrganizationAggregate,
	userAggregate *aggregate.UserAggregate,
	roleAggregate *aggregate.RoleAggregate) (*aggregate.OrganizationUserAggregate, error) {

	organizationUserAggregate, err := service.organizationUserRepository.Find(ctx, userAggregate.User.ID, organizationAggregate.Organization.ID)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	if organizationUserAggregate == nil {
		organizationUserAggregate = &aggregate.OrganizationUserAggregate{
			Organization:     organizationAggregate.Organization,
			Application:      organizationAggregate.Application,
			User:             userAggregate.User,
			OrganizationRole: roleAggregate.Role,
		}

		organizationUserAggregate, err = service.organizationUserRepository.Create(ctx, organizationUserAggregate)
		if err != nil {
			return nil, xerror.Wrap(err)
		}

	} else {
		organizationUserAggregate.OrganizationRole = roleAggregate.Role

		organizationUserAggregate, err = service.organizationUserRepository.Update(ctx, organizationUserAggregate)
		if err != nil {
			return nil, xerror.Wrap(err)
		}
	}

	return organizationUserAggregate, nil
}

func (service *OrganizationService) DeleteOrganizationUser(
	ctx context.Context,
	organizationUserAggregate *aggregate.OrganizationUserAggregate) error {

	err := service.organizationUserRepository.Delete(ctx, organizationUserAggregate)
	if err != nil {
		return xerror.Wrap(err)
	}

	return nil
}

func NewOrganizationService(
	organizationRepository contract.IOrganizationRepository,
	organizationUserRepository contract.IOrganizationUserRepository,
) *OrganizationService {
	return &OrganizationService{
		organizationRepository:     organizationRepository,
		organizationUserRepository: organizationUserRepository,
	}
}
