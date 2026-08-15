package handlers

import (
	"encoding/json"
	"fmt"
	"hivebox/internal/responses"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

var HTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

type TemperatureProvider interface {
	GetTemperature(boxID string) (float64, error)
}

type OpenSenseMapProvider struct {
	client *http.Client
}

func NewOpenSenseMapProvider() *OpenSenseMapProvider {
	return &OpenSenseMapProvider{
		client: HTTPClient,
	}
}

func (p *OpenSenseMapProvider) GetTemperature(boxID string) (float64, error) {
	url := fmt.Sprintf(
		"https://api.opensensemap.org/boxes/%s",
		boxID,
	)

	resp, err := p.client.Get(url)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf(
			"API returned status %d",
			resp.StatusCode,
		)
	}

	var box struct {
		Sensors []struct {
			Title string `json:"title"`

			LastMeasurement struct {
				CreatedAt time.Time `json:"createdAt"`
				Value     float64   `json:"value,string"`
			} `json:"lastMeasurement"`
		} `json:"sensors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&box); err != nil {
		return 0, err
	}

	for _, sensor := range box.Sensors {
		titleLower := strings.ToLower(sensor.Title)

		if !strings.Contains(titleLower, "temp") {
			continue
		}

		if time.Since(sensor.LastMeasurement.CreatedAt) > time.Hour {
			log.Printf(
				"Box %s: temperature measurement is too old: %s",
				boxID,
				sensor.LastMeasurement.CreatedAt,
			)

			return 0, fmt.Errorf(
				"temperature measurement is older than 1 hour",
			)
		}

		log.Printf(
			"Box %s: temperature = %.2f°C, measured at %s",
			boxID,
			sensor.LastMeasurement.Value,
			sensor.LastMeasurement.CreatedAt,
		)

		return sensor.LastMeasurement.Value, nil
	}

	return 0, fmt.Errorf("temperature sensor not found")
}

type temperatureResult struct {
	ID          string
	Temperature float64
	Err         error
}

func TemperatureHandler(provider TemperatureProvider) fiber.Handler {
	return func(c fiber.Ctx) error {
		id1 := os.Getenv("BOX_ID1")
		id2 := os.Getenv("BOX_ID2")
		id3 := os.Getenv("BOX_ID3")

		if id1 == "" || id2 == "" || id3 == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(
				responses.APIResponse{
					Status:  500,
					Message: "Internal Server Error: sensor IDs missing",
					Data:    nil,
				},
			)
		}

		boxIDs := []string{id1, id2, id3}

		resultsCh := make(chan temperatureResult, len(boxIDs))

		for _, boxID := range boxIDs {
			go func(id string) {
				temperature, err := provider.GetTemperature(id)

				resultsCh <- temperatureResult{
					ID:          id,
					Temperature: temperature,
					Err:         err,
				}
			}(boxID)
		}

		var sum float64
		var validCount int

		for range boxIDs {
			result := <-resultsCh

			if result.Err != nil {
				continue
			}

			sum += result.Temperature
			validCount++
		}

		if validCount == 0 {
			return c.Status(fiber.StatusServiceUnavailable).JSON(
				responses.APIResponse{
					Status:  503,
					Message: "No valid temperature measurements available",
					Data:    nil,
				},
			)
		}

		average := sum / float64(validCount)

		return c.JSON(
			responses.APIResponse{
				Status:  200,
				Message: "Average temperature retrieved",
				Data:    average,
			},
		)
	}
}