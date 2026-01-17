package enum

type WechatOpenIDPlatform string

const (
	WechatOpenIDPlatformMiniProgram WechatOpenIDPlatform = "miniprogram"
	WechatOpenIDPlatformUnknown     WechatOpenIDPlatform = "unknown"
)

func (w WechatOpenIDPlatform) String() string {
	return string(w)
}

func GetAllWechatOpenIDPlatform() []WechatOpenIDPlatform {
	return []WechatOpenIDPlatform{
		WechatOpenIDPlatformMiniProgram,
		WechatOpenIDPlatformUnknown,
	}
}

func ParseWechatOpenIDPlatform(platform string) WechatOpenIDPlatform {
	switch platform {
	case "miniprogram":
		return WechatOpenIDPlatformMiniProgram
	default:
		return WechatOpenIDPlatformUnknown
	}
}
