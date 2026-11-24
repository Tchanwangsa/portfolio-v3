package portfolioapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunningWeeks(t *testing.T) {
	now := time.Date(2026, time.November, 24, 12, 0, 0, 0, melbourne)
	first := firstRunningWeek(now)
	if first.Format(time.DateOnly) != "2026-10-12" {
		t.Fatalf("first week starts %s", first)
	}
	current := first.AddDate(0, 0, 42)
	runs := []runEvent{
		{At: first.Add(-time.Nanosecond), DistanceMeters: 10000},
		{At: first, DistanceMeters: 2000},
		{At: current.Add(-time.Nanosecond), DistanceMeters: 4000},
		{At: current, ID: "123", Name: "Monday run", DistanceMeters: 5000, ElapsedSeconds: 1500},
		{At: now.Add(time.Second), DistanceMeters: 10000},
	}
	got := summarizeRuns(runs, now)
	for i, want := range []float64{2, 0, 0, 0, 0, 4, 5} {
		if got.WeeklyKM[i].KM != want {
			t.Errorf("week %d has %v km, want %v", i, got.WeeklyKM[i].KM, want)
		}
	}
	if got.WeeklyKM[6].Label != "Now" || got.LatestRun == nil || got.LatestRun.Name != "Monday run" ||
		got.LatestRun.Pace != "5:00 /km" || got.LatestRun.When != "Nov 23" {
		t.Fatalf("unexpected current week or latest run: %+v", got)
	}
}

func TestRunningWebhookUpdatesPublicStatsOnce(t *testing.T) {
	webhookToken := strings.Repeat("w", 32)
	env := map[string]string{
		"STATS_PUBLISH_TOKEN":   strings.Repeat("s", 32),
		"STATS_DATA_FILE":       filepath.Join(t.TempDir(), "agentic.json"),
		"STATS_ALLOWED_ORIGINS": "https://tchanwangsa.xyz",
		"RUNS_WEBHOOK_TOKEN":    webhookToken,
	}
	api, err := New(func(key string) string { return env[key] })
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	payload, err := json.Marshal(map[string]string{
		"name": "Evening run", "activityType": "Run", "createdAt": now.Format(time.RFC3339),
		"distanceMeters": "5000", "elapsedTimeInSeconds": "1500",
		"linkToActivity": "https://www.strava.com/activities/123",
	})
	if err != nil {
		t.Fatal(err)
	}
	request := func(method, token string, want int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, runsPath, strings.NewReader(string(payload)))
		req.Header.Set("Origin", "https://tchanwangsa.xyz")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		api.ServeHTTP(response, req)
		if response.Code != want {
			t.Fatalf("%s returned HTTP %d, want %d: %s", method, response.Code, want, response.Body)
		}
		return response
	}
	request(http.MethodPost, "", http.StatusUnauthorized)
	request(http.MethodPost, env["STATS_PUBLISH_TOKEN"], http.StatusUnauthorized)
	request(http.MethodPost, webhookToken, http.StatusNoContent)
	request(http.MethodPost, webhookToken, http.StatusNoContent)
	response := request(http.MethodGet, "", http.StatusOK)
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "https://tchanwangsa.xyz" {
		t.Fatalf("unexpected CORS header: %q", got)
	}
	var snapshot runningSnapshot
	if err := json.Unmarshal(response.Body.Bytes(), &snapshot); err != nil {
		t.Fatal(err)
	}
	var total float64
	for _, week := range snapshot.WeeklyKM {
		total += week.KM
	}
	if total != 5 || snapshot.LatestRun == nil || snapshot.LatestRun.URL != "https://www.strava.com/activities/123" {
		t.Fatalf("expected one public run, got %+v", snapshot)
	}
}

func TestParseRunTimeAcceptsIFTTTFormat(t *testing.T) {
	want := time.Date(2026, time.November, 24, 10, 36, 0, 0, melbourne)
	for _, value := range []string{"November 24, 2026 at 10:36AM", "2026-11-23T23:36:00Z"} {
		got, err := parseRunTime(value)
		if err != nil || !got.Equal(want) {
			t.Errorf("parseRunTime(%q) = %v, %v; want %v", value, got, err, want)
		}
	}
	if got, err := parseRunTime("November 4, 2026 at 7:05PM"); err != nil ||
		!got.Equal(time.Date(2026, time.November, 4, 19, 5, 0, 0, melbourne)) {
		t.Errorf("single-digit day and hour parsed as %v, %v", got, err)
	}
	if _, err := parseRunTime("yesterday"); err == nil {
		t.Error("expected an error for an unknown date format")
	}
}
