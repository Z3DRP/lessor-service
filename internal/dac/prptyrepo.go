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

type PropertyRepo struct {
	Store
}

func InitPrptyRepo(db Store) PropertyRepo {
	return PropertyRepo{
		Store: db,
	}
}

func (p *PropertyRepo) Fetch(ctx context.Context, fltr filters.Filter) (interface{}, error) {
	var prpty model.Property
	limit := utils.DeterminRecordLimit(fltr.Limit)
	err := p.BdB.NewSelect().Model(&prpty).
		Where("? = ?", bun.Ident("pid"), fltr.Identifier).Limit(limit).Offset(10*(fltr.Page-1)).Scan(ctx, &prpty)

	if err != nil {
		return nil, ErrFetchFailed{Model: "Property", Err: err}
	}

	return prpty, nil
}

func (p *PropertyRepo) FetchAll(ctx context.Context, fltr filters.Filter) ([]model.Property, error) {
	var propertys []model.Property
	limit := utils.DeterminRecordLimit(fltr.Limit)
	err := p.BdB.NewSelect().Model(&propertys).Limit(limit).Offset(10*(fltr.Page-1)).Scan(ctx, &propertys)

	if err != nil {
		return nil, ErrFetchFailed{Model: "Property", Err: err}
	}

	return propertys, nil
}

func (p *PropertyRepo) Insert(ctx context.Context, prpty any) (interface{}, error) {
	property, ok := prpty.(*model.Property)

	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Property{}, Got: prpty}
	}

	tx, err := p.BdB.BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	err = tx.NewInsert().Model(property).Scan(ctx, property)

	if err != nil {
		log.Printf("db err: %v", err)
		if err = tx.Rollback(); err != nil {
			return model.Property{}, err
		}
		return nil, ErrInsertFailed{Model: "Property", Err: err}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("failed to commit tx %v", err)
		return model.Property{}, err
	}

	return property, nil
}

func (p *PropertyRepo) Update(ctx context.Context, prpty any) (interface{}, error) {
	proptery, ok := prpty.(model.Property)

	if !ok {
		return model.Property{}, cmerr.ErrUnexpectedData{Wanted: model.Property{}, Got: prpty}
	}

	tx, err := p.BdB.BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		return model.Property{}, ErrTransactionStartFailed{Err: err}
	}

	rslt, err := tx.NewUpdate().Model(&proptery).Where("? = ?", bun.Ident("pid"), proptery.Pid).Returning("*").Exec(ctx, &proptery)

	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	log.Printf("property returned from update %v", proptery)
	log.Printf("result from property update %v", rslt)

	return rslt, nil
}

func (p *PropertyRepo) Delete(ctx context.Context, prpty any) error {
	property, ok := prpty.(model.Property)

	if !ok {
		return cmerr.ErrUnexpectedData{Wanted: model.Property{}, Got: prpty}
	}

	tx, err := p.BdB.BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		return ErrTransactionStartFailed{Err: err}
	}

	_, err = tx.NewDelete().Model(&property).Where("? = ?", bun.Ident("pid"), property.Pid).Exec(ctx)

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
