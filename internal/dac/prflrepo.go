package dac

import (
	"context"
	"database/sql"

	"github.com/Z3DRP/lessor-service/internal/cmerr"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/uptrace/bun"
)

type ProfileRepo struct {
	Persister
	Limit int
}

func InitPrflRepo(db Persister) ProfileRepo {
	return ProfileRepo{
		Persister: db,
	}
}

func (p *ProfileRepo) Fetch(ctx context.Context, fltr filters.UuidFilter) (interface{}, error) {
	var prfl model.Profile
	err := p.GetBunDB().NewSelect().Model(&prfl).
		Where("? = ?", bun.Ident("uid"), fltr.Identifier).Limit(p.Limit).Offset(10 * (fltr.Page - 1)).Scan(ctx)
	if err != nil {
		return nil, ErrFetchFailed{Model: "Profile", Err: err}
	}

	return prfl, nil
}

func (p *ProfileRepo) FetchAll(ctx context.Context, fltr filters.Filter) ([]model.Profile, error) {
	var prfls []model.Profile
	err := p.GetBunDB().NewSelect().Model(&prfls).Limit(p.Limit).Offset(10 * (fltr.Page - 1)).Scan(ctx)

	if err != nil {
		return nil, ErrFetchFailed{Model: "Profile", Err: err}
	}

	return prfls, nil
}

func (p *ProfileRepo) Insert(ctx context.Context, prfl any) (interface{}, error) {
	pf, ok := prfl.(model.Profile)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Profile{}, Got: prfl}
	}

	tx, err := p.GetBunDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	rslt, err := tx.NewInsert().Model(&pf).Returning("*").Exec(ctx)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return model.Profile{}, err
		}
		return nil, ErrInsertFailed{Model: "Profile", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return model.Profile{}, err
	}
	return rslt, nil
}

func (p *ProfileRepo) Update(ctx context.Context, prfl any) (interface{}, error) {
	pf, ok := prfl.(model.Profile)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Profile{}, Got: prfl}
	}

	tx, err := p.GetBunDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	rslt, err := tx.NewUpdate().Model(&pf).Where("? = ?", bun.Ident("uid"), pf.Uid).Returning("*").Exec(ctx)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return model.Profile{}, err
		}
		return nil, ErrUpdateFailed{Model: "Profile", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return model.Profile{}, err
	}

	return rslt, nil
}

func (p *ProfileRepo) Delete(ctx context.Context, prfl any) error {
	pf, ok := prfl.(model.Profile)
	if !ok {
		return cmerr.ErrUnexpectedData{Wanted: model.Profile{}, Got: prfl}
	}

	tx, err := p.GetBunDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return ErrTransactionStartFailed{Err: err}
	}

	_, err = tx.NewDelete().Model(&pf).Where("? = ?", bun.Ident("uid"), pf.Uid).Exec(ctx)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return err
		}
		return ErrDeleteFailed{Model: "Profile", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
