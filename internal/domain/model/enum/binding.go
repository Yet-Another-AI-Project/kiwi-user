package enum

type BindingType string

const (
	BindingUnknown      BindingType = "unknown"
	BindingTypeWechat   BindingType = "wechat"
	BindingTypeQyWechat BindingType = "qy_wechat"
	BindingTypeWXID     BindingType = "wxid"
	BindingTypePhone    BindingType = "phone"
	BindingTypePassword BindingType = "password"
	BindingTypeEmail    BindingType = "email"
	BindingTypeGoogle   BindingType = "google"
)

func (b BindingType) String() string {
	return string(b)
}

func GetAllBindingTypes() []BindingType {
	return []BindingType{
		BindingTypeWechat,
		BindingTypeQyWechat,
		BindingTypeWXID,
		BindingTypePhone,
		BindingTypePassword,
		BindingTypeEmail,
		BindingTypeGoogle,
		BindingUnknown,
	}
}

func ParseBindingType(bindingType string) BindingType {
	switch bindingType {
	case "wechat":
		return BindingTypeWechat
	case "qy_wechat":
		return BindingTypeQyWechat
	case "wxid":
		return BindingTypeWXID
	case "phone":
		return BindingTypePhone
	case "password":
		return BindingTypePassword
	case "email":
		return BindingTypeEmail
	case "google":
		return BindingTypeGoogle
	default:
		return BindingUnknown
	}
}
