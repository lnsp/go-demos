package main

import (
	"fmt"
	"net/http"

	"github.com/lnsp/go-demos/weather"
)

var (
	service weather.Service
)

func listCities(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "/city")
}

func showTemperature(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "/city/")
}

func sendReport(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "/report")
}

func main() {
	service = weather.NewInMemoryService()
	http.HandleFunc("/city", listCities)
	http.HandleFunc("/city/", showTemperature)
	http.HandleFunc("/report", sendReport)
	if err := http.ListenAndServe("localhost:8080", nil); err != nil {
		panic(err)
	}
}
