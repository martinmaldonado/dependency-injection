package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mitchellh/mapstructure"
	di "mercadolibre.com/di/practice"
	"mercadolibre.com/di/practice/entities"
)

type beerForecastCalculator interface {
	Get(rp *entities.RequestParams) ([]*entities.BeerForecast, error)
}

type BeerForecastFetcher struct {
	bfCalculator beerForecastCalculator
}

func NewBeerForecastResolver(bfCalculator beerForecastCalculator) *BeerForecastFetcher {
	return &BeerForecastFetcher{bfCalculator: bfCalculator}
}

func (bp *BeerForecastFetcher) Do(w io.Writer, r *http.Request) (*entities.HandlerResult, error) {
	queries := r.URL.Query()

	rp, err := getValuesFromVars(queries)
	if err != nil {
		return nil, err
	}

	forecastBeerData, err := bp.bfCalculator.Get(rp)
	if err != nil {
		return nil, err
	}

	return &entities.HandlerResult{
		Status: http.StatusOK,
		Body:   forecastBeerData,
	}, nil
}

func getValuesFromVars(queries url.Values) (*entities.RequestParams, error) {
	stringParams := []string{"country", "city", "state"}
	uintParams := []string{"attendees", "pack_units", "forecast_days"}

	paramsMap := make(map[string]interface{}, len(stringParams)+len(uintParams))

	for _, p := range stringParams {
		name := p
		value := queries.Get(name)
		if value == "" {
			return nil, &di.CustomError{
				Cause:   di.ErrBadRequest,
				Message: fmt.Sprintf("param with name %s must not be empty", name),
			}
		}

		paramsMap[name] = value
	}

	for _, up := range uintParams {
		name := up
		value := queries.Get(name)
		if value == "" {
			return nil, &di.CustomError{
				Cause:   di.ErrBadRequest,
				Message: fmt.Sprintf("param with name %s must not be empty", name),
			}
		}

		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, &di.CustomError{
				Cause:   di.ErrBadRequest,
				Message: fmt.Sprintf("param with name %s must be an integer value", name),
			}
		}

		paramsMap[name] = parsed
	}

	var rq entities.RequestParams
	if err := mapstructure.Decode(paramsMap, &rq); err != nil {
		return nil, &di.CustomError{
			Cause:   err,
			Message: "could not read query params",
		}
	}

	return &rq, nil
}
