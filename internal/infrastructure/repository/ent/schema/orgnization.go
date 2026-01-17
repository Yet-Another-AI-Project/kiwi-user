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

type Organization struct {
	ent.Schema
}

func (Organization) Fields() []ent.Field {
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
		field.String("name").NotEmpty(),
		field.Enum("status").Values(convertStingerSliceToStringSlice(enum.GetAllOrganizationStatus())...),
		field.Time("refresh_at").Optional(),
		field.Time("expires_at").Optional(),
		field.String("permission_code").Default("init_permission_code"),
		field.String("logo_url").Default("").Comment("staff显示的logo url"),
	}
}

func (Organization) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("application", Application.Type).
			Field("application_id").
			Ref("organizations").
			Required().
			Unique(),
	}
}

func (Organization) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("application_id", "name").Unique(),
	}
}
