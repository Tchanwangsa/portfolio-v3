// Package stats defines the agentic coding snapshot that the collector
// publishes and the API serves.
package stats

import (
	"errors"
	"time"
)

// Path is the API route for reading and publishing the snapshot.
const Path = "/api/agentic-stats"

// maxCount is the largest integer a JavaScript client reads exactly.
const maxCount = 1<<53 - 1

type DayUsage struct {
	Day    string `json:"day"`
	Tokens int64  `json:"tokens"`
}

type Snapshot struct {
	AsOf                string     `json:"asOf"`
	DailyTokens         []DayUsage `json:"dailyTokens"`
	MostUsedModel       *string    `json:"mostUsedModel"`
	MostUsedModelTokens *int64     `json:"mostUsedModelTokens,omitempty"`
	TotalTokens         int64      `json:"totalTokens"`
	Sessions            int        `json:"sessions"`
}

func (s Snapshot) Validate() error {
	lastDay, err := time.Parse(time.DateOnly, s.AsOf)
	if err != nil {
		return errors.New("asOf must be a date in YYYY-MM-DD format")
	}
	if len(s.DailyTokens) != 7 {
		return errors.New("dailyTokens must contain seven days")
	}
	var total int64
	for i, day := range s.DailyTokens {
		if day.Day != lastDay.AddDate(0, 0, i-6).Format("Mon") || day.Tokens < 0 || day.Tokens > maxCount {
			return errors.New("dailyTokens must match the dates and contain valid counts")
		}
		total += day.Tokens
	}
	if s.TotalTokens != total || total > maxCount || s.Sessions < 0 || int64(s.Sessions) > maxCount {
		return errors.New("invalid weekly totals")
	}
	if s.MostUsedModel != nil && (*s.MostUsedModel == "" || len(*s.MostUsedModel) > 160) {
		return errors.New("invalid model name")
	}
	if s.MostUsedModelTokens != nil {
		tokens := *s.MostUsedModelTokens
		if s.MostUsedModel == nil || tokens <= 0 || tokens > total {
			return errors.New("invalid most-used model token count")
		}
	}
	return nil
}
