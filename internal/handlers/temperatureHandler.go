package handlers

import (
	"encoding/json"
	"fmt"
	"hivebox/internal/responses"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
)

type fetchResult struct {
	ID   string          `json:"id"`
	Data json.RawMessage `json:"data,omitempty"`
	Err  string          `json:"error,omitempty"`
}

type temperatureResult struct {
	ID          string
	Temperature float64
	Err         error
}

func TemperatureHandler(c fiber.Ctx) error {
	id1 := os.Getenv("BOX_ID1")
	id2 := os.Getenv("BOX_ID2")
	id3 := os.Getenv("BOX_ID3")

	if id1 == "" || id2 == "" || id3 == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.APIResponse{
			Status:  500,
			Message: "Internal Server Error: sensor IDs missing",
			Data:    nil,
		})
	}

	resultsCh := make(chan temperatureResult, 3)

	go fetchTemperatureData(id1, resultsCh)
	go fetchTemperatureData(id2, resultsCh)
	go fetchTemperatureData(id3, resultsCh)

	var sum float64
	var validCount int

	for i := 0; i < 3; i++ {
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

	return c.JSON(responses.APIResponse{
		Status:  200,
		Message: "Average temperature retrieved",
		Data:    average,
	})
}

func fetchTemperatureData(boxID string, resultsCh chan<- temperatureResult) {
	url := fmt.Sprintf(
		"https://api.opensensemap.org/boxes/%s",
		boxID,
	)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		resultsCh <- temperatureResult{
			ID:  boxID,
			Err: err,
		}
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		resultsCh <- temperatureResult{
			ID:  boxID,
			Err: fmt.Errorf("API returned status %d", resp.StatusCode),
		}
		return
	}

	var box struct {
		Sensors []struct {
			Title           string `json:"title"`
			LastMeasurement struct {
				CreatedAt time.Time `json:"createdAt"`
				Value     float64   `json:"value,string"`
			} `json:"lastMeasurement"`
		} `json:"sensors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&box); err != nil {
		resultsCh <- temperatureResult{
			ID:  boxID,
			Err: err,
		}
		return
	}

	for _, sensor := range box.Sensors {
		if sensor.Title != "Temperatur" {
			continue
		}

		if time.Since(sensor.LastMeasurement.CreatedAt) > time.Hour {
			log.Printf(
				"Box %s: temperature measurement is too old: %s",
				boxID,
				sensor.LastMeasurement.CreatedAt,
			)
			resultsCh <- temperatureResult{
				ID:  boxID,
				Err: fmt.Errorf("temperature measurement is older than 1 hour"),
			}
			return
		}
		log.Printf(
			"Box %s: temperature = %.2f°C, measured at %s",
			boxID,
			sensor.LastMeasurement.Value,
			sensor.LastMeasurement.CreatedAt,
		)

		resultsCh <- temperatureResult{
			ID:          boxID,
			Temperature: sensor.LastMeasurement.Value,
		}
		return
	}

	resultsCh <- temperatureResult{
		ID:  boxID,
		Err: fmt.Errorf("temperature sensor not found"),
	}
}
