package portfolioapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type githubSearchTransport func(*http.Request) (*http.Response, error)

func (f githubSearchTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGitHubPrivateCommitCountsKeepLatestPublic(t *testing.T) {
	for _, publicThisWeek := range []bool{true, false} {
		t.Run(fmt.Sprintf("publicThisWeek=%t", publicThisWeek), func(t *testing.T) {
			now := time.Now().UTC().Truncate(time.Second)
			publicDate := now.AddDate(0, 0, -1)
			if !publicThisWeek {
				publicDate = now.AddDate(0, 0, -10)
			}
			publicCalls, privateCalls := 0, 0
			github := &githubStats{
				username: "owner", token: "public", privateToken: "scoped",
				client: http.Client{Transport: githubSearchTransport(func(r *http.Request) (*http.Response, error) {
					query := r.URL.Query().Get("q")
					var body string
					switch {
					case strings.Contains(query, "is:private"):
						privateCalls++
						if r.Header.Get("Authorization") != "Bearer scoped" {
							t.Fatal("private search did not use the scoped token")
						}
						body = fmt.Sprintf(`{"total_count":1,"items":[{
							"html_url":"https://github.com/owner/secret-repo/commit/2",
							"repository":{"full_name":"owner/secret-repo","private":true},
							"commit":{"message":"secret commit message","author":{"date":%q}}
						}]}`, now.Format(time.RFC3339))
					case strings.Contains(query, "is:public"):
						publicCalls++
						if r.Header.Get("Authorization") != "Bearer public" {
							t.Fatal("public search did not use the public token")
						}
						body = fmt.Sprintf(`{"total_count":1,"items":[{
							"html_url":"https://github.com/owner/public-repo/commit/1",
							"repository":{"full_name":"owner/public-repo","private":false},
							"commit":{"message":"Latest public commit","author":{"date":%q}}
						}]}`, publicDate.Format(time.RFC3339))
						if !publicThisWeek && strings.Contains(query, "author-date:") {
							body = `{"total_count":0,"items":[]}`
						}
					default:
						t.Fatalf("commit search has no visibility constraint: %s", query)
					}
					return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
				})},
			}
			api := &statsAPI{github: github}
			response := httptest.NewRecorder()
			api.getGitHub(response, httptest.NewRequest(http.MethodGet, githubStatsPath, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("GET returned %d: %s", response.Code, response.Body)
			}
			if strings.Contains(response.Body.String(), "secret") {
				t.Fatal("private commit metadata reached the public API")
			}
			var snap githubSnapshot
			if err := json.Unmarshal(response.Body.Bytes(), &snap); err != nil {
				t.Fatal(err)
			}
			wantPrevious := 0
			wantPublicCalls := 2
			if publicThisWeek {
				wantPrevious, wantPublicCalls = 1, 1
			}
			if snap.DailyCommits[6].Commits != 1 || snap.DailyCommits[5].Commits != wantPrevious {
				t.Fatalf("unexpected daily totals: %+v", snap.DailyCommits)
			}
			if snap.LatestCommit == nil || snap.LatestCommit.Message != "Latest public commit" {
				t.Fatalf("unexpected latest commit: %+v", snap.LatestCommit)
			}
			if privateCalls != 1 || publicCalls != wantPublicCalls {
				t.Fatalf("made %d private and %d public queries", privateCalls, publicCalls)
			}
		})
	}
}
