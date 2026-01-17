package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Device struct {
	ent.Schema
}

func (Device) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Time("created_at").
			Immutable().
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("deleted_at").Optional(),
		field.String("user_id"),
		field.UUID("organization_id", uuid.UUID{}).Optional(),
		field.String("device_type").NotEmpty(),
		field.String("device_id").NotEmpty(),
		field.String("refresh_token").NotEmpty(),
		field.Time("refresh_token_expires_at").Default(timeOneDayLater),
	}
}

func (Device) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("devices").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (Device) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("refresh_token").Unique(),
		index.Fields("user_id", "device_type", "device_id").Unique(),
	}
}

func timeOneDayLater() time.Time {
	return time.Now().Add(time.Hour * 24)
}
