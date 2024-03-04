package webapi

import (
	"encoding/json"
	"errors"
	"gophermart/internal/app/config"
	"gophermart/internal/app/entity"
	"io"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

type AccrualWebAPI struct {
	C *config.Config
}

func (webAPI *AccrualWebAPI) GetAccrualRequest(order string) (*entity.AccrualResponse, error) {
	req, err := retryablehttp.NewRequest(http.MethodGet, webAPI.C.AcrualSystemAddress+"/api/orders/"+order, nil)
	if err != nil {
		return nil, err
	}
	retryClient := retryablehttp.NewClient()
	res, err := retryClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNoContent {
		return nil, errors.New("no content")
	} else if res.StatusCode == http.StatusInternalServerError {
		return nil, errors.New("internal server error")
	}
	data, err := io.ReadAll(io.Reader(res.Body))
	if err != nil {
		return nil, err
	}
	var response *entity.AccrualResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}
	if response.Status == "REGISTERED" {
		return nil, errors.New("REGISTERED status")
	}
	return response, nil
}
