package repository

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/infrastructure/repository/ent"
	"kiwi-user/internal/infrastructure/repository/ent/organization"
	"kiwi-user/internal/infrastructure/repository/ent/organizationapplication"

	"entgo.io/ent/dialect/sql"
	"github.com/futurxlab/golanggraph/xerror"
	"github.com/google/uuid"
)

type organizationApplicationImpl struct {
	baseImpl
}

func (u *organizationApplicationImpl) FindByID(ctx context.Context, ID uuid.UUID) (*aggregate.OrganizationApplicationAggregate, error) {
	db := u.getEntClient(ctx)
	result, err := db.OrganizationApplication.Query().Where(organizationapplication.IDEQ(ID)).Only(ctx)
	if err != nil {
		return nil, err
	}
	applicationDO, err := result.QueryApplication().Only(ctx)
	if err != nil {
		return nil, err
	}

	return &aggregate.OrganizationApplicationAggregate{
		OrganizationApplication: convertOrganizationApplicationDoToEntity(result),
		Application:             convertApplicationDOToEntity(applicationDO),
	}, nil
}

func (u *organizationApplicationImpl) FindByUserID(ctx context.Context, request *entity.OrganizationApplicationEntity) ([]*aggregate.OrganizationApplicationAggregate, error) {
	db := u.getEntClient(ctx)

	baseQuery := db.OrganizationApplication.Query().Where(organizationapplication.UserIDEQ(request.UserID))

	if request.Name != "" {
		baseQuery = baseQuery.Where(organizationapplication.NameContains(request.Name))
	}

	organizationApplicationDOs, err := baseQuery.All(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	var result []*aggregate.OrganizationApplicationAggregate

	for _, oa := range organizationApplicationDOs {
		applicationDO, err := oa.QueryApplication().Only(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, &aggregate.OrganizationApplicationAggregate{
			OrganizationApplication: convertOrganizationApplicationDoToEntity(oa),
			Application:             convertApplicationDOToEntity(applicationDO),
		})
	}

	return result, nil
}

func (u *organizationApplicationImpl) PageFind(ctx context.Context, request *entity.OrganizationApplicationEntity, offset int, limit int) ([]*aggregate.OrganizationApplicationAggregate, int, error) {
	db := u.getEntClient(ctx)
	baseQuery := db.OrganizationApplication.Query()
	countBaseQuery := db.OrganizationApplication.Query()

	if request.Name != "" {
		baseQuery = baseQuery.Where(organizationapplication.NameContains(request.Name))
		countBaseQuery = countBaseQuery.Where(organizationapplication.NameContains(request.Name))
	}

	// 分页
	query := baseQuery.
		Offset(offset).
		Limit(limit).
		Order(
			organizationapplication.ByCreatedAt(
				sql.OrderDesc(),
			),
		)

	orgApplicationDOs, err := query.All(ctx)

	if err != nil {
		return nil, 0, xerror.Wrap(err)
	}

	count, err := countBaseQuery.Count(ctx)
	if err != nil {
		return nil, 0, xerror.Wrap(err)
	}

	var result []*aggregate.OrganizationApplicationAggregate

	for _, orgApplicationDO := range orgApplicationDOs {
		applicationDO, err := orgApplicationDO.QueryApplication().Only(ctx)
		if err != nil {
			return nil, 0, err
		}

		result = append(result, &aggregate.OrganizationApplicationAggregate{
			OrganizationApplication: convertOrganizationApplicationDoToEntity(orgApplicationDO),
			Application:             convertApplicationDOToEntity(applicationDO),
		})
	}

	return result, count, nil
}

func (u *organizationApplicationImpl) Create(ctx context.Context, orgAppEntity *entity.OrganizationApplicationEntity, appEntity *entity.ApplicationEntity) (*aggregate.OrganizationApplicationAggregate, error) {
	db := u.getEntClient(ctx)

	// todo : 考虑申请名称重名问题
	// 存在企业名称相同且状态不为审核失败的记录，说明重复申请
	//usedOrganizationApplication, err := db.OrganizationApplication.Query().Where(organizationapplication.NameEQ(orgAppEntity.Name)).Where(organizationapplication.ReviewStatusNEQ(organizationapplication.ReviewStatus(enum.OrganizationRequestStatusReject))).Only(ctx)
	//if err != nil && !ent.IsNotFound(err) {
	//	return nil, err
	//}
	//if usedOrganizationApplication != nil {
	//	return nil, xerror.New("repeat request already exist in organization_application")
	//}

	// 企业名称不能相同
	usedOrganization, err := db.Organization.Query().Where(organization.Name(orgAppEntity.Name)).Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}
	if usedOrganization != nil {
		return nil, xerror.New("name already exist in organization")
	}

	organizationApplication, err := db.OrganizationApplication.Create().
		SetName(orgAppEntity.Name).
		SetStatus(organizationapplication.StatusTrial).
		SetApplicationID(appEntity.ID).
		SetBrandShortName(orgAppEntity.BrandShortName).
		SetPrimaryBusiness(orgAppEntity.PrimaryBusiness).
		SetUsageScenario(orgAppEntity.UsageScenario).
		SetDiscoveryWay(orgAppEntity.DiscoveryWay).
		SetReferrerName(orgAppEntity.ReferrerName).
		SetTrialDays(orgAppEntity.TrailDays).
		SetReviewStatus(organizationapplication.ReviewStatusPending).
		SetUserID(orgAppEntity.UserID).
		SetOrgRoleName(orgAppEntity.OrgRoleName).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return &aggregate.OrganizationApplicationAggregate{
		OrganizationApplication: convertOrganizationApplicationDoToEntity(organizationApplication),
	}, nil
}

func (u *organizationApplicationImpl) Update(ctx context.Context, request *entity.OrganizationApplicationEntity) (*aggregate.OrganizationApplicationAggregate, error) {
	db := u.getEntClient(ctx)

	baseUpdate := db.OrganizationApplication.UpdateOneID(request.ID).
		SetReviewStatus(organizationapplication.ReviewStatus(request.ReviewStatus)).
		SetReviewComment(request.ReviewComment)

	organizationApplication, err := baseUpdate.Save(ctx)

	if err != nil {
		return nil, err
	}

	return &aggregate.OrganizationApplicationAggregate{
		OrganizationApplication: convertOrganizationApplicationDoToEntity(organizationApplication),
	}, nil
}

func NewOrganizationApplicationImpl(db *Client) contract.IOrganizationApplicationRepository {
	return &organizationApplicationImpl{
		baseImpl{
			db: db,
		},
	}
}
