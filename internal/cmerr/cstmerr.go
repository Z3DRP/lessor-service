package cmerr

import "fmt"

type ErrUnexpectedData struct {
	Wanted any
	Got    any
}

func (e ErrUnexpectedData) Error() string {
	return fmt.Sprintf("unexpected data wanted: %T, but got: %T", e.Wanted, e.Got)
}
