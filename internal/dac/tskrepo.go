package dac

import (
	"context"
	"database/sql"

	"github.com/Z3DRP/lessor-service/internal/cmerr"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/uptrace/bun"
)

type TaskRepo struct {
	Store
	Limit int
}

func InitTskRepo(db Store) *TaskRepo {
	return &TaskRepo{
		Store: db,
		Limit: DefaultRecordLimit,
	}
}

func (t *TaskRepo) Fetch(ctx context.Context, fltr filters.PrimaryKeyFilter) (interface{}, error) {
	var tsk model.Task
	err := t.BdB.NewSelect().Model(&tsk).
		Where("? = ?", bun.Ident("id"), fltr.PK).Limit(t.Limit).Offset(10 * (fltr.Page - 1)).Scan(ctx)

	if err != nil {
		return nil, ErrFetchFailed{Model: "Task", Err: err}
	}

	return tsk, nil
}

func (t TaskRepo) FetchAll(ctx context.Context, pg int) ([]model.Task, error) {
	var tsks []model.Task
	err := t.BdB.NewSelect().Model(&tsks).Limit(t.Limit).Offset(10 * (pg - 1)).Scan(ctx)

	if err != nil {
		return nil, ErrFetchFailed{Model: "Task", Err: err}
	}

	return tsks, nil
}

func (t *TaskRepo) Insert(ctx context.Context, tsk any) (interface{}, error) {
	tk, ok := tsk.(model.Task)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
	}

	tx, err := t.BdB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	rslt, err := tx.NewInsert().Model(&tk).Returning("*").Exec(ctx)
	if err != nil {
		tx.Rollback()
		return nil, ErrInsertFailed{Model: "Task", Err: err}
	}

	tx.Commit()
	return rslt, err
}

func (t *TaskRepo) Update(ctx context.Context, tsk any) (interface{}, error) {
	tk, ok := tsk.(model.Task)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
	}

	tx, err := t.BdB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	rslt, err := tx.NewUpdate().Model(&tk).WherePK().Returning("*").Exec(ctx)
	if err != nil {
		tx.Rollback()
		return nil, ErrUpdateFailed{Model: "Task", Err: err}
	}

	tx.Commit()
	return rslt, nil
}

func (t *TaskRepo) Delete(ctx context.Context, tsk any) error {
	tk, ok := tsk.(model.Task)
	if !ok {
		return cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
	}

	tx, err := t.BdB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return ErrTransactionStartFailed{Err: err}
	}

	_, err = tx.NewDelete().Model(&tk).WherePK().Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
