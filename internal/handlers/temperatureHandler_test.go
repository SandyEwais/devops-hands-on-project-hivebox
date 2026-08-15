package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"errors"
	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

func TestTemperatureHandler_Integration(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("failed to load .env: %v", err)
	}

	box1 := os.Getenv("BOX_ID1")
	box2 := os.Getenv("BOX_ID2")
	box3 := os.Getenv("BOX_ID3")

	if box1 == "" || box2 == "" || box3 == "" {
		t.Fatal("BOX_ID1, BOX_ID2 and BOX_ID3 must be set")
	}

	t.Logf(
		"Testing boxes: %s, %s, %s",
		box1,
		box2,
		box3,
	)

	// Real implementation.
	provider := NewOpenSenseMapProvider()

	app := fiber.New()

	app.Get(
		"/temperature",
		TemperatureHandler(provider),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/temperature",
		nil,
	)

	resp, err := app.Test(req)

	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	t.Logf("HTTP status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			resp.StatusCode,
		)
	}

	var result struct {
		Status  int     `json:"status"`
		Message string  `json:"message"`
		Data    float64 `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	t.Logf("Response: %+v", result)

	if result.Status != http.StatusOK {
		t.Fatalf(
			"expected response status %d, got %d",
			http.StatusOK,
			result.Status,
		)
	}

	if result.Data <= 0 {
		t.Fatalf(
			"expected a positive average temperature, got %f",
			result.Data,
		)
	}
}

// MockTemperatureProvider implements TemperatureProvider.
type MockTemperatureProvider struct {
	temperatures map[string]float64
	errors       map[string]error
}

func (m *MockTemperatureProvider) GetTemperature(boxID string) (float64, error) {
	if err, exists := m.errors[boxID]; exists {
		return 0, err
	}

	return m.temperatures[boxID], nil
}

func TestTemperatureHandler(t *testing.T) {
	tests := []struct {
		name           string
		temperatures   map[string]float64
		errors         map[string]error
		expectedStatus int
		expectedData   float64
	}{
		{
			name: "three valid temperatures",

			temperatures: map[string]float64{
				"box1": 20,
				"box2": 22,
				"box3": 24,
			},

			errors: map[string]error{},

			expectedStatus: http.StatusOK,
			expectedData:   22,
		},
		{
			name: "one box fails",

			temperatures: map[string]float64{
				"box1": 20,
				"box3": 24,
			},

			errors: map[string]error{
				"box2": errors.New("API error"),
			},

			expectedStatus: http.StatusOK,
			expectedData:   22,
		},
		{
			name: "all boxes fail",

			temperatures: map[string]float64{},

			errors: map[string]error{
				"box1": errors.New("API error"),
				"box2": errors.New("API error"),
				"box3": errors.New("API error"),
			},

			expectedStatus: http.StatusServiceUnavailable,
			expectedData:   0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Arrange
			provider := &MockTemperatureProvider{
				temperatures: test.temperatures,
				errors:       test.errors,
			}

			t.Setenv("BOX_ID1", "box1")
			t.Setenv("BOX_ID2", "box2")
			t.Setenv("BOX_ID3", "box3")

			app := fiber.New()

			app.Get(
				"/temperature",
				TemperatureHandler(provider),
			)

			// Act
			req := httptest.NewRequest(
				http.MethodGet,
				"/temperature",
				nil,
			)

			resp, err := app.Test(req)

			if err != nil {
				t.Fatal(err)
			}

			defer resp.Body.Close()

			// Assert
			if resp.StatusCode != test.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					test.expectedStatus,
					resp.StatusCode,
				)
			}

			if test.expectedStatus != http.StatusOK {
				return
			}

			var result struct {
				Status int     `json:"status"`
				Data   float64 `json:"data"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatal(err)
			}

			if result.Data != test.expectedData {
				t.Fatalf(
					"expected average %.2f, got %.2f",
					test.expectedData,
					result.Data,
				)
			}
		})
	}
}