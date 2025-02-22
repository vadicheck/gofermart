package uporder

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vadicheck/gofermart/internal/app/config"
	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/log"
	"github.com/vadicheck/gofermart/internal/app/services/gofermart/order"
	"github.com/vadicheck/gofermart/internal/app/storage/postgres"
	"github.com/vadicheck/gofermart/internal/app/storage/ptest"
)

func TestNew(t *testing.T) {
	type request struct {
		UserID  int    `json:"user_id"`
		Content string `json:"content"`
	}
	type userData struct {
		ID       int    `json:"id"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	type responseError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	type want struct {
		contentType   string
		statusCode    int
		responseError responseError
	}
	users := []userData{
		{ID: 1, Login: "user1", Password: "passw0rd"},
		{ID: 2, Login: "user2", Password: "passw0rd"},
		{ID: 3, Login: "user3", Password: "passw0rd"},
	}
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "error number (luhn) #1",
			want: want{
				contentType:   "application/json",
				statusCode:    http.StatusUnprocessableEntity,
				responseError: responseError{},
			},
			request: request{
				UserID:  1,
				Content: "123457",
			},
		},
		{
			name: "uporder test #2",
			want: want{
				contentType:   "application/json",
				statusCode:    http.StatusAccepted,
				responseError: responseError{},
			},
			request: request{
				UserID:  1,
				Content: "123456789007",
			},
		},
		{
			name: "duplicate order number #3",
			want: want{
				contentType:   "application/json",
				statusCode:    http.StatusOK,
				responseError: responseError{},
			},
			request: request{
				UserID:  1,
				Content: "123456789007",
			},
		},
		{
			name: "conflict order number #4",
			want: want{
				contentType:   "application/json",
				statusCode:    http.StatusConflict,
				responseError: responseError{},
			},
			request: request{
				UserID:  2,
				Content: "123456789007",
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

	for _, user := range users {
		err = testStorage.CreateUser(ctx, user.ID, user.Login, user.Password, logger)
		if err != nil {
			panic(err)
		}
	}

	orderService := order.New(storage)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.request.Content))
			req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()

			handler := New(
				ctx,
				logger,
				storage,
				orderService,
			)

			req.Header.Set(string(constants.XUserID), strconv.Itoa(tt.request.UserID))

			handler(w, req)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}
