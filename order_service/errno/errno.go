package errno

import "errors"

var (
	ErrQueryFailed = errors.New("query db failed")

	ErrUnderstock = errors.New("understock")
)
