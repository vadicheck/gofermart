package currentbalance

import (
	"context"
	"encoding/json"
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
	type response struct {
		Current   float32 `json:"current"`
		Withdrawn float32 `json:"withdrawn"`
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
	type transactionData struct {
		UserID  int     `json:"user_id"`
		OrderID string  `json:"order_id"`
		Sum     float32 `json:"sum"`
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
	users := []userData{
		{ID: 101, Login: "currentbalance1", Password: "passw0rd", Balance: 1000},
		{ID: 102, Login: "currentbalance2", Password: "passw0rd", Balance: 1000},
		{ID: 103, Login: "currentbalance3", Password: "passw0rd", Balance: 1000},
	}
	orders := []orderData{
		{UserID: 101, OrderID: "1135306643", Accrual: 100, Status: "NEW"},
		{UserID: 103, OrderID: "3398206148", Accrual: 100, Status: "NEW"},
		{UserID: 103, OrderID: "2959282498", Accrual: 100, Status: "NEW"},
		{UserID: 103, OrderID: "4105553319", Accrual: 100, Status: "NEW"},
	}
	transactions := []transactionData{
		{UserID: 101, OrderID: "1135306643", Sum: 100},
		{UserID: 103, OrderID: "3398206148", Sum: 100},
		{UserID: 103, OrderID: "2959282498", Sum: 100},
		{UserID: 103, OrderID: "4105553319", Sum: 100},
	}
	tests := []struct {
		name     string
		userID   int
		response response
		want     want
	}{
		{
			name:   "0 withdrawn #1",
			userID: 102,
			want: want{
				contentType:   "application/json",
				statusCode:    http.StatusOK,
				responseError: responseError{},
				response: response{
					Current:   1000,
					Withdrawn: 0,
				},
			},
		},
		{
			name:   "300 withdrawn #2",
			userID: 103,
			want: want{
				contentType:   "application/json",
				statusCode:    http.StatusOK,
				responseError: responseError{},
				response: response{
					Current:   1000,
					Withdrawn: 300,
				},
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

	for _, t := range transactions {
		err = testStorage.CreateTransaction(ctx, t.UserID, t.OrderID, t.Sum)
		if err != nil {
			panic(err)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/user/balance", http.NoBody)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := New(
				ctx,
				logger,
				storage,
			)

			req.Header.Set(string(constants.XUserID), strconv.Itoa(tt.userID))

			handler(w, req)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			if tt.want.statusCode == http.StatusOK {
				var resp response

				dec := json.NewDecoder(result.Body)
				err = dec.Decode(&resp)

				assert.NoError(t, err)
				assert.Equal(t, tt.want.response.Current, resp.Current)
				assert.Equal(t, tt.want.response.Withdrawn, resp.Withdrawn)
			}
		})
	}
}
