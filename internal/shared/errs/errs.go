package errs

import "github.com/cristiano-pacheco/pingo/pkg/errs"

var (
	ErrRecordNotFound = errs.New("RECORD_NOT_FOUND", "Record not found", 404, nil)
)
