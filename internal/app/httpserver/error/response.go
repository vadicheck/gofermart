package error

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func RespondWithJSON(w http.ResponseWriter, status int, payload interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if encodeErr := json.NewEncoder(w).Encode(payload); encodeErr != nil {
		return fmt.Errorf("cannot encode response JSON body: %s", encodeErr)
	}

	return nil
}
