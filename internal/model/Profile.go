package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Profile struct {
	bun.BaseModel `bun:"table:profile,alias:prf"`

	Id         int64     `bun:"column:id,pk,autoincrement"`
	Uid        uuid.UUID `bun:"type:uuid,notnull,unique"`
	FirstName  string    `bun:"type:varchar(250),notnull"`
	LastName   string    `bun:"type:varchar(250),notnull"`
	Address    string    `bun:"type:json,json_use_number"`
	Email      string    `bun:"type:varchar(150),unique,notnull"`
	Phone      string    `bun:"type:varchar(12),notnull"`
	Username   string    `bun:"type:varchar(30),unique,notnull"`
	Password   string    `bun:"type:varchar(128),nullzero"`
	IsActive   bool      `bun:"type:boolean,notnull"`
	AvatarFile string    `bun:"type:varchar(100),nullzero"`
	CreatedAt  time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`

	UpdatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
}

var _ bun.BeforeAppendModelHook = (*Profile)(nil)

func (p *Profile) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.UpdateQuery:
		p.UpdatedAt = time.Now()
	}
	return nil
}

func (p Profile) Info() string {
	return fmt.Sprintf("%#v\n", p)
}
