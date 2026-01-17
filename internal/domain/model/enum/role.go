package enum

type RoleType string

const (
	RoleTypeUnknown      RoleType = "unknown"
	RoleTypePersonal     RoleType = "personal"
	RoleTypeOrganization RoleType = "organization"
)

func (b RoleType) String() string {
	return string(b)
}

func GetAllRoleTypes() []RoleType {
	return []RoleType{
		RoleTypeUnknown,
		RoleTypePersonal,
		RoleTypeOrganization,
	}
}

func ParseRoleType(roleType string) RoleType {
	switch roleType {
	case "personal":
		return RoleTypePersonal
	case "organization":
		return RoleTypeOrganization
	default:
		return RoleTypeUnknown
	}
}

type DefaultRoleType string

const (
	DefaultRoleTypeUnknown           DefaultRoleType = "unknown"
	DefaultRoleTypePersonal          DefaultRoleType = "personal"
	DefaultRoleTypeOrganization      DefaultRoleType = "organization"
	DefaultRoleTypeOrganizationAdmin DefaultRoleType = "organization_admin"
)

func (b DefaultRoleType) String() string {
	return string(b)
}

func GetAllDefaultRoleTypes() []DefaultRoleType {
	return []DefaultRoleType{
		DefaultRoleTypeUnknown,
		DefaultRoleTypePersonal,
		DefaultRoleTypeOrganization,
		DefaultRoleTypeOrganizationAdmin,
	}
}

func ParseDefaultRoleType(roleType string) DefaultRoleType {
	switch roleType {
	case "personal":
		return DefaultRoleTypePersonal
	case "organization":
		return DefaultRoleTypeOrganization
	case "organization_admin":
		return DefaultRoleTypeOrganizationAdmin
	default:
		return DefaultRoleTypeUnknown
	}
}
