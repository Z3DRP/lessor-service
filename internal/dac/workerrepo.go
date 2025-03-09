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

type WorkerRepo struct {
	Persister
}

func InitWorkerRepo(db Persister) WorkerRepo {
	return WorkerRepo{
		Persister: db,
	}
}

func (p *WorkerRepo) Fetch(ctx context.Context, fltr filters.Filter) (interface{}, error) {
	var wrkr model.Worker
	limit := utils.DeterminRecordLimit(fltr.Limit)
	err := p.GetBunDB().NewSelect().Model(&wrkr).
		Where("? = ?", bun.Ident("pid"), fltr.Identifier).Limit(limit).Offset(10*(fltr.Page-1)).Scan(ctx, &wrkr)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoResults{Shape: "worker", Identifier: fltr.Identifier, Err: err}
		}
		return nil, ErrFetchFailed{Model: "worker", Err: err}
	}

	return wrkr, nil
}

func (p *WorkerRepo) FetchAll(ctx context.Context, fltr filters.Filter) ([]model.Worker, error) {
	var workers []model.Worker
	limit := utils.DeterminRecordLimit(fltr.Limit)
	err := p.GetBunDB().NewSelect().Model(&workers).Limit(limit).Offset(10*(fltr.Page-1)).Scan(ctx, &workers)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoResults{Shape: model.Worker{}, Identifier: "[fetch-all]", Err: err}
		}
		return nil, ErrFetchFailed{Model: "worker", Err: err}
	}

	return workers, nil
}

func (p *WorkerRepo) Insert(ctx context.Context, wrkr any) (interface{}, error) {
	worker, ok := wrkr.(*model.Worker)

	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Worker{}, Got: wrkr}
	}

	tx, err := p.GetBunDB().BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	err = tx.NewInsert().Model(worker).Scan(ctx, worker)

	if err != nil {
		log.Printf("db err: %v", err)
		if err = tx.Rollback(); err != nil {
			return model.Worker{}, err
		}
		return nil, ErrInsertFailed{Model: "worker", Err: err}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("failed to commit tx %v", err)
		return model.Worker{}, err
	}

	return worker, nil
}

func (p *WorkerRepo) Update(ctx context.Context, wrkr any) (interface{}, error) {
	worker, ok := wrkr.(model.Worker)

	if !ok {
		e := cmerr.ErrUnexpectedData{Wanted: model.Worker{}, Got: wrkr}
		log.Printf("failed type asert in repo %v", e)
		return model.Worker{}, cmerr.ErrUnexpectedData{Wanted: model.Worker{}, Got: wrkr}
	}

	tx, err := p.GetBunDB().BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		log.Printf("transaction start fail %v", err)
		return model.Worker{}, ErrTransactionStartFailed{Err: err}
	}

	err = tx.NewUpdate().Model(&worker).OmitZero().Where("? = ?", bun.Ident("uid"), worker.Uid).Returning("*").Scan(ctx, &worker)

	if err != nil {
		log.Printf("failed to update %v", err)
		if err = tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	log.Printf("worker returned from update %v", worker)
	log.Printf("result from worker update %v", worker)

	if err = tx.Commit(); err != nil {
		log.Printf("failed to commit transaction %v", err)
		return model.Worker{}, err
	}

	return worker, nil
}

func (p *WorkerRepo) Delete(ctx context.Context, wrkr any) error {
	worker, ok := wrkr.(model.Worker)
	log.Printf("deleting worker: %v", worker)

	if !ok {
		return cmerr.ErrUnexpectedData{Wanted: model.Worker{}, Got: wrkr}
	}

	tx, err := p.GetBunDB().BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		return ErrTransactionStartFailed{Err: err}
	}

	_, err = tx.NewDelete().Model(&worker).Where("? = ?", bun.Ident("Uid"), worker.Uid).Exec(ctx)

	if err != nil {
		if err = tx.Rollback(); err != nil {
			return err
		}
		return ErrDeleteFailed{Model: "User", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (p *WorkerRepo) GetExisting(ctx context.Context, pid string) (model.Worker, error) {
	var worker model.Worker
	err := p.GetBunDB().NewSelect().Model(&worker).Column("Uid", "image").Where("? = ?", bun.Ident("Uid"), pid).Scan(ctx, &worker)

	if err != nil {
		return model.Worker{}, err
	}

	return worker, nil
}
