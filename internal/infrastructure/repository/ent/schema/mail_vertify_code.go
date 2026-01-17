package schema

import (
	"kiwi-user/internal/domain/model/enum"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type MailVertifyCode struct {
	ent.Schema
}

func (MailVertifyCode) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("deleted_at").Optional(),
		field.Enum("type").Values(convertStingerSliceToStringSlice(enum.GetAllVertificationCodeTypes())...),
		field.String("email").NotEmpty(),
		field.String("code").NotEmpty(),
		field.Time("expires_at"),
	}
}

func (MailVertifyCode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type", "email").Unique(),
	}
}
