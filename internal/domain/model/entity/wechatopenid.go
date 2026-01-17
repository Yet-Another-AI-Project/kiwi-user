package entity

import (
	"kiwi-user/internal/domain/model/enum"

	"github.com/google/uuid"
)

type WechatOpenIDEntity struct {
	ID       uuid.UUID
	OpenID   string
	Platform enum.WechatOpenIDPlatform
}
