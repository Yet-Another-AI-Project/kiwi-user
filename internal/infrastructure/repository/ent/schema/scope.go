package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Scope struct {
	ent.Schema
}

func (Scope) Fields() []ent.Field {
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
		field.UUID("role_id", uuid.UUID{}),
		field.String("name").NotEmpty(),
		field.Bool("hidden_in_jwt").Default(false),
	}
}

func (Scope) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_id", "name").Unique(),
	}
}

func (Scope) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("role", Role.Type).
			Ref("scopes").
			Field("role_id").
			Unique().
			Required(),
	}
}
