package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type StripeEvent struct {
	ent.Schema
}

func (StripeEvent) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Immutable().
			Default(time.Now),
		field.String("event_id").NotEmpty().Unique(),
		field.String("event_type").NotEmpty(),
		field.Bool("processed").Default(false),
		field.Time("processed_at").Optional(),
	}
}

func (StripeEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_id").Unique(),
		index.Fields("event_type", "processed"),
	}
}
