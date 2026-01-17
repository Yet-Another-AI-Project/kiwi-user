package enum

type OrganizationStatus string

const (
	OrganizationStatusNormal   OrganizationStatus = "normal"
	OrganizationStatusAdvanced OrganizationStatus = "advanced"
	OrganizationStatusPremium  OrganizationStatus = "premium"
	OrganizationStatusTrial    OrganizationStatus = "trial"
	OrganizationStatusUnknown  OrganizationStatus = "unknown"
)

func (o OrganizationStatus) String() string {
	return string(o)
}

func GetAllOrganizationStatus() []OrganizationStatus {
	return []OrganizationStatus{
		OrganizationStatusNormal,
		OrganizationStatusAdvanced,
		OrganizationStatusPremium,
		OrganizationStatusTrial,
		OrganizationStatusUnknown,
	}
}

func ParseOrganizationStatus(organizationStatus string) OrganizationStatus {
	switch organizationStatus {
	case OrganizationStatusNormal.String():
		return OrganizationStatusNormal
	case OrganizationStatusAdvanced.String():
		return OrganizationStatusAdvanced
	case OrganizationStatusPremium.String():
		return OrganizationStatusPremium
	case OrganizationStatusTrial.String():
		return OrganizationStatusTrial
	default:
		return OrganizationStatusUnknown
	}
}

type OrganizationRequestType string

const (
	OrganizationRequestTypeCreation OrganizationRequestType = "creation"
	OrganizationRequestTypeJoin     OrganizationRequestType = "join"
	OrganizationRequestTypeUnknown  OrganizationRequestType = "unknown"
)

func (o OrganizationRequestType) String() string {
	return string(o)
}

func GetAllOrganizationRequestTypes() []OrganizationRequestType {
	return []OrganizationRequestType{
		OrganizationRequestTypeCreation,
		OrganizationRequestTypeJoin,
		OrganizationRequestTypeUnknown,
	}
}

func ParseOrganizationRequestType(organizationRequestType string) OrganizationRequestType {
	switch organizationRequestType {
	case "creation":
		return OrganizationRequestTypeCreation
	case "join":
		return OrganizationRequestTypeJoin
	default:
		return OrganizationRequestTypeUnknown
	}
}

type OrganizationRequestStatus string

const (
	OrganizationRequestStatusPending  OrganizationRequestStatus = "pending"
	OrganizationRequestStatusApproved OrganizationRequestStatus = "approved"
	OrganizationRequestStatusReject   OrganizationRequestStatus = "reject"
	OrganizationRequestStatusUnknown  OrganizationRequestStatus = "unknown"
)

func (o OrganizationRequestStatus) String() string {
	return string(o)
}

func GetAllOrganizationRequestStatus() []OrganizationRequestStatus {
	return []OrganizationRequestStatus{
		OrganizationRequestStatusPending,
		OrganizationRequestStatusApproved,
		OrganizationRequestStatusReject,
		OrganizationRequestStatusUnknown,
	}
}

func ParseOrganizationRequestStatus(organizationRequestStatus string) OrganizationRequestStatus {
	switch organizationRequestStatus {
	case "pending":
		return OrganizationRequestStatusPending
	case "approved":
		return OrganizationRequestStatusApproved
	case "reject":
		return OrganizationRequestStatusReject
	default:
		return OrganizationRequestStatusUnknown
	}
}

// 申请行业
type OrganizationIndustryType string

const (
	OrganizationIndustryTypeMedical                  OrganizationIndustryType = "medical"
	OrganizationIndustryTypeEducation                OrganizationIndustryType = "education"
	OrganizationIndustryTypeFinance                  OrganizationIndustryType = "finance"
	OrganizationIndustryTypeLegal                    OrganizationIndustryType = "legal"
	OrganizationIndustryTypeITSoftware               OrganizationIndustryType = "itSoftware"
	OrganizationIndustryTypeManufacturingSupplyChain OrganizationIndustryType = "manufacturing_supply_chain"
	OrganizationIndustryTypeConsumerGoodsRetail      OrganizationIndustryType = "consumer_goods_retail"
	OrganizationIndustryTypeRealEstateProperty       OrganizationIndustryType = "real_estate_property"
)

func (o OrganizationIndustryType) String() string {
	return string(o)
}

func GetAllOrganizationIndustryType() []OrganizationIndustryType {
	return []OrganizationIndustryType{
		OrganizationIndustryTypeMedical,
		OrganizationIndustryTypeEducation,
		OrganizationIndustryTypeFinance,
		OrganizationIndustryTypeLegal,
		OrganizationIndustryTypeITSoftware,
		OrganizationIndustryTypeManufacturingSupplyChain,
		OrganizationIndustryTypeConsumerGoodsRetail,
		OrganizationIndustryTypeRealEstateProperty,
	}
}
