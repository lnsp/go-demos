package main

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lnsp/go-demos/weather"
)

type ServiceMock struct{}

func (s *ServiceMock) Cities() []string {
	return []string{"munich"}
}

func (s *ServiceMock) Report(c string, t float64, u weather.Unit) error {
	if c == "" {
		return errors.New("no city reported")
	}
	return nil
}

func (s *ServiceMock) TemperatureIn(c string, u weather.Unit) (float64, error) {
	if c != "munich" {
		return 0.0, errors.New("city not found")
	}
	switch u {
	case weather.Kelvin:
		return 294.15, nil
	case weather.Fahrenheit:
		return 69.8, nil
	case weather.Celsius:
		return 21.0, nil
	}
	return 0.0, errors.New("unknown unit")
}

func TestListCities(t *testing.T) {
	testcases := []struct {
		req    *http.Request
		status int
		text   string
	}{
		{httptest.NewRequest("POST", "/city", nil), http.StatusMethodNotAllowed, "method not allowed\n"},
		{httptest.NewRequest("GET", "/city", nil), http.StatusOK, `["munich"]` + "\n"},
	}

	for _, tc := range testcases {
		rr := httptest.NewRecorder()
		handler := listCitiesHandler(&ServiceMock{})
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
		{httptest.NewRequest("POST", "/city/munich", nil), http.StatusMethodNotAllowed, "method not allowed\n"},
		{httptest.NewRequest("GET", "/city/munich", nil), http.StatusOK, `{"city":"munich","temperature":21,"unit":"celsius"}` + "\n"},
		{httptest.NewRequest("GET", "/city/munich?unit=kelvin", nil), http.StatusOK, `{"city":"munich","temperature":294.15,"unit":"kelvin"}` + "\n"},
		{httptest.NewRequest("GET", "/city/munich?unit=celsius", nil), http.StatusOK, `{"city":"munich","temperature":21,"unit":"celsius"}` + "\n"},
		{httptest.NewRequest("GET", "/city/munich?unit=fahrenheit", nil), http.StatusOK, `{"city":"munich","temperature":69.8,"unit":"fahrenheit"}` + "\n"},
		{httptest.NewRequest("GET", "/city/munich?unit=useless", nil), http.StatusBadRequest, "unknown unit of measurement\n"},
		{httptest.NewRequest("GET", "/city/", nil), http.StatusNotFound, "city not found\n"},
		{httptest.NewRequest("GET", "/city/seattle", nil), http.StatusNotFound, "city not found\n"},
	}

	for _, tc := range testcases {
		rr := httptest.NewRecorder()
		handler := showTemperatureHandler(&ServiceMock{})
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
		handler := sendReportHandler(&ServiceMock{})
		handler.ServeHTTP(rr, tc.req)

		if status := rr.Code; status != tc.status {
			t.Errorf("expected status %d, got %d", tc.status, status)
		}

		if text := rr.Body.String(); text != tc.text {
			t.Errorf("expected response %s, got %s", tc.text, text)
		}
	}
}
