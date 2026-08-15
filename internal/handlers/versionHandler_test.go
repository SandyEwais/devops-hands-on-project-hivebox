package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestVersionHandler(t *testing.T) {
	type testCase struct {
		expectedStatus int
		expectedBody   string
	}
	t.Run("Test application version", func(t *testing.T) {
		test := &testCase{
			expectedStatus: 200,
			expectedBody:   "version: 1.0.0",
		}
		// Arrange
		app := fiber.New()
		app.Get("/version", VersionHandler)

		req := httptest.NewRequest(
			http.MethodGet,
			"/version",
			nil,
		)

		// Act
		resp, err := app.Test(req)

		if err != nil {
			t.Fatal(err)
		}

		defer resp.Body.Close()
		if resp.StatusCode != test.expectedStatus {
			t.Errorf(
				"expected status %d, got %d",
				test.expectedStatus,
				resp.StatusCode,
			)
		}

		body, err := io.ReadAll(resp.Body)

		if err != nil {
			t.Fatal(err)
		}

		if string(body) != test.expectedBody {
			t.Errorf(
				"expected body %q, got %q",
				test.expectedBody,
				string(body),
			)
		}
	})
}