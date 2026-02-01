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
		field.Enum("channel").Values(convertStingerSliceToStringSlice(enum.GetAllPaymentChannel())...),
		field.String("service").NotEmpty(),
		field.Int("amount").Positive(),
		field.String("currency").NotEmpty(),
		field.String("description").NotEmpty(),
		field.Enum("status").Values(convertStingerSliceToStringSlice(enum.GetAllPaymentStatus())...),
		field.Time("paid_at").Optional(),

		// channel info
		field.Enum("payment_type").
			Values(convertStingerSliceToStringSlice(enum.GetAllPaymentType())...).
			Default(string(enum.PaymentTypeOneTime)),
		field.String("wechat_platform").Optional(),
		field.String("wechat_open_id").Optional(),
		field.String("wechat_transaction_id").Optional(),
		field.String("stripe_subscription_id").Optional(),
		field.String("stripe_subscription_status").Optional(),
		field.String("stripe_interval").Optional(),
		field.Time("stripe_current_period_start").
			Optional(),
		field.Time("stripe_current_period_end").
			Optional(),
		field.String("stripe_customer_id").
			Optional(),
		field.String("stripe_customer_email").
			Optional(),
		field.String("stripe_checkout_session_id").
			Optional(),
		field.String("stripe_invoice_id").
			Optional(),
	}
}

func (Payment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("out_trade_no").Unique(),
		index.Fields("stripe_subscription_id"),
		index.Fields("stripe_customer_id"),
		index.Fields("stripe_checkout_session_id"),
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
