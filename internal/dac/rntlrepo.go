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

type RentalPropertyRepo struct {
	Persister
}

func InitRentalPrptyRepo(db Persister) RentalPropertyRepo {
	return RentalPropertyRepo{
		Persister: db,
	}
}

func (p *RentalPropertyRepo) Fetch(ctx context.Context, fltr filters.Filter) (interface{}, error) {
	var prpty model.RentalProperty
	limit := utils.DeterminRecordLimit(fltr.Limit)
	err := p.GetBunDB().NewSelect().Model(&prpty).
		Where("? = ?", bun.Ident("pid"), fltr.Identifier).Limit(limit).Offset(10*(fltr.Page-1)).Scan(ctx, &prpty)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoResults{Shape: "Rental Property", Identifier: fltr.Identifier, Err: err}
		}
		return nil, ErrFetchFailed{Model: "Rental Property", Err: err}
	}

	return prpty, nil
}

func (p *RentalPropertyRepo) FetchAll(ctx context.Context, fltr filters.Filter) ([]model.RentalProperty, error) {
	var propertys []model.RentalProperty
	limit := utils.DeterminRecordLimit(fltr.Limit)
	err := p.GetBunDB().NewSelect().Model(&propertys).Limit(limit).Offset(10*(fltr.Page-1)).Scan(ctx, &propertys)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoResults{Shape: model.RentalProperty{}, Identifier: "[fetch-all]", Err: err}
		}
		return nil, ErrFetchFailed{Model: "Rental Property", Err: err}
	}

	return propertys, nil
}

func (p *RentalPropertyRepo) Insert(ctx context.Context, prpty any) (interface{}, error) {
	property, ok := prpty.(*model.RentalProperty)

	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.RentalProperty{}, Got: prpty}
	}

	tx, err := p.GetBunDB().BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	err = tx.NewInsert().Model(property).Scan(ctx, property)

	if err != nil {
		log.Printf("db err: %v", err)
		if err = tx.Rollback(); err != nil {
			return model.RentalProperty{}, err
		}
		return nil, ErrInsertFailed{Model: "Rental Property", Err: err}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("failed to commit tx %v", err)
		return model.RentalProperty{}, err
	}

	return property, nil
}

func (p *RentalPropertyRepo) Update(ctx context.Context, prpty any) (interface{}, error) {
	property, ok := prpty.(model.RentalProperty)

	if !ok {
		e := cmerr.ErrUnexpectedData{Wanted: model.RentalProperty{}, Got: prpty}
		log.Printf("failed type asert in repo %v", e)
		return model.RentalProperty{}, cmerr.ErrUnexpectedData{Wanted: model.RentalProperty{}, Got: prpty}
	}

	tx, err := p.GetBunDB().BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		log.Printf("transaction start fail %v", err)
		return model.RentalProperty{}, ErrTransactionStartFailed{Err: err}
	}

	err = tx.NewUpdate().Model(&property).OmitZero().Where("? = ?", bun.Ident("pid"), property.Pid).Returning("*").Scan(ctx, &property)

	if err != nil {
		log.Printf("failed to update %v", err)
		if err = tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		log.Printf("failed to commit transaction %v", err)
		return model.RentalProperty{}, err
	}

	return property, nil
}

func (p *RentalPropertyRepo) Delete(ctx context.Context, prpty any) error {
	property, ok := prpty.(model.RentalProperty)

	if !ok {
		return cmerr.ErrUnexpectedData{Wanted: model.RentalProperty{}, Got: prpty}
	}

	tx, err := p.GetBunDB().BeginTx(ctx, &sql.TxOptions{})

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

func (p *RentalPropertyRepo) GetExisting(ctx context.Context, pid string) (model.RentalProperty, error) {
	var property model.RentalProperty
	err := p.GetBunDB().NewSelect().Model(&property).Column("pid").Where("? = ?", bun.Ident("pid"), pid).Scan(ctx, &property)

	if err != nil {
		return model.RentalProperty{}, err
	}

	return property, nil
}
