
package http

import (
	errors "errors"
)

var (
	ErrWrongFormat = errors.New("Wrong http header format")
	ErrHeaderHasWritten = errors.New("Http response header has been written")
)
