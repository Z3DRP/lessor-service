package dac

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/Z3DRP/lessor-service/internal/cmerr"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/Z3DRP/lessor-service/pkg/utils"
	"github.com/uptrace/bun"
)

type NotificationRepo struct {
	Persister
	Limit int
}

func InitNotificationRepo(db Persister) NotificationRepo {
	return NotificationRepo{
		Persister: db,
	}
}

func (n *NotificationRepo) Fetch(ctx context.Context, fltr filters.Filter) (interface{}, error) {
	var noti model.Notification
	limit := utils.DeterminRecordLimit(fltr.Limit)
	err := n.GetBunDB().NewSelect().Model(&noti).
		Where("? = ?", bun.Ident("id"), fltr.Identifier).Relation("User").
		Relation("Property").Relation("Task").Limit(limit).Offset(10 * (fltr.Page - 1)).Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoResults{Shape: noti, Identifier: fltr.Identifier, Err: err}
		}
		return nil, ErrFetchFailed{Model: "Notification", Err: err}
	}

	return noti, nil
}

func (n *NotificationRepo) FetchAll(ctx context.Context, fltr filters.Filter) ([]model.Notification, error) {
	var notifs []model.Notification
	limit := utils.DeterminRecordLimit(fltr.Limit)

	err := n.GetBunDB().NewSelect().Model(&notifs).
		Where("? = ?", bun.Ident("notif.user_id"), fltr.Identifier).Where("void_at > ?", time.Now()).Where("not viewed").Relation("User").
		Relation("Property").Relation("Task").Limit(limit).Offset(10*(fltr.Page-1)).Scan(ctx, &notifs)

	// query := n.GetBunDB().NewSelect().Model(&notifs).
	// 	Where("? = ?", bun.Ident("notif.user_id"), fltr.Identifier).Where("void_at > ?", time.Now()).Where("not viewed").Relation("User").
	// 	Relation("Property").Relation("Task").Limit(limit).Offset(10*(fltr.Page-1)).Scan(ctx, &notifs)
	// fmt.Println(query.String())

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, ErrFetchFailed{Model: "Notification", Err: err}
	}

	return notifs, nil
}

func (n *NotificationRepo) Insert(ctx context.Context, notif any) (interface{}, error) {
	noti, ok := notif.(model.Notification)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Notification{}, Got: notif}
	}

	noti.VoidAt = time.Now().AddDate(0, 0, model.TtlDays)

	tx, err := n.GetBunDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Printf("failed to start transaction %v", err)
		return nil, ErrTransactionStartFailed{}
	}

	if err = tx.NewInsert().Model(&noti).Returning("*").Scan(ctx, &noti); err != nil {
		log.Printf("error inserting notification %v", err)
		if rbErr := tx.Rollback(); rbErr != nil {
			return nil, ErrRollbackFailed{rbErr}
		}
		return nil, ErrInsertFailed{Model: "Notification", Err: err}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("error commiting transactio %v", err)
		return nil, ErrTransactionCommitFail{err}
	}

	return noti, nil
}

func (n *NotificationRepo) UpdateViewed(ctx context.Context, notifId int) (interface{}, error) {
	var noti model.Notification
	tx, err := n.GetBunDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	if err = tx.NewUpdate().Model(&noti).OmitZero().Set("viewed = ?", true).Where("? = ?", bun.Ident("id"), notifId).Returning("*").Scan(ctx, &noti); err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, ErrRollbackFailed{err}
		}
		return nil, ErrUpdateFailed{Model: "Notification", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return nil, ErrTransactionCommitFail{err}
	}

	return noti, nil
}
