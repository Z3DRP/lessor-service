package filters

import "fmt"

type ErrFailedToMakeFilter struct {
	FilterType string
}

func NewFailedToMakeFilterErr(fltrType string) ErrFailedToMakeFilter {
	return ErrFailedToMakeFilter{FilterType: fltrType}
}

func (e ErrFailedToMakeFilter) Error() string {
	return fmt.Sprintf("failed to create %v filter type", e.FilterType)
}

type ErrInvalidUuidFormat struct {
	Err error
}

func (e ErrInvalidUuidFormat) Error() string {
	return fmt.Sprintf("invalid uuid format, %v", e.Err)
}
