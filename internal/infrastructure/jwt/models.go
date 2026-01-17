package jwt

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/futurxlab/golanggraph/xerror"
)

const (
	RS256ALGRAS    = "RS256"
	ACCESS         = "access"
	REGISTERVERIFY = "register_verify"
	PASSWORDRESET  = "password_reset"
)

type JWTToken struct {
	Head    string
	Payload string
	Sign    string
}

func (jwt *JWTToken) String() string {
	return fmt.Sprintf("%s.%s.%s", jwt.Head, jwt.Payload, jwt.Sign)
}

func (jwt *JWTToken) B64String() string {
	plain := fmt.Sprintf("%s.%s.%s", jwt.Head, jwt.Payload, jwt.Sign)
	return b64.RawURLEncoding.EncodeToString([]byte(plain))
}

func (jwt *JWTToken) UnmarshalPayload(v interface{}) error {
	b, err := b64.RawURLEncoding.DecodeString(jwt.Payload)
	if err != nil {
		return xerror.Wrap(err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		return xerror.Wrap(err)
	}

	return nil
}

type Head struct {
	Type string `json:"type"`
	Alg  string `json:"alg"`
}

type Payload struct {
	Type   string `json:"type"`
	Create int64  `json:"iat"`
	Expire int64  `json:"exp"`
}

type AccessPayload struct {
	Payload
	UserID         string   `json:"sub"`
	Application    string   `json:"iss"`
	PersonalRole   string   `json:"roles"`
	Scopes         []string `json:"scopes"`
	DeviceType     string   `json:"device_type"`
	DeviceID       string   `json:"device_id"`
	OrganizationID string   `json:"organization_id"`
}
