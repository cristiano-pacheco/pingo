package errs

import (
	"net/http"

	"github.com/cristiano-pacheco/pingo/pkg/errs"
)

var (
	ErrInvalidContactType = errs.New("MONITOR_01", "Invalid contact type", http.StatusBadRequest, nil)
)
