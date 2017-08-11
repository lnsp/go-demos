// Package weather provides a collection of weather service implementations.
package weather

import (
	"errors"
	"regexp"
	"strings"
	"sync"
)

// Unit is a kind of temperature measurement unit
type Unit int

var (
	// ErrInvalidTemperature is thrown when the converted temperature exceeds the allowed value range.
	ErrInvalidTemperature = errors.New("given combination of value and unit of temperature invalid")
	// ErrNotFound is thrown when a city does not exist in the database.
	ErrNotFound = errors.New("given city could not be found")
)

var (
	// reduceToTagRehexp does match everything except alphabet characters.
	reduceToTagRegexp = regexp.MustCompile("[^a-zA-Z]+")
)

// Units of measurement for temperature ranges.
const (
	Kelvin Unit = iota
	Celsius
	Fahrenheit
)

// InMemoryService is a in-memory implementation of a weather data service.
type InMemoryService struct {
	lock sync.RWMutex
	data map[string]float64
}

// Report stores the current temperature report in the memory.
func (service *InMemoryService) Report(city string, temperature float64, unit Unit) error {
	service.lock.Lock()
	defer service.lock.Unlock()

	city = reduceToTag(city)
	kelvin := convertTemperature(temperature, unit, Kelvin)
	if kelvin < 0.0 {
		return ErrInvalidTemperature
	}
	service.data[city] = kelvin
	return nil
}

// TemperatureIn retrieves the current temperature in the city.
func (service *InMemoryService) TemperatureIn(city string, unit Unit) (float64, error) {
	service.lock.RLock()
	defer service.lock.RUnlock()

	city = reduceToTag(city)
	temp, found := service.data[city]
	if !found {
		return 0, ErrNotFound
	}
	return convertTemperature(temp, Kelvin, unit), nil
}

// Cities returns a list of all stored cities.
func (service *InMemoryService) Cities() []string {
	index := 0
	cities := make([]string, len(service.data))
	for city := range service.data {
		cities[index] = city
		index++
	}
	return cities
}

// NewInMemoryService creates a new InMemoryService.
func NewInMemoryService() *InMemoryService {
	return &InMemoryService{
		lock: sync.RWMutex{},
		data: make(map[string]float64),
	}
}

// Service stores and retrieves local weather information.
type Service interface {
	Report(city string, temperature float64, unit Unit) error
	TemperatureIn(city string, unit Unit) (float64, error)
	Cities() []string
}

// reduceToTag minimizes the name by reducing it to lowercase-only alphabet characters.
func reduceToTag(name string) string {
	return strings.ToLower(reduceToTagRegexp.ReplaceAllString(name, ""))
}

// convertTemperature converts a given value between different measurement units.
func convertTemperature(temperature float64, from, to Unit) float64 {
	if from == to {
		return temperature
	}

	kelvin := temperature
	switch from {
	case Celsius:
		kelvin = temperature + 273.15
	case Fahrenheit:
		kelvin = (temperature + 459.67) * 5.0 / 9.0
	}
	temp := kelvin
	switch to {
	case Celsius:
		temp = kelvin - 273.15
	case Fahrenheit:
		temp = kelvin*9.0/5.0 - 459.67
	}
	return temp
}
