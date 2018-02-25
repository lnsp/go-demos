package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/lnsp/go-demos/weather"
)

type weatherAPI struct {
	backend weather.Service
	mux     *mux.Router
}

func newWeatherAPI(backend weather.Service) *weatherAPI {
	api := &weatherAPI{
		backend: backend,
		mux:     mux.NewRouter(),
	}
	api.mux.HandleFunc("/reports", api.ListStations).Methods("GET")
	api.mux.HandleFunc("/reports", api.CreateReport).Methods("POST")
	api.mux.HandleFunc("/reports/{city}", api.ShowTemperature).Methods("GET")
	return api
}

func (api *weatherAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.mux.ServeHTTP(w, r)
}

func (api *weatherAPI) ListStations(w http.ResponseWriter, r *http.Request) {
	cities := api.backend.Cities()
	json.NewEncoder(w).Encode(cities)
}

func (api *weatherAPI) ShowTemperature(w http.ResponseWriter, r *http.Request) {
	city := mux.Vars(r)["city"]
	unit := r.URL.Query()["unit"]
	if unit == nil {
		unit = []string{"celsius"}
	}
	requestedUnit, ok := parseTemperatureUnit(unit[0])
	if !ok {
		http.Error(w, "Invalid unit of measurement", http.StatusBadRequest)
		return
	}
	report, err := api.backend.TemperatureIn(city, requestedUnit)
	if err != nil {
		http.Error(w, "Could not find weather station", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(struct {
		City        string  `json:"city"`
		Temperature float64 `json:"temperature"`
		Unit        string  `json:"unit"`
		Timestamp   int64   `json:"timestamp"`
	}{
		City:        city,
		Unit:        unit[0],
		Temperature: report.Temperature,
		Timestamp:   report.Timestamp,
	})
}

func (api *weatherAPI) CreateReport(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	report := struct {
		City        string  `json:"city"`
		Temperature float64 `json:"temperature"`
		Unit        string  `json:"unit"`
	}{}
	if err := decoder.Decode(&report); err != nil {
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}
	unit, ok := parseTemperatureUnit(report.Unit)
	if !ok {
		http.Error(w, "Unknown unit of measurement", http.StatusBadRequest)
		return
	}
	if _, err := api.backend.Report(report.City, report.Temperature, unit); err != nil {
		http.Error(w, "Failure while saving report to datastore", http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "OK")
}

func parseTemperatureUnit(unit string) (weather.Unit, bool) {
	switch strings.ToLower(unit) {
	case "kelvin":
		return weather.Kelvin, true
	case "celsius":
		return weather.Celsius, true
	case "fahrenheit":
		return weather.Fahrenheit, true
	}
	return weather.Kelvin, false
}

func main() {
	backend := weather.NewInMemoryService()
	if err := http.ListenAndServe(":8080", newWeatherAPI(backend)); err != nil {
		panic(err)
	}
}
