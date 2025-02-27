package withdraw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/vadicheck/gofermart/internal/app/config"
	"github.com/vadicheck/gofermart/internal/app/constants"
	"github.com/vadicheck/gofermart/internal/app/log"
	"github.com/vadicheck/gofermart/internal/app/services/gofermart/balance"
	"github.com/vadicheck/gofermart/internal/app/storage/postgres"
	"github.com/vadicheck/gofermart/internal/app/storage/ptest"
)

func TestNew(t *testing.T) {
	type request struct {
		OrderID string `json:"order"`
		Sum     int    `json:"sum"`
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
		{ID: 1, Login: "user1", Password: "passw0rd", Balance: 1000},
		{ID: 2, Login: "user2", Password: "passw0rd", Balance: 1000},
		{ID: 3, Login: "user3", Password: "passw0rd", Balance: 1000},
	}
	orders := []orderData{
		{UserID: 1, OrderID: "123456789007", Accrual: 100, Status: "NEW"},
		{UserID: 3, OrderID: "123456789015", Accrual: 100, Status: "NEW"},
		{UserID: 3, OrderID: "123456789023", Accrual: 100, Status: "NEW"},
		{UserID: 3, OrderID: "123456789031", Accrual: 100, Status: "NEW"},
	}
	tests := []struct {
		name    string
		userID  int
		request request
		want    want
	}{
		{
			name:   "success withdraw #1",
			userID: 1,
			want: want{
				contentType:   "application/json",
				statusCode:    http.StatusOK,
				responseError: responseError{},
			},
			request: request{
				OrderID: "123456789346",
				Sum:     500,
			},
		},
		{
			name:   "order has been processed #2",
			userID: 1,
			want: want{
				contentType:   "application/json",
				statusCode:    http.StatusUnprocessableEntity,
				responseError: responseError{},
			},
			request: request{
				OrderID: "123456789346",
				Sum:     500,
			},
		},
		{
			name:   "insufficient funds #3",
			userID: 1,
			want: want{
				contentType:   "application/json",
				statusCode:    http.StatusPaymentRequired,
				responseError: responseError{},
			},
			request: request{
				OrderID: "123456789270",
				Sum:     1100,
			},
		},
		{
			name:   "incorrect order number #4",
			userID: 1,
			want: want{
				contentType:   "application/json",
				statusCode:    http.StatusUnprocessableEntity,
				responseError: responseError{},
			},
			request: request{
				OrderID: "123456780",
				Sum:     1100,
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
		err = testStorage.CreateUser(ctx, u.ID, u.Login, u.Password, u.Balance)
		if err != nil {
			panic(err)
		}
	}

	for _, o := range orders {
		err = testStorage.CreateOrder(ctx, o.UserID, o.OrderID, o.Accrual, o.Status)
		if err != nil {
			panic(err)
		}
	}

	balanceService := balance.New(storage, logger)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.request)
			if err != nil {
				fmt.Println("Ошибка кодирования в JSON:", err)
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := New(
				ctx,
				logger,
				storage,
				*validator.New(),
				balanceService,
			)

			req.Header.Set(string(constants.XUserID), strconv.Itoa(tt.userID))

			handler(w, req)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}
