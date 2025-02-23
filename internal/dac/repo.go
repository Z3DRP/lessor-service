package dac

import (
	"context"
	"fmt"

	"github.com/Z3DRP/lessor-service/internal/dtos"
	"github.com/Z3DRP/lessor-service/internal/filters"
)

type Fetcher interface {
	Fetch(context.Context, filters.Filterer) (interface{}, error)
}

type Inserter interface {
	Insert(context.Context, any) error
}

type Updater interface {
	Update(context.Context, any) (interface{}, error)
}

type Deleter interface {
	Delete(context.Context, dtos.DeleteRequest) error
}

type Repoer interface {
	Fetcher
	Inserter
	Updater
	Deleter
}

type ErrFetchFailed struct {
	Model string
	Err   error
}

func (e ErrFetchFailed) Error() string {
	return fmt.Sprintf("failed to fetch %v: %v", e.Model, e.Err)
}

func (e ErrFetchFailed) Unwrap() error {
	return e.Err
}

type ErrInsertFailed struct {
	Model string
	Err   error
}

func (e ErrInsertFailed) Error() string {
	return fmt.Sprintf("failed to insert %v: %v", e.Model, e.Err)
}

func (e ErrInsertFailed) Unwrap() error {
	return e.Err
}

type ErrUpdateFailed struct {
	Model string
	Err   error
}

func (e ErrUpdateFailed) Error() string {
	return fmt.Sprintf("failed to update %v: %v", e.Model, e.Err)
}

func (e ErrUpdateFailed) Unwrap() error {
	return e.Err
}

type ErrDeleteFailed struct {
	Model string
	Err   error
}

func (e ErrDeleteFailed) Error() string {
	return fmt.Sprintf("failed to delete %v: %v", e.Model, e.Err)
}

func (e ErrDeleteFailed) Unwrap() error {
	return e.Err
}

type ErrTransactionStartFailed struct {
	Err error
}

func (e ErrTransactionStartFailed) Error() string {
	return fmt.Sprintf("could not start transaction: %v", e.Err)
}

func (e ErrTransactionStartFailed) Unwrap() error {
	return e.Err
}

type ErrRollbackFailed struct {
	Err error
}

func (e ErrRollbackFailed) Error() string {
	return fmt.Sprintf("rollback failed %v", e.Err)
}
