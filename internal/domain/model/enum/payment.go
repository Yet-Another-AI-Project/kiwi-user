package enum

type PaymentChannel string

const (
	PaymentChannelWechat PaymentChannel = "wechat"
	PaymentChannelStripe PaymentChannel = "stripe"
)

type PaymentType string

const (
	PaymentTypeOneTime      PaymentType = "one_time"
	PaymentTypeSubscription PaymentType = "subscription"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusPastDue  SubscriptionStatus = "past_due"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
	SubscriptionStatusUnpaid   SubscriptionStatus = "unpaid"
)

type SubscriptionInterval string

const (
	SubscriptionIntervalMonthly SubscriptionInterval = "monthly"
	SubscriptionIntervalYearly  SubscriptionInterval = "yearly"
)

type PaymentStatus string

const (
	PaymentStatusNotPay  PaymentStatus = "NOTPAY"
	PaymentStatusSuccess PaymentStatus = "SUCCESS"
	PaymentStatusClosed  PaymentStatus = "CLOSED"
	PaymentStatusRefund  PaymentStatus = "REFUND"
)

func (p PaymentChannel) String() string {
	return string(p)
}

func (p PaymentStatus) String() string {
	return string(p)
}

func GetAllPaymentChannel() []PaymentChannel {
	return []PaymentChannel{
		PaymentChannelWechat,
		PaymentChannelStripe,
	}
}

func GetAllPaymentStatus() []PaymentStatus {
	return []PaymentStatus{
		PaymentStatusNotPay,
		PaymentStatusSuccess,
		PaymentStatusClosed,
		PaymentStatusRefund,
	}
}

func ParsePaymentChannel(channel string) PaymentChannel {
	switch channel {
	case "wechat":
		return PaymentChannelWechat
	case "stripe":
		return PaymentChannelStripe
	default:
		return ""
	}
}

func ParsePaymentStatus(status string) PaymentStatus {
	switch status {
	case "NOTPAY":
		return PaymentStatusNotPay
	case "SUCCESS":
		return PaymentStatusSuccess
	case "CLOSED":
		return PaymentStatusClosed
	case "REFUND":
		return PaymentStatusRefund
	default:
		return ""
	}
}

func (p PaymentType) String() string {
	return string(p)
}

func GetAllPaymentType() []PaymentType {
	return []PaymentType{
		PaymentTypeOneTime,
		PaymentTypeSubscription,
	}
}

func ParsePaymentType(t string) PaymentType {
	switch t {
	case "one_time":
		return PaymentTypeOneTime
	case "subscription":
		return PaymentTypeSubscription
	default:
		return PaymentTypeOneTime
	}
}

func (s SubscriptionStatus) String() string {
	return string(s)
}

func GetAllSubscriptionStatus() []SubscriptionStatus {
	return []SubscriptionStatus{
		SubscriptionStatusActive,
		SubscriptionStatusPastDue,
		SubscriptionStatusCanceled,
		SubscriptionStatusUnpaid,
	}
}

func ParseSubscriptionStatus(status string) SubscriptionStatus {
	switch status {
	case "active":
		return SubscriptionStatusActive
	case "past_due":
		return SubscriptionStatusPastDue
	case "canceled":
		return SubscriptionStatusCanceled
	case "unpaid":
		return SubscriptionStatusUnpaid
	default:
		return ""
	}
}

func (i SubscriptionInterval) String() string {
	return string(i)
}

func GetAllSubscriptionInterval() []SubscriptionInterval {
	return []SubscriptionInterval{
		SubscriptionIntervalMonthly,
		SubscriptionIntervalYearly,
	}
}

func ParseSubscriptionInterval(interval string) SubscriptionInterval {
	switch interval {
	case "monthly":
		return SubscriptionIntervalMonthly
	case "yearly":
		return SubscriptionIntervalYearly
	default:
		return ""
	}
}
