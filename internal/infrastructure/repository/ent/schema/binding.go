package schema

import (
	"kiwi-user/internal/domain/model/enum"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Binding struct {
	ent.Schema
}

func (Binding) Fields() []ent.Field {
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
		field.Enum("type").Values(convertStingerSliceToStringSlice(enum.GetAllBindingTypes())...),
		field.String("identity").NotEmpty(),
		field.String("email").Optional(),
		field.Bool("verified").Default(false),
		field.String("salt").Optional(),
		field.String("user_id"),
		field.UUID("application_id", uuid.UUID{}).Optional(),
	}
}

func (Binding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type", "identity").Annotations(entsql.PrefixColumn("identity", 10)),
		index.Fields("user_id", "type").Unique(),
		// index.Fields("application_id", "type", "identity").Unique(),
	}
}

func (Binding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("bindings").
			Field("user_id").
			Unique().
			Required(),
	}
}
