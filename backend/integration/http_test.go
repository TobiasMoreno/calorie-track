package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"calorietrack/backend/internal/application"
	"calorietrack/backend/internal/domain"
	"calorietrack/backend/internal/infrastructure/postgres"
	httptransport "calorietrack/backend/internal/transport/http"
)

func TestMealEntryWorkflowOverHTTP(t *testing.T) {
	databaseURL := os.Getenv("CALORIETRACK_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("CALORIETRACK_TEST_DATABASE_URL is not configured")
	}
	db, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repository := postgres.NewRepository(db)
	service := application.NewService(repository)
	server := httptest.NewServer(httptransport.NewHandler(service, "http://localhost:5173").Routes())
	defer server.Close()

	date := time.Now().UTC().AddDate(30, 0, int(time.Now().UnixNano()%300)).Format("2006-01-02")
	defer db.Exec(`DELETE FROM daily_days WHERE local_date=$1::date`, date)

	var dashboard domain.Dashboard
	getJSON(t, server.URL+"/api/dashboard?date="+date, &dashboard)
	if len(dashboard.Tasks) != 4 {
		t.Fatalf("tasks = %d, want 4", len(dashboard.Tasks))
	}

	var foods []domain.Food
	getJSON(t, server.URL+"/api/foods?q=milanesa", &foods)
	if len(foods) != 1 || len(foods[0].Portions) == 0 {
		t.Fatalf("expected seeded milanesa with portions, got %#v", foods)
	}
	portion := foods[0].Portions[0]
	for _, candidate := range foods[0].Portions {
		if candidate.IsDefault {
			portion = candidate
			break
		}
	}

	payload := domain.CreateMealEntryInput{
		Date:   date,
		TaskID: dashboard.Tasks[0].ID,
		Items: []domain.CreateItemInput{{
			FoodID:        foods[0].ID,
			Quantity:      2,
			Unit:          portion.Unit,
			PortionID:     &portion.ID,
			CookingMethod: "Al horno",
		}},
	}
	var entry domain.MealEntry
	requestJSON(t, http.MethodPost, server.URL+"/api/meal-entries", payload, http.StatusCreated, &entry)
	if entry.TotalCalories != 831.6 {
		t.Fatalf("calories = %v, want 831.6", entry.TotalCalories)
	}

	getJSON(t, server.URL+"/api/dashboard?date="+date, &dashboard)
	if dashboard.Tasks[0].Status != domain.TaskRegistered || dashboard.ConsumedCalories != 831.6 {
		t.Fatalf("unexpected registered dashboard: %#v", dashboard)
	}

	requestJSON(t, http.MethodDelete, fmt.Sprintf("%s/api/meal-entries/%d", server.URL, entry.ID), nil, http.StatusNoContent, nil)
	getJSON(t, server.URL+"/api/dashboard?date="+date, &dashboard)
	if dashboard.Tasks[0].Status != domain.TaskPending || dashboard.ConsumedCalories != 0 {
		t.Fatalf("unexpected restored dashboard: %#v", dashboard)
	}
}

func getJSON(t *testing.T, url string, target any) {
	t.Helper()
	requestJSON(t, http.MethodGet, url, nil, http.StatusOK, target)
}

func requestJSON(t *testing.T, method, url string, body any, expectedStatus int, target any) {
	t.Helper()
	var payload bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&payload).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, url, &payload)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != expectedStatus {
		t.Fatalf("%s %s status = %d, want %d", method, url, response.StatusCode, expectedStatus)
	}
	if target != nil {
		if err := json.NewDecoder(response.Body).Decode(target); err != nil {
			t.Fatal(err)
		}
	}
}
