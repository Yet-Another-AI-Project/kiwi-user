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

type Role struct {
	ent.Schema
}

func (Role) Fields() []ent.Field {
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
		field.UUID("application_id", uuid.UUID{}),
		field.String("name").NotEmpty(),
		field.Enum("type").Values(convertStingerSliceToStringSlice(enum.GetAllRoleTypes())...),
	}
}

func (Role) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("application_id", "name").Unique(),
	}
}

func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("scopes", Scope.Type),

		edge.From("application", Application.Type).
			Ref("roles").
			Field("application_id").
			Unique().
			Required(),
	}
}
