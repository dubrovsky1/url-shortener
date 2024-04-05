package errs

import (
	"errors"
)

var ErrUniqueIndex = errors.New("unique index error")
var ErrShortURLNotFound = errors.New("not found short_url error")
