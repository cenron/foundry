package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cenron/foundry/internal/shared"
)

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func RespondError(w http.ResponseWriter, err error) {
	var notFound *shared.NotFoundError
	var validation *shared.ValidationError
	var conflict *shared.ConflictError

	switch {
	case errors.As(err, &notFound):
		RespondJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case errors.As(err, &validation):
		RespondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.As(err, &conflict):
		RespondJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	default:
		RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}
