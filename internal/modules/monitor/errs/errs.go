package errs

import (
	"net/http"

	"github.com/cristiano-pacheco/pingo/pkg/errs"
)

var (
	ErrInvalidContactType      = errs.New("MONITOR_01", "Invalid contact type", http.StatusBadRequest, nil)
	ErrContactNameAlreadyInUse = errs.New("MONITOR_02", "Contact name already in use", http.StatusConflict, nil)
	ErrInvalidContactEmail     = errs.New("MONITOR_03", "Invalid email address for contact", http.StatusBadRequest, nil)
	ErrInvalidContactWebhook   = errs.New("MONITOR_04", "Invalid webhook URL for contact", http.StatusBadRequest, nil)
)
