package schema

import (
	"kiwi-user/internal/domain/model/enum"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type BindingVerify struct {
	ent.Schema
}

func (BindingVerify) Fields() []ent.Field {
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
		field.String("user_id").NotEmpty(),
		field.Enum("type").Values(convertStingerSliceToStringSlice(enum.GetAllBindingTypes())...),
		field.String("identity").NotEmpty(),
		field.String("code").NotEmpty(),
		field.Time("expires_at"),
		field.Time("verified_at").Default(time.Unix(0, 0)),
	}
}

func (BindingVerify) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "type", "identity").Unique(),
	}
}

func (BindingVerify) Edges() []ent.Edge {
	return []ent.Edge{}
}
