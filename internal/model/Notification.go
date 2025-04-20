package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type NotificationType string

const (
	PropertyAlert NotificationType = "property"
	TaskAlert     NotificationType = "task"
	UserAlert     NotificationType = "user"
	WorkerAlert   NotificationType = "worker"
	TenantAlert   NotificationType = "tenant"
	GeneralAlert  NotificationType = "general"
)

type Notification struct {
	bun.BaseModel `bun:"table:notifications,alias:notif"`
	Id            int              `bun:"column:id,ok,autoincrement" json:"-"`
	Title         string           `bun:"type:varchar(100),notnull," json:"title"`
	Message       string           `bun:"type:varchar(255),notnull," json:"message"`
	LessorId      uuid.UUID        `bun:"type:uuid,notnull,nullzero" json:"lessorId"`
	TaskId        uuid.UUID        `bun:"type:uuid,nullzero" json:"taskId"`
	Task          *Task            `bun:"rel:belongs-to,join:task_id=tid" json:"task"`
	UserId        uuid.UUID        `bun:"type:uuid,nullzero" json:"userId"`
	User          *User            `bun:"rel:belongs-to,join:user_id=uid" json:"user"`
	PropertyId    uuid.UUID        `bun:"type:uuid,nullzero" json:"propertyId"`
	Property      *Property        `bun:"rel:belongs-to,join:property_id=pid" json:"property"`
	Category      NotificationType `bun:"type:notification_type,notnull,nullzero,default:general" json:"category"`
	Viewed        bool             `bun:"type:boolean,notnull,nullzero,default:false" json:"viewed"`
	CreatedAt     time.Time        `bun:"type:timestamptz,notnull,nullzero" json:"createdAt"`
	VoidAt        time.Time        `bun:"type:timestamptz,notnull" json:"voidAt"`
}

func (n Notification) Str() string {
	return fmt.Sprintf("%+v", n)
}
