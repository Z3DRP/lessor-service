package filters

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Z3DRP/lessor-service/pkg/utils"
	"github.com/google/uuid"
)

type Filterer interface {
	Validate() error
}

type Filter struct {
	Identifier string
	Page       int
	Limit      int
}

func NewFilter(idnfr string, pg, lmt int) Filter {
	return Filter{
		Identifier: idnfr,
		Page:       pg,
		Limit:      lmt,
	}
}

func GenFilter(r *http.Request) (Filter, error) {
	query := r.URL.Query()
	id := r.PathValue("id")

	if id == "" {
		return Filter{}, errors.New("failed to generate primary key filter, primary key not found in request")
	}

	page, err := strconv.Atoi(query.Get("page"))

	if err != nil {
		return Filter{}, fmt.Errorf("page was not included with request %v", err)
	}

	lmt, err := strconv.Atoi(query.Get("limit"))
	if err != nil {
		lmt = utils.DeterminRecordLimit(0)
	}

	fltr := Filter{Identifier: id, Page: page, Limit: lmt}
	err = fltr.Validate()

	if err != nil {
		return Filter{}, err
	}

	return fltr, nil

}

func (f Filter) Validate() error {
	if f.Page == 0 {
		f.Page = 1
	}

	if f.Page >= 1000 {
		return errors.New("invalid page, cannot paginate 1000 or more results")
	}

	if f.Limit <= 0 {
		return errors.New("invalid limit, must be positive number")
	}

	if f.Limit > 25 {
		return errors.New("invalid limit, must be less than 25")
	}

	return nil
}

type UuidFilter struct {
	Filter
}

func (u UuidFilter) Validate() error {
	if u.Page == 0 {
		u.Page = 1
	}

	if err := uuid.Validate(u.Identifier); err != nil {
		invalidFormatErr := ErrInvalidUuidFormat{Err: err}
		return invalidFormatErr
	}

	return nil
}

func GenUuidFilter(r *http.Request) (UuidFilter, error) {
	query := r.URL.Query()
	id := r.PathValue("identifier")

	if id == "" {
		return UuidFilter{}, errors.New("failed to generate primary key filter, primary key not found in request")
	}

	page, err := strconv.Atoi(query.Get("page"))

	if err != nil {
		return UuidFilter{}, err
	}

	fltr := Filter{Identifier: id, Page: page}
	ufltr := UuidFilter{Filter: fltr}
	err = ufltr.Validate()

	if err != nil {
		return UuidFilter{}, err
	}

	return ufltr, nil
}

type PropertyFilter struct {
}

type PrimaryKeyFilter struct {
	PK int64
	Filter
}

func (p PrimaryKeyFilter) Validate() error {
	if p.PK == 0 {
		return fmt.Errorf("invalid primary key %v for primary key filter", p.PK)
	}
	return nil
}

func NewPrimaryKeyFilter(pk int64, fltr Filter) *PrimaryKeyFilter {
	return &PrimaryKeyFilter{
		PK:     pk,
		Filter: fltr,
	}
}

func GenPkFilter(r *http.Request) (PrimaryKeyFilter, error) {
	query := r.URL.Query()
	pk := r.PathValue("id")

	if pk == "" {
		return PrimaryKeyFilter{}, errors.New("failed to generate primary key filter, primary key not found in request")
	}

	pkey, err := strconv.ParseInt(pk, 10, 64)

	if err != nil {
		return PrimaryKeyFilter{}, errors.New("could not parse primary key from value supplied by request")
	}

	page, err := strconv.Atoi(query.Get("page"))

	if err != nil {
		return PrimaryKeyFilter{}, err
	}

	fltr := Filter{Page: page}
	err = fltr.Validate()

	if err != nil {
		return PrimaryKeyFilter{}, err
	}

	return PrimaryKeyFilter{PK: pkey, Filter: fltr}, nil
}

func GenFilterWithNoSearch(r *http.Request) (Filter, error) {
	query := r.URL.Query()
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil {
		return Filter{}, err
	}

	return Filter{Page: page, Identifier: ""}, nil
}

type Creds struct {
	Email    string
	Password string
}

func (c Creds) Validate() error {
	return nil
}
