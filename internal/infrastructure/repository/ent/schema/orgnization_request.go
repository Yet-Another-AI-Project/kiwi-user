package schema

import (
	"kiwi-user/internal/domain/model/enum"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type OrganizationRequest struct {
	ent.Schema
}

func (OrganizationRequest) Fields() []ent.Field {
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
		field.Enum("type").Values(convertStingerSliceToStringSlice(enum.GetAllOrganizationRequestTypes())...),
		field.Enum("status").Values(convertStingerSliceToStringSlice(enum.GetAllOrganizationRequestStatus())...),
		field.String("user_id"),
		field.UUID("organization_id", uuid.UUID{}),
		field.String("organization_name"),
	}
}

func (OrganizationRequest) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (OrganizationRequest) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("organization_id").Annotations(entsql.Prefix(18)),
	}
}
