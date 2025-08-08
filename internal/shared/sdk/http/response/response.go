package response

import (
	"encoding/json"
	"errors"
	"log/slog"
	"maps"
	"net/http"

	"github.com/cristiano-pacheco/pingo/internal/shared/modules/errs"
)

func Error(w http.ResponseWriter, err error) {
	var rError *errs.Error
	if !errors.As(err, &rError) {
		// If it's not our custom error type, convert it to a generic error
		httpStatus := http.StatusInternalServerError
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		if encodeErr := json.NewEncoder(w).Encode(Envelope{
			"error": map[string]string{
				"code":    "internal_server_error",
				"message": "Internal server error",
			},
		}); encodeErr != nil {
			//nolint:sloglint // this is a response writer
			slog.Error("Failed to encode error response", "error", encodeErr)
		}
		return
	}

	if rError.Status == 0 {
		rError.Status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(rError.Status)
	if encodeErr := json.NewEncoder(w).Encode(rError); encodeErr != nil {
		//nolint:sloglint // this is a response writer
		slog.Error("Failed to encode error response", "error", encodeErr)
	}
}

func JSON(w http.ResponseWriter, status int, envelope Envelope, headers http.Header) {
	js, err := json.MarshalIndent(envelope, "", "\t")
	if err != nil {
		//nolint:sloglint // it's ok to have a global logger here
		slog.Error("Failed to marshal response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	js = append(js, '\n')

	maps.Copy(w.Header(), headers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, writeErr := w.Write(js); writeErr != nil {
		//nolint:sloglint // it's ok to have a global logger here
		slog.Error("Failed to write response", "error", writeErr)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
