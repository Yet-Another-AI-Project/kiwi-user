package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Application struct {
	ent.Schema
}

func (Application) Fields() []ent.Field {
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
		field.String("name").NotEmpty(),
	}
}

func (Application) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Unique(),
	}
}

func (Application) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("users", User.Type),
		edge.To("organizations", Organization.Type),
		edge.To("roles", Role.Type),
		edge.To("default_personal_role", Role.Type).Unique(),
		edge.To("default_org_role", Role.Type).Unique(),
		edge.To("default_org_admin_role", Role.Type).Unique(),
		edge.To("organization_application", OrganizationApplication.Type),
	}
}
