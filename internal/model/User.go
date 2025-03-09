package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	Id          int64                  `bun:"column:id,pk,autoincrement" json:"-"`
	Uid         uuid.UUID              `bun:"type:uuid,notnull,unique" json:"uid"`
	FirstName   string                 `bun:"type:varchar(250),notnull" json:"firstName"`
	LastName    string                 `bun:"type:varchar(250),notnull" json:"lastName"`
	Address     map[string]interface{} `bun:"type:json,json_use_number" json:"address"`
	Email       string                 `bun:"type:varchar(150),unique,notnull" json:"email"`
	Phone       string                 `bun:"type:varchar(12),notnull" json:"phone"`
	ProfileType string                 `bun:"type:varchar(100),notnull" json:"profileType"`
	Username    string                 `bun:"type:varchar(30),unique,notnull" json:"username"`
	Password    string                 `bun:"type:varchar(128),nullzero" json:"-"`
	IsActive    bool                   `bun:"column:is_active,notnull" json:"isActive"`
	AvatarFile  string                 `bun:"type:varchar(100),nullzero" json:"avatarFile"`
	CreatedAt   time.Time              `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp" json:"-"`
	UpdatedAt   time.Time              `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp" json:"-"`
}

var _ bun.BeforeAppendModelHook = (*User)(nil)

func (u *User) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.UpdateQuery:
		u.UpdatedAt = time.Now()
	}
	return nil
}

func (u User) Info() string {
	return fmt.Sprintf("%#v\n", u)
}
