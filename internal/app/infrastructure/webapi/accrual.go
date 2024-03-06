package webapi

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/entity"

	"github.com/hashicorp/go-retryablehttp"
)

type AccrualWebAPI struct {
	Config *config.Config
}

func (webAPI *AccrualWebAPI) GetAccrualRequest(order string) (*entity.AccrualResponse, error) {
	req, err := retryablehttp.NewRequest(http.MethodGet, webAPI.Config.AcrualSystemAddress+"/api/orders/"+order, nil)
	if err != nil {
		return nil, err
	}
	retryClient := retryablehttp.NewClient()
	res, err := retryClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
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
	log.Print("request")
	log.Print(response.Order, response.Status)
	if err != nil {
		return nil, err
	}
	if response.Status == "REGISTERED" {
		return nil, errors.New("REGISTERED status")
	}
	if response.Status == "PROCESSING" {
		return nil, errors.New("PROCESSING status")
	}
	return response, nil
}
