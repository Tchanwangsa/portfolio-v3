# Portfolio stats API

Tokscale produces a rolling seven-day snapshot on the Mac. The Go collector saves a local copy and sends the same snapshot to the API. The API stores the latest snapshot on disk and serves it to the portfolio.

## Development

Run commands from the repo root with Go 1.22 or newer. Copy `.env.dev.example` to `.env.dev` inside `server/` and set `STATS_PUBLISH_TOKEN` to a random value of at least 32 characters. The local server and collector use the same file.

```sh
go run ./server -env-file server/.env.dev
go run ./scripts -env-file server/.env.dev
npm run dev
```

Run the server continuously and run the collector once to seed the API. The dev site reads `http://127.0.0.1:8787/api/agentic-stats` and can fall back to the local snapshot. `GET /healthz` returns `204` while the API is up.

## Deployment

1. Build from the repo root with `go build -o portfolio-stats-api ./server`. Run the binary behind HTTPS at `api.tchanwangsa.xyz` and mount persistent storage for `STATS_DATA_FILE`. Configure it with the settings in `.env.deploy.example`. The server also accepts a host-provided `PORT` when `STATS_LISTEN_ADDR` is unset.
2. Store a dedicated production `STATS_PUBLISH_TOKEN` in the hosting service and in the Mac collector's private configuration. Set the collector's `STATS_PUBLISH_URL` to `https://api.tchanwangsa.xyz/api/agentic-stats`.
3. Build the portfolio for `tchanwangsa.xyz`. It defaults to the API URL above. To override the URL, set `VITE_AGENTIC_STATS_API_URL` at build time. Allow the portfolio's exact origin in `STATS_ALLOWED_ORIGINS` on the API.

The API accepts a snapshot only when `PUT /api/agentic-stats` includes the publish token as a Bearer credential. `GET /api/agentic-stats` is public and returns the daily counts, total, top model and its token count, session count, and snapshot date. Keep the publish token in server and collector settings.

### Deploy on Vercel

The Vercel project serves the portfolio and the Go API from the same deployment.
Connect a private Blob store to Production for the agentic coding snapshot. Set
`BLOB_STORE_ID` to the store ID, use the connected `BLOB_READ_WRITE_TOKEN`, and
add the stats publish and GitHub credentials in the Vercel project settings.

## GitHub stats

Set `GITHUB_USERNAME` to your GitHub login and `GITHUB_TOKEN` to a
fine-grained personal access token with **Public repositories** selected and no
added permissions. Use `server/.env.dev` locally. In production, set both values
on the Go service using its environment or secret settings. The token belongs on
the Go server. Vite only needs the API URL.

The card reads `GET /api/github-stats`. Locally it defaults to
`http://127.0.0.1:8787/api/github-stats`. In production it defaults to
`https://api.tchanwangsa.xyz/api/github-stats`. Set `VITE_GITHUB_STATS_API_URL`
at build time to use a different endpoint. Add the portfolio's exact origin to
`STATS_ALLOWED_ORIGINS` when hosting the API separately.

The endpoint exposes seven days of public default-branch commits grouped by UTC
author date, the latest public commit GitHub Search finds, and the total stars
earned across public repositories owned by the configured account. The
server caches a successful response for 30 minutes and reuses it if GitHub
temporarily fails. When credentials are absent or the first fetch fails the
card shows an unavailable status.

## Running stats

Set a dedicated `RUNS_WEBHOOK_TOKEN` in production to enable the running API.
IFTTT sends authenticated POST requests to `/api/runs` with the run name, activity
type, timestamp, distance in metres, elapsed seconds, and Strava activity URL.
The public GET response supplies the weekly totals and latest run shown on the
portfolio. Weeks start Monday in Melbourne. The API discards deliveries outside
the seven-week window and deduplicates retries by activity ID.
