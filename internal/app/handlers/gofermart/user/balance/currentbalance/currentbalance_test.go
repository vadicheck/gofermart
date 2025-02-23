package currentbalance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vadicheck/gofermart/internal/app/config"
	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/log"
	"github.com/vadicheck/gofermart/internal/app/storage/postgres"
	"github.com/vadicheck/gofermart/internal/app/storage/ptest"
)

func TestNew(t *testing.T) {
	type response struct {
		Current   float32 `json:"current"`
		Withdrawn float32 `json:"withdrawn"`
	}
	type userData struct {
		ID       int    `json:"id"`
		Login    string `json:"login"`
		Password string `json:"password"`
		Balance  int    `json:"balance"`
	}
	type orderData struct {
		UserID  int    `json:"user_id"`
		OrderID int    `json:"order_id"`
		Accrual int    `json:"accrual"`
		Status  string `json:"status"`
	}
	type transactionData struct {
		UserID  int `json:"user_id"`
		OrderID int `json:"order_id"`
		Sum     int `json:"sum"`
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
		{ID: 1, Login: "user1", Password: "passw0rd", Balance: 1000},
		{ID: 2, Login: "user2", Password: "passw0rd", Balance: 1000},
		{ID: 3, Login: "user3", Password: "passw0rd", Balance: 1000},
	}
	orders := []orderData{
		{UserID: 1, OrderID: 123456789007, Accrual: 100, Status: "NEW"},
		{UserID: 3, OrderID: 123456789015, Accrual: 100, Status: "NEW"},
		{UserID: 3, OrderID: 123456789023, Accrual: 100, Status: "NEW"},
		{UserID: 3, OrderID: 123456789031, Accrual: 100, Status: "NEW"},
	}
	transactions := []transactionData{
		{UserID: 1, OrderID: 123456789007, Sum: 100},
		{UserID: 3, OrderID: 123456789015, Sum: 100},
		{UserID: 3, OrderID: 123456789023, Sum: 100},
		{UserID: 3, OrderID: 123456789031, Sum: 100},
	}
	tests := []struct {
		name     string
		userID   int
		response response
		want     want
	}{
		{
			name:   "0 withdrawn #1",
			userID: 2,
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
			userID: 3,
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

	err = testStorage.DeleteAllUsers(ctx, logger)
	if err != nil {
		panic(err)
	}

	for _, u := range users {
		err = testStorage.CreateUser(ctx, u.ID, u.Login, u.Password, u.Balance, logger)
		if err != nil {
			panic(err)
		}
	}

	for _, o := range orders {
		err = testStorage.CreateOrder(ctx, o.UserID, o.OrderID, o.Accrual, o.Status, logger)
		if err != nil {
			panic(err)
		}
	}

	for _, t := range transactions {
		err = testStorage.CreateTransaction(ctx, t.UserID, t.OrderID, t.Sum, logger)
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
