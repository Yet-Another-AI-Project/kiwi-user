package enum

type VertificationCodeType string

const (
	VertificationCodeTypeLogin   VertificationCodeType = "login"
	VertificationCodeTypeUnknown VertificationCodeType = "unknown"
)

func (v VertificationCodeType) String() string {
	return string(v)
}

func GetAllVertificationCodeTypes() []VertificationCodeType {
	return []VertificationCodeType{
		VertificationCodeTypeLogin,
		VertificationCodeTypeUnknown,
	}
}

func ParseVertificationCodeType(s string) VertificationCodeType {
	switch s {
	case "login":
		return VertificationCodeTypeLogin
	default:
		return VertificationCodeTypeUnknown
	}
}
