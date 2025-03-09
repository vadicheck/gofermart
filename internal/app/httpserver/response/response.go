package response

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vadicheck/gofermart/internal/app/models/gofermart/response"
	"github.com/vadicheck/gofermart/pkg/logger"
)

func RespondWithJSON(w http.ResponseWriter, status int, payload interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if encodeErr := json.NewEncoder(w).Encode(payload); encodeErr != nil {
		return fmt.Errorf("cannot encode response JSON body: %s", encodeErr)
	}

	return nil
}

func Error(w http.ResponseWriter, err *response.Error, logger logger.LogClient) {
	if responseErr := RespondWithJSON(w, err.Code, err); responseErr != nil {
		logger.Error(fmt.Errorf("error responding with error: %w", responseErr))
	}
}
