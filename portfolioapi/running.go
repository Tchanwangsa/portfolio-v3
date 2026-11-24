package portfolioapi

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "time/tzdata"
)

const runsPath = "/api/runs"
const weekCount = 7

var stravaActivityLink = regexp.MustCompile(`^https?://(?:www\.)?strava\.com/activities/([1-9][0-9]{0,19})/?$`)

// IFTTT's CreatedAt ingredient looks like "November 24, 2026 at 10:36AM" in the
// IFTTT account's time zone, which is assumed to be Melbourne.
var iftttDateLayouts = []string{"January 2, 2006 at 3:04PM", "Jan 2, 2006 at 3:04PM"}
var melbourne = loadMelbourne()

func loadMelbourne() *time.Location {
	location, err := time.LoadLocation("Australia/Melbourne")
	if err != nil {
		panic(err)
	}
	return location
}

type runningAPI struct {
	data      document
	tokenHash [sha256.Size]byte
}

type runEvent struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	At             time.Time `json:"at"`
	DistanceMeters float64   `json:"distanceMeters"`
	ElapsedSeconds int       `json:"elapsedSeconds"`
}

type weekDistance struct {
	Label string  `json:"label"`
	KM    float64 `json:"km"`
}

type latestRun struct {
	Name       string  `json:"name"`
	DistanceKM float64 `json:"distanceKm"`
	Duration   string  `json:"duration"`
	Pace       string  `json:"pace"`
	When       string  `json:"when"`
	URL        string  `json:"url"`
}

type runningSnapshot struct {
	WeeklyKM  [weekCount]weekDistance `json:"weeklyKm"`
	LatestRun *latestRun              `json:"latestRun"`
}

func firstRunningWeek(now time.Time) time.Time {
	local := now.In(melbourne)
	day := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, melbourne)
	daysSinceMonday := (int(day.Weekday()) + 6) % 7
	return day.AddDate(0, 0, -daysSinceMonday-7*(weekCount-1))
}

func summarizeRuns(runs []runEvent, now time.Time) runningSnapshot {
	var snapshot runningSnapshot
	first := firstRunningWeek(now)
	var latest *runEvent
	for i := range snapshot.WeeklyKM {
		label := fmt.Sprintf("W-%d", weekCount-1-i)
		if i == weekCount-1 {
			label = "Now"
		}
		snapshot.WeeklyKM[i].Label = label
	}
	for i := range runs {
		run := &runs[i]
		if run.At.Before(first) || run.At.After(now) {
			continue
		}
		for week := range snapshot.WeeklyKM {
			if run.At.Before(first.AddDate(0, 0, 7*(week+1))) {
				snapshot.WeeklyKM[week].KM += run.DistanceMeters / 1000
				break
			}
		}
		if latest == nil || run.At.After(latest.At) || (run.At.Equal(latest.At) && run.ID > latest.ID) {
			latest = run
		}
	}
	if latest != nil {
		pace := "—"
		if latest.DistanceMeters >= 1 {
			seconds := int(math.Round(float64(latest.ElapsedSeconds) * 1000 / latest.DistanceMeters))
			pace = fmt.Sprintf("%d:%02d /km", seconds/60, seconds%60)
		}
		snapshot.LatestRun = &latestRun{
			Name:       latest.Name,
			DistanceKM: math.Round(latest.DistanceMeters/10) / 100,
			Duration:   fmt.Sprintf("%dm %ds", latest.ElapsedSeconds/60, latest.ElapsedSeconds%60),
			Pace:       pace,
			When:       latest.At.In(melbourne).Format("Jan 2"),
			URL:        "https://www.strava.com/activities/" + latest.ID,
		}
	}
	return snapshot
}

func (api *runningAPI) get(w http.ResponseWriter, r *http.Request) {
	var runs []runEvent
	data, err := api.data.read(r.Context())
	if errors.Is(err, os.ErrNotExist) {
		err = nil
	} else if err == nil {
		err = json.Unmarshal(data, &runs)
	}
	if err != nil {
		runStatsError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summarizeRuns(runs, time.Now()))
}

