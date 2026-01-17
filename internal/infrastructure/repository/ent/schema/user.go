package schema

import (
	"kiwi-user/internal/domain/model/entity"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
		field.Time("created_at").
			Immutable().
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("deleted_at").Optional(),
		field.UUID("application_id", uuid.UUID{}),
		field.String("name").NotEmpty(),
		field.String("display_name").Optional(),
		field.String("avatar").Optional(),
		field.JSON("referral_channel", entity.UserRefferalChannel{}).Optional(),
		// 增加 "部门" 字段
		field.String("department").
			Default(""). // 默认为空字符串
			Optional().  // 设为可选，使其在数据库中 NULLable
			Comment("部门"),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("application_id", "name").Unique(),
		index.Fields("name"),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("bindings", Binding.Type),
		edge.To("devices", Device.Type),
		edge.To("personal_role", Role.Type).Unique(),
		edge.To("open_ids", WechatOpenID.Type),
		edge.To("qy_wechat_user_ids", QyWechatUserID.Type),
		edge.To("payments", Payment.Type),

		edge.From("application", Application.Type).
			Ref("users").
			Field("application_id").
			Unique().
			Required(),

		edge.From("organization_users", OrganizationUser.Type).
			Ref("user"),
	}
}
