package stats

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const Path = "/api/agentic-stats"

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

// Write replaces the previous snapshot atomically, including when the feed is a symlink.
func (s Snapshot) Write(output string) error {
	if target, err := filepath.EvalSymlinks(output); err == nil {
		output = target
	} else if !os.IsNotExist(err) {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(output), ".agentic-stats-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encodeErr := encoder.Encode(s)
	closeErr := file.Close()
	if encodeErr != nil {
		return encodeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(file.Name(), output)
}