func (api *runningAPI) receive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if !hasBearerToken(r, api.tokenHash) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		http.Error(w, "Invalid webhook token", http.StatusUnauthorized)
		return
	}
	var payload struct {
		Name                 string `json:"name"`
		ActivityType         string `json:"activityType"`
		CreatedAt            string `json:"createdAt"`
		DistanceMeters       string `json:"distanceMeters"`
		ElapsedTimeInSeconds string `json:"elapsedTimeInSeconds"`
		LinkToActivity       string `json:"linkToActivity"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil || decoder.Decode(new(any)) != io.EOF {
		http.Error(w, "Expected one Strava webhook event", http.StatusBadRequest)
		return
	}
	switch strings.ToLower(strings.TrimSpace(payload.ActivityType)) {
	case "run", "trail run", "trailrun", "virtual run", "virtualrun":
	default:
		log.Printf("running webhook ignored activity type %q", payload.ActivityType)
		http.Error(w, "Expected a run activity", http.StatusUnprocessableEntity)
		return
	}
	now := time.Now()
	activityID, linkErr := stravaActivityID(payload.LinkToActivity)
	at, dateErr := parseRunTime(payload.CreatedAt)
	distance, distanceErr := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(payload.DistanceMeters), ",", ""), 64)
	elapsedFloat, elapsedErr := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(payload.ElapsedTimeInSeconds), ",", ""), 64)
	elapsed := int(math.Round(elapsedFloat))
	name := strings.TrimSpace(payload.Name)
	var invalid []string
	if linkErr != nil {
		invalid = append(invalid, "linkToActivity")
	}
	if dateErr != nil || at.After(now.Add(5*time.Minute)) {
		invalid = append(invalid, "createdAt")
	}
	if distanceErr != nil || math.IsNaN(distance) || distance < 0 || distance > 10_000_000 {
		invalid = append(invalid, "distanceMeters")
	}
	if elapsedErr != nil || math.IsNaN(elapsedFloat) || elapsed < 0 || elapsed > 7*24*60*60 {
		invalid = append(invalid, "elapsedTimeInSeconds")
	}
	if name == "" || len(name) > 1024 {
		invalid = append(invalid, "name")
	}
	if len(invalid) > 0 {
		log.Printf("running webhook rejected %v: createdAt=%q distanceMeters=%q elapsedTimeInSeconds=%q linkToActivity=%q linkErr=%v",
			invalid, payload.CreatedAt, payload.DistanceMeters, payload.ElapsedTimeInSeconds, payload.LinkToActivity, linkErr)
		http.Error(w, "Invalid run fields: "+strings.Join(invalid, ", "), http.StatusUnprocessableEntity)
		return
	}
	first := firstRunningWeek(now)
	if at.Before(first) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	run := runEvent{ID: activityID, Name: name, At: at.UTC(), DistanceMeters: distance, ElapsedSeconds: elapsed}
	err := api.data.update(r.Context(), func(current []byte) ([]byte, error) {
		var runs []runEvent
		if current != nil {
			if err := json.Unmarshal(current, &runs); err != nil {
				return nil, err
			}
		}
		recent := runs[:0]
		for _, existing := range runs {
			if existing.ID == run.ID {
				return nil, nil
			}
			if !existing.At.Before(first) {
				recent = append(recent, existing)
			}
		}
		return json.Marshal(append(recent, run))
	})
	if err != nil {
		runStatsError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseRunTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if at, err := time.Parse(time.RFC3339, value); err == nil {
		return at, nil
	}
	var err error
	for _, layout := range iftttDateLayouts {
		var at time.Time
		if at, err = time.ParseInLocation(layout, value, melbourne); err == nil {
			return at, nil
		}
	}
	return time.Time{}, err
}

func stravaActivityID(link string) (string, error) {
	match := stravaActivityLink.FindStringSubmatch(strings.TrimSpace(link))
	if len(match) != 2 {
		return "", fmt.Errorf("not a Strava activity link: %q", link)
	}
	return match[1], nil
}

func runStatsError(w http.ResponseWriter, err error) {
	log.Printf("running stats: %v", err)
	http.Error(w, "Running stats are unavailable", http.StatusInternalServerError)
}
