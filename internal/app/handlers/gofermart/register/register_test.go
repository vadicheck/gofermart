package register

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-playground/validator/v10"
	"github.com/vadicheck/gofermart/internal/app/config"
	"github.com/vadicheck/gofermart/internal/app/log"
	"github.com/vadicheck/gofermart/internal/app/storage/postgres"
	"github.com/vadicheck/gofermart/internal/app/storage/ptest"
)

func TestNew(t *testing.T) {
	type request struct {
		Login    string `json:"login" validate:"required,alphanum,max=140"`
		Password string `json:"password" validate:"required,max=140"`
	}
	type response struct {
		Result string `json:"result"`
	}
	type responseError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	type want struct {
		contentType   string
		statusCode    int
		response      response
		responseError responseError
	}
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "register test #1",
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusOK,
				response: response{
					Result: "",
				},
				responseError: responseError{},
			},
			request: request{
				Login:    "login",
				Password: "passw0rd",
			},
		},
		{
			name: "duplicate login #2",
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusConflict,
				response: response{
					Result: "",
				},
				responseError: responseError{
					Code:    http.StatusConflict,
					Message: "Login already exists",
				},
			},
			request: request{
				Login:    "login",
				Password: "passw0rd",
			},
		},
	}

	ctx := context.Background()

	cfg, err := config.NewConfig()
	if err != nil {
		panic(fmt.Errorf("config read err %w", err))
	}

	logger, err := log.New(*cfg)
	if err != nil {
		panic(err)
	}

	storage, err := postgres.New(cfg, logger)
	if err != nil {
		panic(err)
	}

	testStorage, err := ptest.New(cfg, logger)
	if err != nil {
		panic(err)
	}

	err = testStorage.DeleteAllUsers(ctx, logger)
	if err != nil {
		panic(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.request)
			if err != nil {
				fmt.Println("Ошибка кодирования в JSON:", err)
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler := New(
				ctx,
				cfg.Jwt,
				logger,
				storage,
				*validator.New(),
			)

			handler(w, req)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			if tt.want.statusCode == http.StatusConflict {
				var resError responseError
				dec := json.NewDecoder(result.Body)
				err = dec.Decode(&resError)
				assert.NoError(t, err)
				assert.Equal(t, tt.want.responseError.Message, resError.Message)
				return
			}
		})
	}
}
