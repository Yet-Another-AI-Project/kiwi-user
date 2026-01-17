package enum

type PaymentChannel string

const (
	PaymentChannelWechat PaymentChannel = "wechat"
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
