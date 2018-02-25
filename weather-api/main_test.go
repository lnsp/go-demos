package main

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lnsp/go-demos/weather"
)

type ServiceMock struct{}

func (s *ServiceMock) Cities() []string {
	return []string{"munich"}
}

func (s *ServiceMock) Report(c string, t float64, u weather.Unit) (int64, error) {
	if c == "" {
		return 0, errors.New("no city reported")
	}
	return time.Now().Unix(), nil
}

func (s *ServiceMock) mockReport(t float64) weather.Report {
	return weather.Report{
		Temperature: t,
		Timestamp:   1,
	}
}

func (s *ServiceMock) TemperatureIn(c string, u weather.Unit) (weather.Report, error) {
	if c != "munich" {
		return weather.Report{}, errors.New("city not found")
	}
	switch u {
	case weather.Kelvin:
		return s.mockReport(294.15), nil
	case weather.Fahrenheit:
		return s.mockReport(69.8), nil
	case weather.Celsius:
		return s.mockReport(21.0), nil
	}
	return s.mockReport(0.0), errors.New("unknown unit")
}

func TestListCities(t *testing.T) {
	testcases := []struct {
		req    *http.Request
		status int
		text   string
	}{
		{httptest.NewRequest("GET", "/reports", nil), http.StatusOK, `["munich"]` + "\n"},
	}

	for _, tc := range testcases {
		rr := httptest.NewRecorder()
		handler := newWeatherAPI(&ServiceMock{})
		handler.ServeHTTP(rr, tc.req)

		if status := rr.Code; status != tc.status {
			t.Errorf("expected status %d, got %d", tc.status, status)
		}

		if text := rr.Body.String(); text != tc.text {
			t.Errorf("expected response %s, got %s", tc.text, text)
		}
	}
}

func TestShowTemperature(t *testing.T) {
	testcases := []struct {
		req    *http.Request
		status int
		text   string
	}{
		{httptest.NewRequest("POST", "/reports/munich", nil), http.StatusMethodNotAllowed, ""},
		{httptest.NewRequest("GET", "/reports/munich", nil), http.StatusOK, `{"city":"munich","temperature":21,"unit":"celsius","timestamp":1}` + "\n"},
		{httptest.NewRequest("GET", "/reports/munich?unit=kelvin", nil), http.StatusOK, `{"city":"munich","temperature":294.15,"unit":"kelvin","timestamp":1}` + "\n"},
		{httptest.NewRequest("GET", "/reports/munich?unit=celsius", nil), http.StatusOK, `{"city":"munich","temperature":21,"unit":"celsius","timestamp":1}` + "\n"},
		{httptest.NewRequest("GET", "/reports/munich?unit=fahrenheit", nil), http.StatusOK, `{"city":"munich","temperature":69.8,"unit":"fahrenheit","timestamp":1}` + "\n"},
		{httptest.NewRequest("GET", "/reports/munich?unit=useless", nil), http.StatusBadRequest, "Invalid unit of measurement\n"},
		{httptest.NewRequest("GET", "/reports/", nil), http.StatusNotFound, "404 page not found\n"},
		{httptest.NewRequest("GET", "/reports/seattle", nil), http.StatusNotFound, "Could not find weather station\n"},
	}

	for _, tc := range testcases {
		rr := httptest.NewRecorder()
		handler := newWeatherAPI(&ServiceMock{})
		handler.ServeHTTP(rr, tc.req)

		if status := rr.Code; status != tc.status {
			t.Errorf("expected status %d, got %d", tc.status, status)
		}

		if text := rr.Body.String(); text != tc.text {
			t.Errorf("expected response %s, got %s", tc.text, text)
		}
	}
}

func TestSendReport(t *testing.T) {
	testcases := []struct {
		req    *http.Request
		status int
		text   string
	}{
		{httptest.NewRequest("GET", "/report", nil), http.StatusMethodNotAllowed, "method not allowed\n"},
		{httptest.NewRequest("POST", "/report", bytes.NewBufferString(`{
			"city": "munich",
			"temperature": 21.0,
			"unit": "celsius"
		}`)), http.StatusOK, "/city/munich\n"},
		{httptest.NewRequest("POST", "/report", bytes.NewBufferString(`{
			"city": "munich",
			"temperature": 69.8,
			"unit": "fahrenheit"
		}`)), http.StatusOK, "/city/munich\n"},
		{httptest.NewRequest("POST", "/report", bytes.NewBufferString(`{
			"city": "munich",
			"temperature": 69.8,
			"unit": "useless"
		}`)), http.StatusBadRequest, "unknown unit of measurement\n"},
		{httptest.NewRequest("POST", "/report", bytes.NewBufferString(`{
			"city": "",
			"temperature": 21.0,
			"unit": "celsius"
		}`)), http.StatusInternalServerError, "failed to save report\n"},
		{httptest.NewRequest("POST", "/report", bytes.NewBufferString(`{
			"city": "",
			"temperat.0,
			"unit": "celsius"
		}`)), http.StatusInternalServerError, "failed to decode json\n"},
	}

	for _, tc := range testcases {
		rr := httptest.NewRecorder()
		handler := newWeatherAPI(&ServiceMock{})
		handler.ServeHTTP(rr, tc.req)

		if status := rr.Code; status != tc.status {
			t.Errorf("expected status %d, got %d", tc.status, status)
		}

		if text := rr.Body.String(); text != tc.text {
			t.Errorf("expected response %s, got %s", tc.text, text)
		}
	}
}
