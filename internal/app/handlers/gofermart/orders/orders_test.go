package orders

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vadicheck/gofermart/internal/app/config"
	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/log"
	storage2 "github.com/vadicheck/gofermart/internal/app/storage"
	"github.com/vadicheck/gofermart/internal/app/storage/postgres"
	"github.com/vadicheck/gofermart/internal/app/storage/ptest"
)

func TestNew(t *testing.T) {
	type request struct {
		UserID int `json:"user_id"`
	}
	type userData struct {
		ID       int     `json:"id"`
		Login    string  `json:"login"`
		Password string  `json:"password"`
		Balance  float32 `json:"balance"`
	}
	type orderData struct {
		UserID  int     `json:"user_id"`
		OrderID string  `json:"order_id"`
		Accrual float32 `json:"accrual"`
		Status  string  `json:"status"`
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
		{ID: 1, Login: "orders1", Password: "passw0rd", Balance: 1000},
		{ID: 2, Login: "orders2", Password: "passw0rd", Balance: 1000},
		{ID: 3, Login: "orders3", Password: "passw0rd", Balance: 1000},
	}
	orders := []orderData{
		{UserID: 1, OrderID: "9617519385", Accrual: 100, Status: "NEW"},
		{UserID: 3, OrderID: "9864048302", Accrual: 100, Status: "NEW"},
		{UserID: 3, OrderID: "6713493507", Accrual: 100, Status: "NEW"},
		{UserID: 3, OrderID: "6713493507", Accrual: 100, Status: "NEW"},
	}
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "exists orders #1",
			want: want{
				contentType:   "application/json",
				statusCode:    http.StatusOK,
				responseError: responseError{},
			},
			request: request{
				UserID: 1,
			},
		},
		{
			name: "no orders #2",
			want: want{
				contentType:   "application/json",
				statusCode:    http.StatusNoContent,
				responseError: responseError{},
			},
			request: request{
				UserID: 2,
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

	logins := make([]string, len(users))

	for i, user := range users {
		logins[i] = user.Login
	}

	err = testStorage.DeleteUsers(ctx, logger, logins)
	if err != nil {
		panic(err)
	}

	for _, u := range users {
		err = testStorage.CreateUser(ctx, u.ID, u.Login, u.Password, u.Balance)
		if err != nil && !errors.Is(err, storage2.ErrLoginAlreadyExists) {
			panic(err)
		}
	}

	for _, o := range orders {
		err = testStorage.CreateOrder(ctx, o.UserID, o.OrderID, o.Accrual, o.Status)
		if err != nil && !errors.Is(err, storage2.ErrOrderAlreadyExists) {
			panic(err)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", http.NoBody)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := New(
				ctx,
				logger,
				storage,
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
