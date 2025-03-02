package dac

import (
	"context"
	"database/sql"

	"github.com/Z3DRP/lessor-service/internal/cmerr"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/Z3DRP/lessor-service/pkg/utils"
	"github.com/uptrace/bun"
)

type TaskRepo struct {
	Persister
	Limit int
}

func InitTskRepo(db Persister) *TaskRepo {
	return &TaskRepo{
		Persister: db,
	}
}

func (t *TaskRepo) Fetch(ctx context.Context, fltr filters.Filter) (interface{}, error) {
	var tsk model.Task
	limit := utils.DeterminRecordLimit(fltr.Limit)
	err := t.GetBunDB().NewSelect().Model(&tsk).
		Where("? = ?", bun.Ident("tid"), fltr.Identifier).Limit(limit).Offset(10 * (fltr.Page - 1)).Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoResults{Shape: tsk, Identifier: fltr.Identifier, Err: err}
		}

		return nil, ErrFetchFailed{Model: "Task", Err: err}
	}

	return tsk, nil
}

func (t TaskRepo) FetchAll(ctx context.Context, fltr filters.Filter) ([]model.Task, error) {
	var tsks []model.Task
	limit := utils.DeterminRecordLimit(fltr.Limit)
	err := t.GetBunDB().NewSelect().Model(&tsks).Limit(limit).Offset(10 * (fltr.Page - 1)).Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoResults{Shape: model.Task{}, Identifier: "[fetch-all]", Err: err}
		}
		return nil, ErrFetchFailed{Model: "Task", Err: err}
	}

	return tsks, nil
}

func (t *TaskRepo) Insert(ctx context.Context, tsk any) (interface{}, error) {
	tk, ok := tsk.(model.Task)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
	}

	tx, err := t.GetBunDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	rslt, err := tx.NewInsert().Model(&tk).Returning("*").Exec(ctx)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, ErrRollbackFailed{err}
		}
		return nil, ErrInsertFailed{Model: "Task", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return nil, ErrTransactionCommitFail{err}
	}
	return rslt, err
}

func (t *TaskRepo) Update(ctx context.Context, tsk any) (interface{}, error) {
	tk, ok := tsk.(model.Task)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
	}

	tx, err := t.GetBunDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	rslt, err := tx.NewUpdate().Model(&tk).WherePK().Returning("*").Exec(ctx)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, ErrRollbackFailed{err}
		}
		return nil, ErrUpdateFailed{Model: "Task", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return nil, ErrTransactionCommitFail{err}
	}
	return rslt, nil
}

func (t *TaskRepo) Delete(ctx context.Context, tsk any) error {
	tk, ok := tsk.(model.Task)
	if !ok {
		return cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
	}

	tx, err := t.GetBunDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return ErrTransactionStartFailed{Err: err}
	}

	_, err = tx.NewDelete().Model(&tk).WherePK().Exec(ctx)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return ErrRollbackFailed{err}
		}
		return err
	}

	if err = tx.Commit(); err != nil {
		return ErrTransactionCommitFail{err}
	}
	return nil
}
