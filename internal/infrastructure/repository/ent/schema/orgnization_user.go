package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type OrganizationUser struct {
	ent.Schema
}

func (OrganizationUser) Fields() []ent.Field {
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
		field.UUID("organization_id", uuid.UUID{}),
		field.String("user_id"),
	}
}

func (OrganizationUser) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("organization", Organization.Type).
			Field("organization_id").
			Required().
			Unique(),
		edge.To("user", User.Type).
			Field("user_id").
			Required().
			Unique(),
		edge.To("role", Role.Type).Unique(),
	}
}

func (OrganizationUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "organization_id").Unique(),
	}
}
