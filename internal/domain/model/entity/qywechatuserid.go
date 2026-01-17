package entity

import (
	"github.com/google/uuid"
)

type QyWechatUserIDEntity struct {
	ID             uuid.UUID
	QyWechatUserID string
	OpenID         string
}
