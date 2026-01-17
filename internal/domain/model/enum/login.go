package enum

type LoginType string

const (
	LoginTypePhone LoginType = "phone"
	LoginTypeEmail LoginType = "email"
	LoginTypeWx    LoginType = "wx"
)

// IsValidLoginType 检查登录类型是否有效
func IsValidLoginType(loginType string) bool {
	switch LoginType(loginType) {
	case LoginTypePhone, LoginTypeEmail, LoginTypeWx:
		return true
	default:
		return false
	}
}

// GetValidLoginTypes 获取所有有效的登录类型
func GetValidLoginTypes() []string {
	return []string{
		string(LoginTypePhone),
		string(LoginTypeEmail),
		string(LoginTypeWx),
	}
}
