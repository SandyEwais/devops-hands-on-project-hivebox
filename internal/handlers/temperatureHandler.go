package handlers

import (
	"encoding/json"
	"fmt"
	"hivebox/internal/responses"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// TemperatureHandler reads up to 3 sensor IDs from environment variables,
// launches fetchTemperatureData concurrently for each ID, and returns the
// aggregated results.
// Supported env formats:
// - SENSOR_IDS="id1,id2,id3"
// - or SENSOR_ID_1, SENSOR_ID_2, SENSOR_ID_3
func TemperatureHandler(c fiber.Ctx) error {
	// gather IDs from env
	idsEnv := os.Getenv("SENSOR_IDS")
	var ids []string
	if idsEnv != "" {
		for _, s := range strings.Split(idsEnv, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				ids = append(ids, s)
			}
		}
	} else {
		for i := 1; i <= 3; i++ {
			k := fmt.Sprintf("SENSOR_ID_%d", i)
			if v := os.Getenv(k); v != "" {
				ids = append(ids, v)
			}
		}
	}

	if len(ids) == 0 {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(responses.APIResponse{
			Status:  400,
			Message: "No sensor IDs provided in environment",
			Data:    nil,
		})
	}

	// limit to 3 IDs
	if len(ids) > 3 {
		ids = ids[:3]
	}

	type fetchResult struct {
		ID   string      `json:"id"`
		Data interface{} `json:"data,omitempty"`
		Err  string      `json:"error,omitempty"`
	}

	resultsCh := make(chan fetchResult, len(ids))

	// fire a goroutine per sensor ID
	for _, id := range ids {
		id := id // capture
		go func(sensorID string) {
			data, err := fetchTemperatureData(sensorID)
			res := fetchResult{ID: sensorID}
			if err != nil {
				res.Err = err.Error()
			} else {
				res.Data = data
			}
			resultsCh <- res
		}(id)
	}

	// collect results
	var results []fetchResult
	for i := 0; i < len(ids); i++ {
		r := <-resultsCh
		results = append(results, r)
	}

	// return aggregated JSON
	return c.JSON(responses.APIResponse{
		Status:  200,
		Message: "Temperature data retrieved",
		Data:    results,
	})
}

// fetchTemperatureData queries the external API for the given sensor ID and
// returns the raw JSON response as a json.RawMessage for inclusion in the
// aggregated response.
func fetchTemperatureData(sensorID string) (interface{}, error) {
	url := fmt.Sprintf("https://api.opensensemap.org/boxes/%s", sensorID)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Keep raw JSON so the handler can pass it through
	var raw json.RawMessage = body
	return raw, nil
}
