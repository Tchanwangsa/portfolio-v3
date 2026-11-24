package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSummarize(t *testing.T) {
	var report graph
	err := json.Unmarshal([]byte(`{
		"contributions": [
			{"date": "2026-11-18", "totals": {"tokens": 100}, "clients": [
				{"modelId": "claude-opus-5-5", "tokens": {"input": 60, "output": 20}},
				{"modelId": "gpt-5", "tokens": {"cacheRead": 20}}
			]},
			{"date": "2026-11-24", "totals": {"tokens": 50}, "clients": [
				{"modelId": "gpt-5", "tokens": {"input": 30, "reasoning": 20}}
			]},
			{"date": "2026-11-17", "totals": {"tokens": 999}}
		],
		"timeMetrics": {"sessionCount": 4}
	}`), &report)
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := summarize(report, time.Date(2026, time.November, 24, 9, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.AsOf != "2026-11-24" || snapshot.TotalTokens != 150 || snapshot.Sessions != 4 {
		t.Fatalf("unexpected totals: %+v", snapshot)
	}
	first, last := snapshot.DailyTokens[0], snapshot.DailyTokens[6]
	if first.Day != "Wed" || first.Tokens != 100 || last.Day != "Tue" || last.Tokens != 50 {
		t.Fatalf("unexpected days: %+v", snapshot.DailyTokens)
	}
	if snapshot.MostUsedModel == nil || *snapshot.MostUsedModel != "claude-opus-5-5" || *snapshot.MostUsedModelTokens != 80 {
		t.Fatalf("unexpected top model: %v", snapshot.MostUsedModel)
	}
}

func TestSummarizeRequiresSessionsWhenTokensExist(t *testing.T) {
	var report graph
	if err := json.Unmarshal([]byte(`{"contributions": [{"date": "2026-11-24", "totals": {"tokens": 1}}]}`), &report); err != nil {
		t.Fatal(err)
	}
	if _, err := summarize(report, time.Date(2026, time.November, 24, 9, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("expected an error for a graph without session metrics")
	}
}

func TestCheckTarget(t *testing.T) {
	token := "0123456789abcdef0123456789abcdef"
	for endpoint, ok := range map[string]bool{
		"https://api.example.com/api/agentic-stats": true,
		"http://127.0.0.1:8787/api/agentic-stats":   true,
		"http://localhost:8787/api/agentic-stats":   true,
		"http://api.example.com/api/agentic-stats":  false,
		"https://user@example.com/api":              false,
		"not a url":                                 false,
	} {
		if err := checkTarget(endpoint, token); (err == nil) != ok {
			t.Errorf("checkTarget(%q) = %v, want ok=%t", endpoint, err, ok)
		}
	}
	if err := checkTarget("https://api.example.com", "short"); err == nil {
		t.Error("expected a short token to be rejected")
	}
}
