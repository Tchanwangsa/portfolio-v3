package portfolioapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const githubStatsPath = "/api/github-stats"

var githubUsername = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{0,38}$`)

type dailyCommitCount struct {
	Day     string `json:"day"`
	Commits int    `json:"commits"`
}

type publicCommit struct {
	Repo    string `json:"repo"`
	Message string `json:"message"`
	Date    string `json:"date"`
	URL     string `json:"url"`
}

type githubSnapshot struct {
	AsOf         string              `json:"asOf"`
	DailyCommits [7]dailyCommitCount `json:"dailyCommits"`
	LatestCommit *publicCommit       `json:"latestCommit"`
}

type githubStats struct {
	username string
	token    string
	client   http.Client

	mu          sync.Mutex
	snapshot    *githubSnapshot
	nextRefresh time.Time
}

func (api *statsAPI) getGitHub(w http.ResponseWriter, r *http.Request) {
	api.setReadHeaders(w, r)
	if api.github == nil {
		http.Error(w, "GitHub stats are unavailable", http.StatusServiceUnavailable)
		return
	}
	snapshot := api.github.get(r.Context())
	if snapshot == nil {
		http.Error(w, "GitHub stats are unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snapshot)
}

func (g *githubStats) get(ctx context.Context) *githubSnapshot {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UTC()
	if now.Before(g.nextRefresh) {
		return g.snapshot
	}

	ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	snapshot, err := g.fetch(ctx, now)
	if err != nil {
		log.Printf("refresh GitHub stats: %v", err)
		g.nextRefresh = now.Add(time.Minute)
		return g.snapshot
	}

	g.snapshot = snapshot
	g.nextRefresh = now.Add(30 * time.Minute)
	return snapshot
}

func authorDateRange(first, last time.Time) string {
	return " author-date:" + first.Format(time.DateOnly) + "T00:00:00Z.." + last.Format(time.DateOnly) + "T23:59:59Z"
}

func (g *githubStats) fetch(ctx context.Context, now time.Time) (*githubSnapshot, error) {
	snapshot := &githubSnapshot{AsOf: now.Format(time.DateOnly)}
	first := now.AddDate(0, 0, 1-len(snapshot.DailyCommits))
	dayIndex := make(map[string]int, len(snapshot.DailyCommits))
	for i := range snapshot.DailyCommits {
		day := first.AddDate(0, 0, i)
		dayIndex[day.Format(time.DateOnly)] = i
		snapshot.DailyCommits[i].Day = day.Format("Mon")
	}

	const pageSize = 100
	filter := authorDateRange(first, now)
	total, commits, err := g.search(ctx, filter, 1, pageSize)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		_, latest, err := g.search(ctx, "", 1, 1)
		if err != nil {
			return nil, err
		}
		if len(latest) != 0 {
			snapshot.LatestCommit = &latest[0]
		}
		return snapshot, nil
	}
	snapshot.LatestCommit = &commits[0]

	// GitHub Search only makes the first 1,000 results available for pagination.
	if total > 1000 {
		for i := range snapshot.DailyCommits {
			day := first.AddDate(0, 0, i)
			count, _, err := g.search(ctx, authorDateRange(day, day), 1, 1)
			if err != nil {
				return nil, err
			}
			snapshot.DailyCommits[i].Commits = count
		}
		return snapshot, nil
	}

	for page := 1; ; page++ {
		for _, commit := range commits {
			i, ok := dayIndex[commit.Date]
			if !ok {
				return nil, errors.New("GitHub commit search returned a date outside the window")
			}
			snapshot.DailyCommits[i].Commits++
		}
		if page*pageSize >= total {
			return snapshot, nil
		}
		var pageTotal int
		pageTotal, commits, err = g.search(ctx, filter, page+1, pageSize)
		if err != nil {
			return nil, err
		}
		if pageTotal != total {
			return nil, errors.New("GitHub commit search changed during pagination")
		}
	}
}

func (g *githubStats) search(ctx context.Context, dateFilter string, page, pageSize int) (int, []publicCommit, error) {
	query := url.Values{
		"q":        {"author:" + g.username + " is:public" + dateFilter},
		"sort":     {"author-date"},
		"order":    {"desc"},
		"page":     {strconv.Itoa(page)},
		"per_page": {strconv.Itoa(pageSize)},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/search/commits?"+query.Encode(), nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")

	res, err := g.client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf("GitHub commit search returned HTTP %d", res.StatusCode)
	}

	var result struct {
		TotalCount int  `json:"total_count"`
		Incomplete bool `json:"incomplete_results"`
		Items      []struct {
			URL        string `json:"html_url"`
			Repository struct {
				Name    string `json:"full_name"`
				Private *bool  `json:"private"`
			} `json:"repository"`
			Commit struct {
				Message string `json:"message"`
				Author  struct {
					Date time.Time `json:"date"`
				} `json:"author"`
			} `json:"commit"`
		} `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return 0, nil, errors.New("invalid GitHub commit search response")
	}
	expected := min(pageSize, max(0, result.TotalCount-(page-1)*pageSize))
	if result.TotalCount < 0 || result.Incomplete || len(result.Items) != expected {
		return 0, nil, errors.New("incomplete GitHub commit search results")
	}

	commits := make([]publicCommit, 0, len(result.Items))
	for _, item := range result.Items {
		if item.Repository.Private == nil || *item.Repository.Private {
			return 0, nil, errors.New("GitHub commit search returned a non-public repository")
		}
		commitURL, err := url.Parse(item.URL)
		if err != nil || commitURL.Scheme != "https" || commitURL.Host != "github.com" ||
			!strings.HasPrefix(commitURL.Path, "/"+item.Repository.Name+"/commit/") || item.Commit.Author.Date.IsZero() {
			return 0, nil, errors.New("GitHub commit search returned invalid commit metadata")
		}
		message, _, _ := strings.Cut(item.Commit.Message, "\n")
		commits = append(commits, publicCommit{
			Repo:    item.Repository.Name,
			Message: strings.TrimSpace(message),
			Date:    item.Commit.Author.Date.UTC().Format(time.DateOnly),
			URL:     item.URL,
		})
	}
	return result.TotalCount, commits, nil
}
