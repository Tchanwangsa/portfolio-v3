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

func TestSummarizeRunsKeepsRecentRuns(t *testing.T) {
	now := time.Date(2026, time.November, 24, 12, 0, 0, 0, melbourne)
	cutoff := now.Add(-runRetention)
	runs := []runEvent{
		{At: cutoff.Add(-time.Nanosecond), DistanceMeters: 10000},
		{At: now.Add(-time.Hour), ID: "123", Name: "Monday run", DistanceMeters: 5000, ElapsedSeconds: 1500},
		{At: cutoff, DistanceMeters: 2000},
		{At: now.Add(time.Second), DistanceMeters: 10000},
	}
	got := summarizeRuns(runs, now)
	if len(got.Runs) != 2 || !got.Runs[0].At.Equal(cutoff) || got.Runs[0].KM != 2 || got.Runs[1].KM != 5 {
		t.Fatalf("unexpected runs: %+v", got.Runs)
	}
	if got.LatestRun == nil || got.LatestRun.Name != "Monday run" || got.LatestRun.Pace != "5:00 /km" ||
		!got.LatestRun.At.Equal(now.Add(-time.Hour)) {
		t.Fatalf("unexpected latest run: %+v", got.LatestRun)
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
	if len(snapshot.Runs) != 1 || snapshot.Runs[0].KM != 5 || snapshot.LatestRun == nil || snapshot.LatestRun.URL != "https://www.strava.com/activities/123" {
		t.Fatalf("expected one public run, got %+v", snapshot)
	}
}

func TestParseRunTimeAcceptsIFTTTFormat(t *testing.T) {
	want := time.Date(2026, time.November, 24, 10, 36, 0, 0, melbourne)
	for _, value := range []string{"November 24, 2026 at 10:36AM", "2026-11-23T23:36:00Z"} {
		got, err := parseRunTime(value, melbourne)
		if err != nil || !got.Equal(want) {
			t.Errorf("parseRunTime(%q) = %v, %v; want %v", value, got, err, want)
		}
	}
	if got, err := parseRunTime("November 4, 2026 at 7:05PM", melbourne); err != nil ||
		!got.Equal(time.Date(2026, time.November, 4, 19, 5, 0, 0, melbourne)) {
		t.Errorf("single-digit day and hour parsed as %v, %v", got, err)
	}
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	if got, err := parseRunTime("November 24, 2026 at 7:00AM", tokyo); err != nil ||
		!got.Equal(time.Date(2026, time.November, 23, 22, 0, 0, 0, time.UTC)) {
		t.Errorf("Tokyo run parsed as %v, %v", got, err)
	}
	if _, err := parseRunTime("yesterday", melbourne); err == nil {
		t.Error("expected an error for an unknown date format")
	}
}

func TestRouteZone(t *testing.T) {
	for _, tc := range []struct{ url, want string }{
		// Polyline from Google's documentation, starting at 38.5, -120.2.
		{"https://maps.googleapis.com/maps/api/staticmap?size=600x300&path=color:0xff0000|weight:3|enc:_p~iF~ps|U_ulLnnqC_mqNvxq`@", "America/Los_Angeles"},
		{"https://maps.googleapis.com/maps/api/staticmap?size=600x300&path=weight%3A3%7Cenc%3A_p~iF~ps%7CU_ulLnnqC", "America/Los_Angeles"},
		{"https://maps.googleapis.com/maps/api/staticmap?markers=-37.8136,144.9631", "Australia/Melbourne"},
		{"https://maps.googleapis.com/maps/api/staticmap?center=35.6762,139.6503&zoom=12", "Asia/Tokyo"},
	} {
		zone, err := routeZone(tc.url)
		if err != nil || zone.String() != tc.want {
			t.Errorf("routeZone(%q) = %v, %v; want %s", tc.url, zone, err, tc.want)
		}
	}
	if _, err := routeZone("https://maps.googleapis.com/maps/api/staticmap?size=600x300"); err == nil {
		t.Error("expected an error for a map without coordinates")
	}
}

func TestLastRunZone(t *testing.T) {
	now := time.Now()
	runs := []runEvent{
		{At: now.Add(-2 * time.Hour), Zone: "Asia/Tokyo"},
		{At: now.Add(-time.Hour)},
		{At: now.Add(-3 * time.Hour), Zone: "Europe/London"},
	}
	if got := lastRunZone(runs).String(); got != "Asia/Tokyo" {
		t.Errorf("lastRunZone = %s, want Asia/Tokyo", got)
	}
	if got := lastRunZone(nil); got != melbourne {
		t.Errorf("lastRunZone(nil) = %s, want Melbourne", got)
	}
}
