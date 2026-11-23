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

1. Build from the repo root with `go build -o portfolio-stats-api ./server`. Run the binary behind HTTPS at `api.tchawangsa.xyz` and mount persistent storage for `STATS_DATA_FILE`. Configure it with the settings in `.env.deploy.example`. The server also accepts a host-provided `PORT` when `STATS_LISTEN_ADDR` is unset.
2. Store a dedicated production `STATS_PUBLISH_TOKEN` in the hosting service and in the Mac collector's private configuration. Set the collector's `STATS_PUBLISH_URL` to `https://api.tchawangsa.xyz/api/agentic-stats`.
3. Build the portfolio for `tchawangsa.xyz`. It defaults to the API URL above. To override the URL, set `VITE_AGENTIC_STATS_API_URL` at build time. Allow the portfolio's exact origin in `STATS_ALLOWED_ORIGINS` on the API.

The API accepts a snapshot only when `PUT /api/agentic-stats` includes the publish token as a Bearer credential. `GET /api/agentic-stats` is public and returns the daily counts, total, top model and its token count, session count, and snapshot date. Keep the publish token in server and collector settings.

## GitHub stats

Set `GITHUB_USERNAME` to your GitHub login and `GITHUB_TOKEN` to a
fine-grained personal access token with **Public repositories** selected and no
added permissions. Use `server/.env.dev` locally. In production, set both values
on the Go service using its environment or secret settings. The token belongs on
the Go server. Vite only needs the API URL.

The card reads `GET /api/github-stats`. Locally it defaults to
`http://127.0.0.1:8787/api/github-stats`. In production it defaults to
`https://api.tchawangsa.xyz/api/github-stats`. Set `VITE_GITHUB_STATS_API_URL`
at build time to use a different endpoint. Add the portfolio's exact origin to
`STATS_ALLOWED_ORIGINS` when hosting the API separately.

The endpoint exposes seven days of public default-branch commits grouped by UTC
author date and a link to the latest public commit GitHub Search finds. The
server caches a successful response for 30 minutes and reuses it if GitHub
temporarily fails. When credentials are absent or the first fetch fails the
card shows an unavailable status.
