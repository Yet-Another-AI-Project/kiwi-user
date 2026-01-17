package dto

type SendPhoneVerifyCodeRequest struct {
	Phone string `json:"phone"`
}

type VerifyPhoneCodeRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}
