package accrualservice

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	jsoniter "github.com/json-iterator/go"

	"github.com/vadicheck/gofermart/pkg/logger"
)

const GetOrderURL = "/api/orders/%s"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type accrualsServiceAPI struct {
	baseURL    string
	httpClient HTTPClient
	logger     logger.LogClient
}

type Service interface {
	GetOrder(ctx context.Context, orderID string) (*GetOrderResponse, error)
}

func New(
	httpClient HTTPClient,
	accrualSystemAddress string,
	logger logger.LogClient,
) Service {
	return &accrualsServiceAPI{
		baseURL:    accrualSystemAddress,
		httpClient: httpClient,
		logger:     logger,
	}
}

func (f *accrualsServiceAPI) GetOrder(ctx context.Context, orderID string) (*GetOrderResponse, error) {
	buf, err := f.doGet(ctx, fmt.Sprintf(GetOrderURL, orderID))
	if err != nil {
		return nil, fmt.Errorf("failed to do get: %w", err)
	}

	if buf == nil {
		return nil, nil
	}

	response := &GetOrderResponse{}

	err = jsoniter.Unmarshal(buf, response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}

func (f *accrualsServiceAPI) doGet(ctx context.Context, path string) ([]byte, error) {
	url := f.makeURL(path)
	reqWithCtx, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	respRaw, err := f.httpClient.Do(reqWithCtx)
	if err != nil {
		return nil, err
	}

	if respRaw == nil {
		return nil, errors.New(`nil response value`)
	}

	if respRaw.StatusCode == http.StatusNoContent {
		f.logger.Info("no content response", "url", url)
		return nil, nil
	}

	buf, err := io.ReadAll(respRaw.Body)
	if err != nil {
		return nil, err
	}

	return buf, respRaw.Body.Close()
}

func (f *accrualsServiceAPI) makeURL(path string) string {
	return f.baseURL + path
}
