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

	Id          int64                  `bun:"column:id,pk,autoincrement"`
	Uid         uuid.UUID              `bun:"type:uuid,notnull,unique"`
	FirstName   string                 `bun:"type:varchar(250),notnull"`
	LastName    string                 `bun:"type:varchar(250),notnull"`
	Address     map[string]interface{} `bun:"type:json,json_use_number"`
	Email       string                 `bun:"type:varchar(150),unique,notnull"`
	Phone       string                 `bun:"type:varchar(12),notnull"`
	ProfileType string                 `bun:"type:varchar(100),notnull"`
	Username    string                 `bun:"type:varchar(30),unique,notnull"`
	Password    string                 `bun:"type:varchar(128),nullzero"`
	IsActive    bool                   `bun:"column:is_active,notnull"`
	AvatarFile  string                 `bun:"type:varchar(100),nullzero"`
	CreatedAt   time.Time              `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
	UpdatedAt   time.Time              `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
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
