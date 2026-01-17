package service

import (
	"context"
	"fmt"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	"time"

	"github.com/futurxlab/golanggraph/xerror"
)

type OrganizationApplicationService struct {
	organizationApplicationRepository contract.IOrganizationApplicationRepository
	organizationUserRepository        contract.IOrganizationUserRepository
	organizationRepository            contract.IOrganizationRepository
	userReadRepository                contract.IUserReadRepository
	roleReadRepository                contract.IRoleReadRepository
}

func (service *OrganizationApplicationService) PageFind(ctx context.Context, name string, offset, limit int) ([]*aggregate.OrganizationApplicationAggregate, int, error) {

	wwsas, count, err := service.organizationApplicationRepository.PageFind(ctx, &entity.OrganizationApplicationEntity{
		Name: name,
	}, offset, limit)

	if err != nil {
		return nil, 0, xerror.Wrap(err)
	}

	return wwsas, count, nil
}

func (service *OrganizationApplicationService) CreateOrganizationApplication(ctx context.Context, orgAppEntity *entity.OrganizationApplicationEntity, appEntity *entity.ApplicationEntity) error {
	_, err := service.organizationApplicationRepository.Create(ctx, orgAppEntity, appEntity)

	if err != nil {
		return xerror.Wrap(err)
	}

	return nil
}

func (service *OrganizationApplicationService) GetOrganizationApplicationAggregates(ctx context.Context, userId string, name string) ([]*aggregate.OrganizationApplicationAggregate, error) {
	organizationApplicationAggregates, err := service.organizationApplicationRepository.FindByUserID(ctx, &entity.OrganizationApplicationEntity{
		UserID: userId,
		Name:   name,
	})

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return organizationApplicationAggregates, nil
}

func (service *OrganizationApplicationService) UpdateOrganizationApplication(
	ctx context.Context,
	request *entity.OrganizationApplicationEntity) (*aggregate.OrganizationApplicationAggregate, error) {
	oaa, err := service.organizationApplicationRepository.Update(ctx, request)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return oaa, nil
}

func (service *OrganizationApplicationService) UpdateOrganizationApplicationWithTx(
	ctx context.Context,
	request *entity.OrganizationApplicationEntity) (*aggregate.OrganizationApplicationAggregate, string, error) {
	var oaa *aggregate.OrganizationApplicationAggregate

	// 查询当前审核条目
	orgAppAggregate, err := service.organizationApplicationRepository.FindByID(ctx, request.ID)
	if err != nil {
		return nil, "", err
	}

	if orgAppAggregate.OrganizationApplication == nil {
		return nil, "", fmt.Errorf("record not found")
	}

	// 审核通过不允许二次审核
	if orgAppAggregate.OrganizationApplication.ReviewStatus == enum.OrganizationRequestStatusApproved {
		return nil, "", fmt.Errorf("record was already audit")
	}

	// 查询用户信息
	userAggregate, err := service.userReadRepository.Find(ctx, orgAppAggregate.OrganizationApplication.UserID)
	if err != nil {
		return nil, "", err
	}

	if userAggregate == nil {
		return nil, "", err
	}
	// 查询是否绑定过手机，没有则返回错误
	phoneExist := false
	var phone string
	for _, binding := range userAggregate.Bindings {
		if binding.Type == enum.BindingTypePhone && binding.Verified == true && binding.Identity != "" {
			phoneExist = true
			phone = binding.Identity
			break
		}
	}
	if !phoneExist {
		return nil, "", fmt.Errorf("user not bind any phone")
	}

	// 开启事务
	if err := service.organizationApplicationRepository.WithTransaction(ctx, func(ctx context.Context) error {
		var err error
		// 审核通过
		if request.ReviewStatus == enum.OrganizationRequestStatusApproved {
			// 创建企业，绑定用户到企业上
			// 1.创建企业（organization）
			oa := &aggregate.OrganizationAggregate{
				Organization: &entity.OrganizationEntity{
					Name: orgAppAggregate.OrganizationApplication.Name,
					// 试用版
					Status: enum.OrganizationStatusTrial,
					// 默认7天试用，后期更新有效期放在组织管理中
					ExpiresAt: time.Now().AddDate(0, 0, 7),
				},
				Application: orgAppAggregate.Application,
			}
			organizationAggregate, err := service.organizationRepository.Create(ctx, oa)
			if err != nil {
				return err
			}

			// 2.绑定企业用户
			// 2.1 查询当前的用户和企业是否已绑定，如果未绑定，则建立绑定，否则返回错误
			organizationUserAggregate, err := service.organizationUserRepository.Find(ctx, orgAppAggregate.OrganizationApplication.UserID, organizationAggregate.Organization.ID)

			if err != nil {
				return err
			}

			// 已绑定，返回失败
			if organizationUserAggregate != nil {
				return fmt.Errorf("already bind error")
			}

			// 获取org_role
			ra, err := service.roleReadRepository.FindByName(ctx, organizationAggregate.Application.Name, orgAppAggregate.OrganizationApplication.OrgRoleName)

			if err != nil {
				return err
			}
			if ra == nil || ra.Role == nil {
				return fmt.Errorf("org role is wrong")
			}

			// 2.2 绑定用户和企业
			organizationUserAggregate = &aggregate.OrganizationUserAggregate{
				Organization:     organizationAggregate.Organization,
				Application:      organizationAggregate.Application,
				User:             userAggregate.User,
				OrganizationRole: ra.Role,
			}

			organizationUserAggregate, err = service.organizationUserRepository.Create(ctx, organizationUserAggregate)
			if err != nil {
				return err
			}
		}
		// 3. 更新企业申请信息
		oaaNew, err := service.organizationApplicationRepository.Update(ctx, request)

		if err != nil {
			return err
		}

		oaa = &aggregate.OrganizationApplicationAggregate{
			OrganizationApplication: oaaNew.OrganizationApplication,
		}

		return nil
	}); err != nil {
		return nil, "", xerror.Wrap(err)
	}

	return oaa, phone, nil
}

func NewOrganizationApplicationService(
	organizationApplicationRepository contract.IOrganizationApplicationRepository,
	organizationUserRepository contract.IOrganizationUserRepository,
	organizationRepository contract.IOrganizationRepository,
	userReadRepository contract.IUserReadRepository,
	roleReadRepository contract.IRoleReadRepository,
) *OrganizationApplicationService {
	return &OrganizationApplicationService{
		organizationApplicationRepository: organizationApplicationRepository,
		organizationUserRepository:        organizationUserRepository,
		organizationRepository:            organizationRepository,
		userReadRepository:                userReadRepository,
		roleReadRepository:                roleReadRepository,
	}
}
