package dac

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/Z3DRP/lessor-service/internal/cmerr"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/uptrace/bun"
)

type UserRepo struct {
	Store
	Limit int
}

func InitUsrRepo(db Store) UserRepo {
	return UserRepo{
		Store: db,
		Limit: DefaultRecordLimit,
	}
}

func (u *UserRepo) Fetch(ctx context.Context, fltr filters.UuidFilter) (interface{}, error) {
	var usr model.User
	err := u.BdB.NewSelect().Model(&usr).
		Where("? = ?", bun.Ident("uid"), fltr.Identifier).Limit(u.Limit).Offset(10*(fltr.Page-1)).Scan(ctx, &usr)
	if err != nil {
		return nil, ErrFetchFailed{Model: "User", Err: err}
	}

	return usr, nil
}

func (u *UserRepo) GetCredentials(ctx context.Context, email string) (interface{}, error) {
	var usr model.User
	log.Printf("email used: %v", email)
	err := u.BdB.NewSelect().Model(&usr).
		Column("uid", "username", "email", "profile_type", "password").
		Where("? = ?", bun.Ident("email"), email).Scan(ctx, &usr)

	log.Printf("user found: %v", usr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("no rows found")
			return nil, nil
		}
		return nil, err
	}

	return usr, nil
}

func (u *UserRepo) FetchAll(ctx context.Context, fltr filters.Filter) ([]model.User, error) {
	// need to add a company or lessor identifier
	var usrs []model.User
	err := u.BdB.NewSelect().Model(&usrs).Limit(u.Limit).Offset(10*(fltr.Page-1)).Scan(ctx, usrs)

	if err != nil {
		return nil, ErrFetchFailed{Model: "User", Err: err}
	}

	return usrs, nil
}

func (u *UserRepo) Insert(ctx context.Context, usr any) (interface{}, error) {
	user, ok := usr.(*model.User)

	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.User{}, Got: usr}
	}

	tx, err := u.BdB.BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	// this below gives rowsEffected not the new user
	//rslt, err := tx.NewInsert().Model(user).Returning("*").Exec(ctx)
	err = tx.NewInsert().Model(user).
		Returning("uid, username, first_name, last_name, profile_type").Scan(ctx, user)

	if err != nil {
		e := err
		log.Printf("db err: %v", err)
		if err = tx.Rollback(); err != nil {
			log.Printf("db rollback err: %v", err)
			return model.User{}, err
		}
		return nil, ErrInsertFailed{Model: "User", Err: e}
	}

	if err = tx.Commit(); err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (u *UserRepo) Update(ctx context.Context, usr any) (interface{}, error) {
	pf, ok := usr.(model.User)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.User{}, Got: usr}
	}

	tx, err := u.BdB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, ErrTransactionStartFailed{Err: err}
	}

	rslt, err := tx.NewUpdate().Model(&pf).Where("? = ?", bun.Ident("uid"), pf.Uid).Returning("*").Exec(ctx, &pf)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, ErrUpdateFailed{Model: "User", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return model.User{}, err
	}

	log.Printf("user result from update %v", pf)
	log.Printf("result from usr update %v", rslt)

	return rslt, nil
}

func (u *UserRepo) Delete(ctx context.Context, usr any) error {
	pf, ok := usr.(model.User)
	if !ok {
		return cmerr.ErrUnexpectedData{Wanted: model.User{}, Got: usr}
	}

	tx, err := u.BdB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return ErrTransactionStartFailed{Err: err}
	}

	_, err = tx.NewDelete().Model(&pf).Where("? = ?", bun.Ident("uid"), pf.Uid).Exec(ctx)
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
