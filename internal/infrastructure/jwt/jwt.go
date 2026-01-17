package jwt

import (
	"crypto"
	"encoding/json"
	"fmt"
	"kiwi-user/config"
	"strings"
	"time"

	b64 "encoding/base64"

	"github.com/futurxlab/golanggraph/xerror"
)

var (
	ErrInvalidJWTToken = fmt.Errorf("invalid jwt token")
)

type JWTHelper struct {
	rsa                     *RSA
	accessTokenExpireSecond int64
}

func NewJWTHelper(config *config.Config, rsa *RSA) *JWTHelper {
	return &JWTHelper{
		rsa:                     rsa,
		accessTokenExpireSecond: config.JWT.AccessTokenExpireSecond,
	}
}

func (j *JWTHelper) GenerateRSA256JWT(payload any) (*JWTToken, error) {
	jwtToken := &JWTToken{}
	head := Head{}
	head.Alg = RS256ALGRAS
	head.Type = "jwt"
	h, err := json.Marshal(head)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	jwtToken.Head = b64.RawURLEncoding.EncodeToString(h)

	p, err := json.Marshal(payload)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	jwtToken.Payload = b64.RawURLEncoding.EncodeToString(p)

	sig, err := j.rsa.SignWithPrivateKey([]byte(fmt.Sprintf("%s.%s", jwtToken.Head, jwtToken.Payload)), crypto.SHA256)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	jwtToken.Sign = b64.RawURLEncoding.EncodeToString(sig)

	return jwtToken, nil
}

func (j *JWTHelper) DecodeAccessPayload(payload string) (*AccessPayload, error) {
	up := &AccessPayload{}
	b, err := b64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, xerror.Wrap(err)
	}
	if err := json.Unmarshal(b, up); err != nil {
		return nil, xerror.Wrap(err)
	}
	return up, nil
}

func (j *JWTHelper) VerifyRS256JWT(token string) (*JWTToken, error) {
	jwtToken := &JWTToken{}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, xerror.Wrap(ErrInvalidJWTToken)
	}

	jwtToken.Head = parts[0]
	jwtToken.Payload = parts[1]

	decodedSign, err := b64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, xerror.Wrap(ErrInvalidJWTToken)
	}

	jwtToken.Sign = parts[2]

	concat := fmt.Sprintf("%s.%s", parts[0], parts[1])

	if err := j.rsa.VerifySignWithPublicKey([]byte(concat), decodedSign, crypto.SHA256); err != nil {
		return nil, xerror.Wrap(ErrInvalidJWTToken)
	}
	return jwtToken, nil
}

func (j *JWTHelper) VerifyB64RS256JWT(b64JWTToken string) (*JWTToken, error) {
	b, err := b64.RawURLEncoding.DecodeString(b64JWTToken)
	if err != nil {
		return nil, xerror.Wrap(ErrInvalidJWTToken)
	}
	jwt, perr := j.VerifyRS256JWT(string(b))
	if perr != nil {
		return nil, perr
	}

	return jwt, nil
}

func (j *JWTHelper) NewAccessPayload(
	userID string,
	personalRole string,
	personalScopes []string,
	application string,
	deviceType string,
	deviceID string,
	organizationID string) *AccessPayload {
	up := &AccessPayload{}
	up.UserID = userID
	up.PersonalRole = personalRole
	up.Scopes = personalScopes
	up.Application = application
	up.DeviceType = deviceType
	up.DeviceID = deviceID
	up.OrganizationID = organizationID

	up.Payload.Type = ACCESS
	up.Payload.Create = time.Now().Unix()
	up.Payload.Expire = time.Now().Unix() + j.accessTokenExpireSecond
	return up
}
