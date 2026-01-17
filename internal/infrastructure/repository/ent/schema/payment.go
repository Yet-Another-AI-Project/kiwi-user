package schema

import (
	"kiwi-user/internal/domain/model/enum"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Payment struct {
	ent.Schema
}

func (Payment) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Immutable().
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.String("out_trade_no").NotEmpty(),
		field.String("user_id").NotEmpty(),
		field.String("transaction_id").Optional(),
		field.String("open_id").Optional(),
		field.Enum("channel").Values(convertStingerSliceToStringSlice(enum.GetAllPaymentChannel())...),
		field.Enum("platform").Values(convertStingerSliceToStringSlice(enum.GetAllWechatOpenIDPlatform())...),
		field.String("service").NotEmpty(),
		field.Int("amount").Positive(),
		field.String("currency").NotEmpty(),
		field.String("description").NotEmpty(),
		field.Enum("status").Values(convertStingerSliceToStringSlice(enum.GetAllPaymentStatus())...),
		field.Time("paid_at").Optional(),
	}
}

func (Payment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("out_trade_no").Unique(),
	}
}

func (Payment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Field("user_id").
			Ref("payments").
			Unique().
			Required(),
	}
}
