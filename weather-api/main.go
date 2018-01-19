package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/lnsp/go-demos/weather"
)

type weatherReport struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temperature"`
	Unit        string  `json:"unit"`
}

func listCitiesHandler(service weather.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		cities := service.Cities()
		json.NewEncoder(w).Encode(cities)
	}
}

func showTemperatureHandler(service weather.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		city := strings.TrimPrefix(r.URL.Path, "/city/")
		unit := r.URL.Query()["unit"]
		if unit == nil {
			unit = []string{"celsius"}
		}
		requestedUnit, ok := parseTemperatureUnit(unit[0])
		if !ok {
			http.Error(w, "unknown unit of measurement", http.StatusBadRequest)
			return
		}
		temp, err := service.TemperatureIn(city, requestedUnit)
		if err != nil {
			http.Error(w, "city not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(weatherReport{
			City:        city,
			Unit:        unit[0],
			Temperature: temp,
		})
	}
}

func sendReportHandler(service weather.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		report := weatherReport{}
		if err := decoder.Decode(&report); err != nil {
			http.Error(w, "failed to decode json", http.StatusInternalServerError)
			return
		}
		unit, ok := parseTemperatureUnit(report.Unit)
		if !ok {
			http.Error(w, "unknown unit of measurement", http.StatusBadRequest)
			return
		}
		if err := service.Report(report.City, report.Temperature, unit); err != nil {
			http.Error(w, "failed to save report", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, path.Join("/city", report.City))
	}
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
	service := weather.NewInMemoryService()
	http.HandleFunc("/city", listCitiesHandler(service))
	http.HandleFunc("/city/", showTemperatureHandler(service))
	http.HandleFunc("/report", sendReportHandler(service))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
