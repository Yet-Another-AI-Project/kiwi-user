package schema

import (
	"kiwi-user/internal/domain/model/enum"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type WechatOpenID struct {
	ent.Schema
}

func (WechatOpenID) Fields() []ent.Field {
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
		field.String("open_id").NotEmpty(),
		field.String("user_id").NotEmpty(),
		field.Enum("platform").Values(convertStingerSliceToStringSlice(enum.GetAllWechatOpenIDPlatform())...),
		field.String("union_id").Optional(),
	}
}

func (WechatOpenID) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("open_ids").
			Unique().
			Required().
			Field("user_id"),
	}
}

func (WechatOpenID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "platform").
			Unique(),
		index.Fields("user_id"),
	}
}
