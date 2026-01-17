package schema

import (
	"kiwi-user/internal/domain/model/enum"
	"time"

	"entgo.io/ent/schema/edge"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type OrganizationApplication struct {
	ent.Schema
}

func (OrganizationApplication) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(func() uuid.UUID {
				id := uuid.Must(uuid.NewV7())
				return id
			}),
		field.Time("created_at").
			Immutable().
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("deleted_at").Optional(),
		field.UUID("application_id", uuid.UUID{}),
		field.String("name").NotEmpty().Comment("企业名称"),
		field.Enum("status").Values(convertStingerSliceToStringSlice(enum.GetAllOrganizationStatus())...),
		field.Int32("trial_days").Optional().Comment("试用天数"),
		field.Enum("review_status").Values(convertStingerSliceToStringSlice(enum.GetAllOrganizationRequestStatus())...),
		field.String("review_comment").Default("").MaxLen(2000).Comment("审核描述"),
		field.String("user_id").NotEmpty(),
		field.String("brand_short_name").NotEmpty().Comment("企业简称"),
		field.String("primary_business").NotEmpty().MaxLen(5000).Comment("主营业务"),
		field.String("usage_scenario").NotEmpty().MaxLen(5000).Comment("使用诉求"),
		field.String("referrer_name").Optional().Comment("推荐人"),
		field.String("discovery_way").Optional().Comment("发现途径"),
		field.String("org_role_name").NotEmpty().Comment("组织角色名称"),
	}
}

func (OrganizationApplication) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("application", Application.Type).
			Field("application_id").
			Ref("organization_application").
			Required().
			Unique(),
	}
}

func (OrganizationApplication) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("application_id", "name").Unique(),
	}
}
