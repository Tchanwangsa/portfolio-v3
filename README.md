# portfolio-v3

Source for [tchanwangsa.xyz](https://tchanwangsa.xyz), built with SolidJS, Tailwind CSS, and Vite. The stats cards read from [portfolio-api](https://github.com/Tchanwangsa/portfolio-api).

| Path | Contents |
| --- | --- |
| `src/content/about.md` | About copy |
| `src/data/` | Projects, blog posts, and the stats API client |
| `src/components/`, `src/pages/` | UI |

## Develop

```sh
npm install
npm run dev          # http://localhost:3000
npm run typecheck
```

The stats cards call `https://api.tchanwangsa.xyz`, in development too. To use a local API, run `VITE_API_URL=http://127.0.0.1:8787 npm run dev`.

If the API is unreachable, the Strava card shows saved sample data, the GitHub card shows as unavailable, and the agentic coding card stays hidden.

## Deploy

Pushing to `main` deploys to Vercel.
