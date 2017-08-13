package weather

import (
	"math"
	"testing"
)

func equal(a, b float64) bool {
	return math.Abs(a-b) < 0.00001
}

func TestConvertTemperature(t *testing.T) {
	testcases := []struct {
		value    float64
		from, to Unit
		expected float64
	}{
		{0.0, Celsius, Kelvin, 273.15},
		{0.0, Celsius, Fahrenheit, 32.0},
		{0.0, Celsius, Celsius, 0.0},
		{0.0, Fahrenheit, Kelvin, 255.3722222},
		{0.0, Fahrenheit, Fahrenheit, 0.0},
		{0.0, Fahrenheit, Celsius, -17.777777},
		{0.0, Kelvin, Kelvin, 0.0},
		{0.0, Kelvin, Fahrenheit, -459.67},
		{0.0, Kelvin, Celsius, -273.15},

		{21.0, Celsius, Kelvin, 294.15},
		{21.0, Celsius, Fahrenheit, 69.8},
		{21.0, Celsius, Celsius, 21.0},
		{69.8, Fahrenheit, Kelvin, 294.15},
		{69.8, Fahrenheit, Fahrenheit, 69.8},
		{69.8, Fahrenheit, Celsius, 21.0},
		{294.15, Kelvin, Kelvin, 294.15},
		{294.15, Kelvin, Fahrenheit, 69.8},
		{294.15, Kelvin, Celsius, 21.0},

		{-10.0, Celsius, Kelvin, 263.15},
		{-10.0, Celsius, Fahrenheit, 14.0},
		{-10.0, Celsius, Celsius, -10.0},
		{14.0, Fahrenheit, Kelvin, 263.15},
		{14.0, Fahrenheit, Fahrenheit, 14.0},
		{14.0, Fahrenheit, Celsius, -10.0},
		{263.15, Kelvin, Kelvin, 263.15},
		{263.15, Kelvin, Fahrenheit, 14.0},
		{263.15, Kelvin, Celsius, -10.0},
	}

	for _, tc := range testcases {
		result := convertTemperature(tc.value, tc.from, tc.to)
		if !equal(result, tc.expected) {
			t.Errorf("expected %f, got %f", tc.expected, result)
		}
	}
}

func TestNewInMemoryService(t *testing.T) {
	s := NewInMemoryService()
	if len(s.data) != 0 {
		t.Error("initial data store should be empty, has length %d", len(s.data))
	}
}

func TestCities(t *testing.T) {
	s := NewInMemoryService()
	s.data = map[string]float64{
		"munich":  0.0,
		"berlin":  1.0,
		"seattle": 2.0,
	}
	expected := []string{
		"munich",
		"berlin",
		"seattle",
	}

	result := s.Cities()
	if len(expected) != len(result) {
		t.Errorf("expected length %d, got %d", len(expected), len(result))
	}

	found := map[string]bool{}
	for _, city := range result {
		found[city] = true
	}
	for _, city := range expected {
		if !found[city] {
			t.Errorf("city %s not found", city)
		}
	}
}

func TestTemperatureIn(t *testing.T) {
	s := NewInMemoryService()
	s.data = map[string]float64{
		"munich":  294.15,
		"berlin":  263.15,
		"seattle": 0.0,
	}

	testcases := []struct {
		city        string
		temperature float64
		unit        Unit
		err         bool
	}{
		{"munich", 21.0, Celsius, false},
		{"ber li n", 14.0, Fahrenheit, false},
		{"seat+++tle!!", 0.0, Kelvin, false},
		{"lenggries", 0.0, Kelvin, true},
		{"", 0.0, Kelvin, true},
	}

	for _, tc := range testcases {
		result, err := s.TemperatureIn(tc.city, tc.unit)
		if err != nil && !tc.err {
			t.Errorf("expected nothing, got error %v", err)
		} else if err == nil && tc.err {
			t.Error("expected error, got nothing")
		}
		if !equal(tc.temperature, result) {
			t.Errorf("expected value %f, got %f", tc.temperature, result)
		}
	}
}

func TestReport(t *testing.T) {
	s := NewInMemoryService()
	input := []struct {
		city        string
		temperature float64
		unit        Unit
		err         bool
	}{
		{"munich", 10.0, Celsius, false},
		{"be rl in", 14.0, Fahrenheit, false},
		{"sea  ttle", 0.0, Kelvin, false},
		{"munich++?", 21.0, Celsius, false},

		{"munich", -400.0, Celsius, true},
		{"berlin", -800.0, Fahrenheit, true},
		{"seattle", -100.0, Kelvin, true},
	}
	expected := map[string]float64{
		"munich":  294.15,
		"berlin":  263.15,
		"seattle": 0.0,
	}

	for _, tc := range input {
		if err := s.Report(tc.city, tc.temperature, tc.unit); err != nil && !tc.err {
			t.Errorf("unexpected error: %v", err)
		} else if err == nil && tc.err {
			t.Error("expected error, got nothing")
		}
	}

	if len(s.data) != len(expected) {
		t.Errorf("expected item count %d, got %d", len(expected), len(s.data))
	}

	for city, temperature := range expected {
		result, ok := s.data[city]
		if !ok {
			t.Errorf("expected city %s not found", city)
		} else if !equal(result, temperature) {
			t.Errorf("expected value %f, got %f", temperature, result)
		}
	}
}
