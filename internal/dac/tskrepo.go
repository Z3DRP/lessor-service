package dac

import (
	"context"
	"database/sql"
	"log"

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

func InitTskRepo(db Persister) TaskRepo {
	return TaskRepo{
		Persister: db,
	}
}

func (t *TaskRepo) Fetch(ctx context.Context, fltr filters.Filter) (interface{}, error) {
	var tsk model.Task
	limit := utils.DeterminRecordLimit(fltr.Limit)
	err := t.GetBunDB().NewSelect().Model(&tsk).
		Where("? = ?", bun.Ident("tid"), fltr.Identifier).Relation("Worker").Relation("Property").Limit(limit).Offset(10 * (fltr.Page - 1)).Scan(ctx)

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
	log.Printf("making db  call now")
	limit := utils.DeterminRecordLimit(fltr.Limit)

	log.Printf("limit is %v", limit)

	err := t.GetBunDB().NewSelect().Model(&tsks).Relation("Property").Relation("Alessor").Relation("Worker").Relation("Worker.User").Scan(ctx, &tsks)
	for _, tk := range tsks {
		log.Println("task worker data: ")
		log.Printf("worker: %v", tk.Worker)
	}
	log.Printf("db err after call %v", err)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("db no rows err %v", err)
			return nil, ErrNoResults{Shape: model.Task{}, Identifier: "[fetch-all]", Err: err}
		}
		log.Printf("db error %v", err)
		return nil, ErrFetchFailed{Model: "Task", Err: err}
	}

	log.Println("no db error")
	log.Printf("task returned %v", tsks)
	return tsks, nil
}

func (t *TaskRepo) Insert(ctx context.Context, tsk any) (interface{}, error) {
	tk, ok := tsk.(*model.Task)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
	}

	tx, err := t.GetBunDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Printf("failed to start transaction %v", err)
		return nil, ErrTransactionStartFailed{Err: err}
	}

	err = tx.NewInsert().Model(tk).Returning("*").Scan(ctx, tk)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return nil, ErrRollbackFailed{rbErr}
		}
		return nil, ErrInsertFailed{Model: "Task", Err: err}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("error committing %v", err)
		return nil, ErrTransactionCommitFail{err}
	}

	return tk, nil
}

func (t *TaskRepo) Update(ctx context.Context, tsk any) (interface{}, error) {
	tk, ok := tsk.(*model.Task)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
	}

	tx, err := t.GetBunDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	err = tx.NewUpdate().Model(tk).OmitZero().Where("? = ?", bun.Ident("tid"), tk.Tid).Returning("*").Scan(ctx, tk)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, ErrRollbackFailed{err}
		}
		return nil, ErrUpdateFailed{Model: "Task", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return nil, ErrTransactionCommitFail{err}
	}
	return tk, nil
}

func (t *TaskRepo) UpdatePriority(ctx context.Context, tsk interface{}) (interface{}, error) {
	tk, ok := tsk.(model.Task)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
	}

	tx, err := t.GetBunDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	err = tx.NewUpdate().Model(&tk).OmitZero().Where("? = ?", bun.Ident("tid"), tk.Tid).Set("priority = ?", tk.Priority).Returning("*").Scan(ctx, &tk)

	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, ErrRollbackFailed{err}
		}
		return nil, ErrUpdateFailed{Model: "Task", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return nil, ErrTransactionCommitFail{err}
	}

	return tk, nil
}

func (t *TaskRepo) UpdateStartedAt(ctx context.Context, tsk interface{}) (interface{}, error) {
	tk, ok := tsk.(model.Task)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
	}

	tx, err := t.GetBunDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	err = tx.NewUpdate().Model(&tk).OmitZero().Where("? = ?", bun.Ident("tid"), tk.Tid).Set("started_at = ?", tk.StartedAt).Returning("*").Scan(ctx, &tk)

	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, ErrRollbackFailed{err}
		}
		return nil, ErrUpdateFailed{Model: "Task", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return nil, ErrTransactionCommitFail{err}
	}

	return tk, nil
}

func (t *TaskRepo) UpdateCompletedAt(ctx context.Context, tsk interface{}) (interface{}, error) {
	tk, ok := tsk.(model.Task)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
	}

	tx, err := t.GetBunDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	err = tx.NewUpdate().Model(&tk).OmitZero().Where("? = ?", bun.Ident("tid"), tk.Tid).Set("completed_at = ?", tk.CompletedAt).Returning("*").Scan(ctx, &tk)

	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, ErrRollbackFailed{err}
		}
		return nil, ErrUpdateFailed{Model: "Task", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return nil, ErrTransactionCommitFail{err}
	}

	return tk, nil
}

func (t *TaskRepo) UpdatePausedAt(ctx context.Context, tsk interface{}) (interface{}, error) {
	tk, ok := tsk.(model.Task)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
	}

	tx, err := t.GetBunDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	err = tx.NewUpdate().Model(&tk).OmitZero().Where("? = ?", bun.Ident("tid"), tk.Tid).Set("paused_at = ?", tk.PausedAt).Returning("*").Scan(ctx, &tk)

	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, ErrRollbackFailed{err}
		}
		return nil, ErrUpdateFailed{Model: "Task", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return nil, ErrTransactionCommitFail{err}
	}

	return tk, nil
}

func (t *TaskRepo) BulkPriorityUpdate(ctx context.Context, tasks []interface{}) ([]model.Task, error) {
	var err error
	uTasks := make([]model.Task, 0)
	for _, tsk := range tasks {
		_, ok := tsk.(model.Task)
		if !ok {
			err = cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
			break
		}
	}

	if err != nil {
		return nil, err
	}

	tx, err := t.GetBunDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	values := t.GetBunDB().NewValues(&tasks)
	err = tx.NewUpdate().With("_data", values).
		Model((*model.Task)(nil)).TableExpr("_data").
		Set("task.priority = _data.priority").Where("task.tid = _data.tid").
		Returning("*").Scan(ctx, &uTasks)

	if err != nil {
		return nil, err
	}

	return uTasks, nil
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

	_, err = tx.NewDelete().Model(&tk).Where("? = ?", bun.Ident("tid"), tk.Tid).Exec(ctx)
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
