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
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const githubStatsPath = "/api/github-stats"

var githubUsername = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{0,38}$`)

// Browsers group commits into their own local days, and every local day of the
// last week started within the past seven days.
const commitWindow = 7 * 24 * time.Hour

// Commit times are rounded down to 15 minutes, which keeps every time zone's day
// boundaries exact without publishing when private work happened to the second.
const commitTimeStep = 15 * 60

type publicCommit struct {
	Repo    string    `json:"repo"`
	Message string    `json:"message"`
	At      time.Time `json:"at"`
	URL     string    `json:"url"`
}

type githubSnapshot struct {
	AsOf         time.Time     `json:"asOf"`
	CommitTimes  []int64       `json:"commitTimes"`
	LatestCommit *publicCommit `json:"latestCommit"`
}

type githubStats struct {
	username     string
	token        string
	privateToken string
	client       http.Client

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

	now := time.Now().UTC().Truncate(time.Second)
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
	if g.privateToken != "" {
		private, _, err := g.commitTimes(ctx, g.privateToken, "private", now.Add(-commitWindow), now)
		if err != nil {
			log.Printf("refresh private GitHub commit times: %v", err)
			g.nextRefresh = now.Add(time.Minute)
			if g.snapshot == nil {
				slices.Sort(snapshot.CommitTimes)
				g.snapshot = snapshot
			}
			return g.snapshot
		}
		snapshot.CommitTimes = append(snapshot.CommitTimes, private...)
	}
	slices.Sort(snapshot.CommitTimes)
	g.snapshot = snapshot
	g.nextRefresh = now.Add(30 * time.Minute)
	return snapshot
}

func authorDateRange(first, last time.Time) string {
	const layout = "2006-01-02T15:04:05Z"
	return " author-date:" + first.UTC().Format(layout) + ".." + last.UTC().Format(layout)
}

func (g *githubStats) fetch(ctx context.Context, now time.Time) (*githubSnapshot, error) {
	from := now.Add(-commitWindow)
	snapshot := &githubSnapshot{AsOf: now}
	times, latest, err := g.commitTimes(ctx, g.token, "public", from, now)
	if err != nil {
		return nil, err
	}
	if latest == nil {
		_, items, err := searchCommits(ctx, g, g.token, "public", "", 1, 1)
		if err != nil {
			return nil, err
		}
		if len(items) != 0 {
			if latest, err = items[0].public(); err != nil {
				return nil, err
			}
		}
	}
	snapshot.LatestCommit = latest
	snapshot.CommitTimes = times
	return snapshot, nil
}

// commitTimes lists the author times of commits between from and to, inclusive,
// along with the newest one when searching public commits. GitHub Search only
// pages through the first 1,000 results, so busier windows are split in half.
func (g *githubStats) commitTimes(ctx context.Context, token, visibility string, from, to time.Time) ([]int64, *publicCommit, error) {
	const pageSize = 100
	filter := authorDateRange(from, to)
	total, items, err := searchCommits(ctx, g, token, visibility, filter, 1, pageSize)
	if err != nil {
		return nil, nil, err
	}
	if total > 1000 {
		if to.Sub(from) < time.Hour {
			return nil, nil, errors.New("too many commits to count")
		}
		mid := from.Add(to.Sub(from) / 2).Truncate(time.Second)
		later, latest, err := g.commitTimes(ctx, token, visibility, mid.Add(time.Second), to)
		if err != nil {
			return nil, nil, err
		}
		earlier, earlierLatest, err := g.commitTimes(ctx, token, visibility, from, mid)
		if err != nil {
			return nil, nil, err
		}
		if latest == nil {
			latest = earlierLatest
		}
		return append(earlier, later...), latest, nil
	}

	times := make([]int64, 0, total)
	var latest *publicCommit
	for page := 1; ; page++ {
		for _, item := range items {
			at := item.Commit.Author.Date
			if at.Before(from) || at.After(to) {
				return nil, nil, errors.New("GitHub commit search returned a date outside the window")
			}
			if visibility == "public" {
				commit, err := item.public()
				if err != nil {
					return nil, nil, err
				}
				if latest == nil {
					latest = commit
				}
			}
			unix := at.Unix()
			times = append(times, unix-((unix%commitTimeStep)+commitTimeStep)%commitTimeStep)
		}
		if page*pageSize >= total {
			return times, latest, nil
		}
		var pageTotal int
		pageTotal, items, err = searchCommits(ctx, g, token, visibility, filter, page+1, pageSize)
		if err != nil {
			return nil, nil, err
		}
		if pageTotal != total {
			return nil, nil, errors.New("GitHub commit search changed during pagination")
		}
	}
}

type commitSearchItem struct {
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
}

// public checks that a search result really is a public commit before any of
// its details reach the public API.
func (item commitSearchItem) public() (*publicCommit, error) {
	if item.Repository.Private == nil || *item.Repository.Private {
		return nil, errors.New("GitHub commit search returned a non-public repository")
	}
	commitURL, err := url.Parse(item.URL)
	if err != nil || commitURL.Scheme != "https" || commitURL.Host != "github.com" ||
		!strings.HasPrefix(commitURL.Path, "/"+item.Repository.Name+"/commit/") || item.Commit.Author.Date.IsZero() {
		return nil, errors.New("GitHub commit search returned invalid commit metadata")
	}
	message, _, _ := strings.Cut(item.Commit.Message, "\n")
	return &publicCommit{
		Repo:    item.Repository.Name,
		Message: strings.TrimSpace(message),
		At:      item.Commit.Author.Date.UTC(),
		URL:     item.URL,
	}, nil
}

func searchCommits(ctx context.Context, g *githubStats, token, visibility, dateFilter string, page, pageSize int) (int, []commitSearchItem, error) {
	query := url.Values{
		"q":        {"author:" + g.username + " is:" + visibility + dateFilter},
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
	req.Header.Set("Authorization", "Bearer "+token)
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
		TotalCount int                `json:"total_count"`
		Incomplete bool               `json:"incomplete_results"`
		Items      []commitSearchItem `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return 0, nil, errors.New("invalid GitHub commit search response")
	}
	expected := min(pageSize, max(0, result.TotalCount-(page-1)*pageSize))
	if result.TotalCount < 0 || result.Incomplete || len(result.Items) != expected {
		return 0, nil, errors.New("incomplete GitHub commit search results")
	}
	return result.TotalCount, result.Items, nil
}
