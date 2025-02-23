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

type AlessorRepo struct {
	Store
	Limit int
}

func InitAlsrRepo(db Store) AlessorRepo {
	return AlessorRepo{
		Store: db,
	}
}

func (a *AlessorRepo) Fetch(ctx context.Context, fltr filters.Filter) (interface{}, error) {
	var alsr model.Alessor
	limit := utils.DeterminRecordLimit(fltr.Limit)
	err := a.BdB.NewSelect().Model(&alsr).
		Where("? = ?", bun.Ident("uid"), fltr.Identifier).Limit(limit).Offset(10 * (fltr.Page - 1)).Scan(ctx)

	if err != nil {
		return nil, ErrFetchFailed{Model: "Alessor", Err: err}
	}

	return alsr, nil
}

func (a *AlessorRepo) FetchAll(ctx context.Context, fltr filters.Filter) ([]model.Alessor, error) {
	var alsrs []model.Alessor
	limit := utils.DeterminRecordLimit(fltr.Limit)
	err := a.BdB.NewSelect().Model(&alsrs).Limit(limit).Offset(10 * (fltr.Page - 1)).Scan(ctx)
	if err != nil {
		return nil, ErrFetchFailed{Model: "Alessor", Err: err}
	}

	return alsrs, nil
}

func (a *AlessorRepo) Insert(ctx context.Context, alsr any) (interface{}, error) {
	al, ok := alsr.(model.Alessor)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Alessor{}, Got: alsr}
	}

	tx, err := a.BdB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	rslt, err := tx.NewInsert().Model(&al).Returning("*").Exec(ctx)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, ErrRollbackFailed{Err: err}
		}
		return nil, ErrInsertFailed{Model: "Alessor", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return model.Alessor{}, err
	}
	return rslt, nil
}

func (a *AlessorRepo) Update(ctx context.Context, alsr any) (interface{}, error) {
	al, ok := alsr.(model.Alessor)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Alessor{}, Got: alsr}
	}

	tx, err := a.BdB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	rslt, err := tx.NewUpdate().Model(&al).Where("? = ?", bun.Ident("uid"), al.Uid).Returning("*").Exec(ctx)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, ErrRollbackFailed{Err: err}
		}
		return nil, ErrUpdateFailed{Model: "Alessor", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return rslt, err
}

func (a *AlessorRepo) Delete(ctx context.Context, alsr any) error {
	al, ok := alsr.(model.Alessor)
	if !ok {
		return cmerr.ErrUnexpectedData{Wanted: model.Alessor{}, Got: alsr}
	}

	tx, err := a.BdB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return ErrTransactionStartFailed{Err: err}
	}

	_, err = tx.NewDelete().Model(&al).Where("? = ?", bun.Ident("uid"), al.Uid).Exec(ctx)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return ErrRollbackFailed{Err: err}
		}
		return ErrDeleteFailed{Model: "Alessor", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
