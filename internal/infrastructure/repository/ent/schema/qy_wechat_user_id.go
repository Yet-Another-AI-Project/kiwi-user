package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type QyWechatUserID struct {
	ent.Schema
}

func (QyWechatUserID) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.Time("created_at").
			Immutable().
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("deleted_at").Optional(),
		field.String("qy_wechat_user_id").Optional(),
		field.String("user_id").NotEmpty(),
		field.String("open_id").Optional(),
	}
}

func (QyWechatUserID) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("qy_wechat_user_ids").
			Unique().
			Required().
			Field("user_id"),
	}
}

func (QyWechatUserID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").
			Unique(),
		index.Fields("qy_wechat_user_id").
			Unique(),
	}
}
