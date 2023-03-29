package model

import "time"

type UserMetaKey string

const (
	UMKAge    UserMetaKey = "age"
	UMKGender UserMetaKey = "gender"
)

var GendersMap = map[string]struct{}{
	"male":   {},
	"female": {},
	"none":   {},
}

var KeysMap = map[UserMetaKey]struct{}{
	UMKAge:    {},
	UMKGender: {},
}

type UserMeta struct {
	ID        uint        `gorm:"Column:id"`
	MetaKey   UserMetaKey `gorm:"Column:meta_key"`
	MetaValue string      `gorm:"Column:meta_value"`
	UserID    uint        `gorm:"Column:user_id"`
	UpdatedAt time.Time   `gorm:"Column:updated_at"`
	CreatedAt time.Time   `gorm:"Column:created_at"`
}
