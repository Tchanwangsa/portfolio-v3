export interface DailyCommitPoint {
  day: string;
  commits: number;
}

export interface GitHubStats {
  asOf: string;
  dailyCommits: DailyCommitPoint[];
  latestCommit: {
    repo: string;
    message: string;
    date: string;
    url: string;
  } | null;
}

export interface DailyTokenPoint {
  day: string;
  tokens: number; // raw token count
}

export interface AgenticCodingStats {
  asOf: string;
  dailyTokens: DailyTokenPoint[];
  mostUsedModel: string | null;
  mostUsedModelTokens?: number;
  totalTokens: number;
  sessions: number;
}

export interface StravaStats {
  weeklyKm: { label: string; km: number }[];
  latestRun: {
    name: string;
    distanceKm: number;
    duration: string;
    pace: string;
    when: string;
    url: string;
  } | null;
}

// Melbourne calendar weeks from the Strava export, through 24 November 2026.
export const STRAVA_STATS: StravaStats = {
  weeklyKm: [
    { label: "W-6", km: 16.9135 },
    { label: "W-5", km: 10.5584 },
    { label: "W-4", km: 6.2405 },
    { label: "W-3", km: 6.1703 },
    { label: "W-2", km: 0 },
    { label: "W-1", km: 20.7928 },
    { label: "Now", km: 0 },
  ],
  latestRun: {
    name: "2XU 10km",
    distanceKm: 10.18,
    duration: "55m 39s",
    pace: "5:28 /km",
    when: "Nov 22",
    url: "https://www.strava.com/activities/21076529003",
  },
};

const stravaStatsEndpoint = import.meta.env.VITE_STRAVA_STATS_API_URL ||
  (import.meta.env.DEV
    ? "http://127.0.0.1:8787/api/runs"
    : "https://api.tchanwangsa.xyz/api/runs");

const agenticStatsEndpoint = import.meta.env.VITE_AGENTIC_STATS_API_URL ||
  (import.meta.env.DEV
    ? "http://127.0.0.1:8787/api/agentic-stats"
    : "https://api.tchanwangsa.xyz/api/agentic-stats");

const githubStatsEndpoint = import.meta.env.VITE_GITHUB_STATS_API_URL ||
  (import.meta.env.DEV
    ? "http://127.0.0.1:8787/api/github-stats"
    : "https://api.tchanwangsa.xyz/api/github-stats");

async function fetchStats<T>(url: string, timeout = 4000): Promise<T | null> {
  try {
    const response = await fetch(url, { cache: "no-cache", signal: AbortSignal.timeout(timeout) });
    return response.ok ? await response.json() : null;
  } catch {
    return null;
  }
}

export async function loadStravaStats(): Promise<StravaStats> {
  return (await fetchStats<StravaStats>(stravaStatsEndpoint)) ?? STRAVA_STATS;
}

export async function loadAgenticCodingStats(): Promise<AgenticCodingStats | null> {
  const stats = await fetchStats<AgenticCodingStats>(agenticStatsEndpoint);
  if (stats || !import.meta.env.DEV) return stats;

  return fetchStats<AgenticCodingStats>(`${import.meta.env.BASE_URL}agentic-stats.json`);
}

export function loadGitHubStats(): Promise<GitHubStats | null> {
  return fetchStats<GitHubStats>(githubStatsEndpoint, 8000);
}

export function formatModelName(model: string): string {
  const name = model.replace(/^(anthropic|openai|google)\//i, "").toLowerCase();
  if (!/^(claude-|gpt-\d|gemini-\d)/.test(name)) return model;

  return name
    .replace(/-(?:\d{8}|\d{4}-\d{2}-\d{2})$/, "")
    .replace(/(\d)-(?=\d)/g, "$1.")
    .split("-")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ")
    .replace(/^Gpt /, "GPT-");
}

export function formatTokens(count: number): string {
  if (count >= 1_000_000_000) {
    return `${(count / 1_000_000_000).toFixed(1)}B`;
  }
  if (count >= 1_000_000) {
    return `${(count / 1_000_000).toFixed(1)}M`;
  }
  if (count >= 1_000) {
    return `${Math.round(count / 1_000)}k`;
  }
  return String(count);
}
