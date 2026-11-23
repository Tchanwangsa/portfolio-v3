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
		http.Error(w, "Expected a run activity", http.StatusUnprocessableEntity)
		return
	}
	match := stravaActivityLink.FindStringSubmatch(strings.TrimSpace(payload.LinkToActivity))
	at, dateErr := time.Parse(time.RFC3339, strings.TrimSpace(payload.CreatedAt))
	distance, distanceErr := strconv.ParseFloat(strings.TrimSpace(payload.DistanceMeters), 64)
	elapsed, elapsedErr := strconv.Atoi(strings.TrimSpace(payload.ElapsedTimeInSeconds))
	name := strings.TrimSpace(payload.Name)
	now := time.Now()
	if len(match) != 2 || dateErr != nil || at.IsZero() || at.After(now.Add(5*time.Minute)) ||
		distanceErr != nil || math.IsNaN(distance) || distance < 0 || distance > 10_000_000 ||
		elapsedErr != nil || elapsed < 0 || elapsed > 7*24*60*60 || name == "" || len(name) > 1024 {
		http.Error(w, "Invalid run fields", http.StatusUnprocessableEntity)
		return
	}
	first := firstRunningWeek(now)
	if at.Before(first) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	run := runEvent{ID: match[1], Name: name, At: at.UTC(), DistanceMeters: distance, ElapsedSeconds: elapsed}
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

func runStatsError(w http.ResponseWriter, err error) {
	log.Printf("running stats: %v", err)
	http.Error(w, "Running stats are unavailable", http.StatusInternalServerError)
}
