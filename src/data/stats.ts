import { formatShortDate } from "../lib/date";

export interface DailyCommitPoint {
  day: string;
  commits: number;
}

export interface GitHubStats {
  dailyCommits: DailyCommitPoint[];
  latestCommit: {
    repo: string;
    message: string;
    date: string; // local YYYY-MM-DD
    url: string;
  } | null;
}

interface GitHubStatsResponse {
  asOf: string;
  commitTimes: number[]; // Unix seconds
  latestCommit: {
    repo: string;
    message: string;
    at: string;
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

interface StravaStatsResponse {
  runs: { at: string; km: number }[];
  latestRun: {
    name: string;
    distanceKm: number;
    duration: string;
    pace: string;
    at: string;
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

const WEEK_COUNT = 7;
const DAY_COUNT = 7;

function startOfLocalDay(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}

function addLocalDays(date: Date, days: number): Date {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate() + days);
}

function localIsoDate(date: Date): string {
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}

/** Totals runs into the viewer's local Monday-to-Sunday weeks. */
export function summarizeRuns(response: StravaStatsResponse, now = new Date()): StravaStats {
  const today = startOfLocalDay(now);
  const thisWeek = addLocalDays(today, -((today.getDay() + 6) % 7));
  const starts = Array.from({ length: WEEK_COUNT + 1 }, (_, i) => addLocalDays(thisWeek, 7 * (i - WEEK_COUNT + 1)));
  const weeklyKm = starts.slice(0, WEEK_COUNT).map((_, i) => ({
    label: i === WEEK_COUNT - 1 ? "Now" : `W-${WEEK_COUNT - 1 - i}`,
    km: 0,
  }));
  for (const run of response.runs) {
    const at = new Date(run.at);
    const week = starts.findIndex((start, i) => i < WEEK_COUNT && at >= start && at < starts[i + 1]);
    if (week !== -1 && at <= now) weeklyKm[week].km += run.km;
  }
  const latest = response.latestRun;
  return {
    weeklyKm,
    latestRun: latest && {
      name: latest.name,
      distanceKm: latest.distanceKm,
      duration: latest.duration,
      pace: latest.pace,
      when: formatShortDate(localIsoDate(new Date(latest.at))),
      url: latest.url,
    },
  };
}

/** Counts commits into the viewer's last seven local days, ending today. */
export function summarizeCommits(response: GitHubStatsResponse, now = new Date()): GitHubStats {
  const today = startOfLocalDay(now);
  const days = Array.from({ length: DAY_COUNT }, (_, i) => addLocalDays(today, i - DAY_COUNT + 1));
  const dailyCommits = days.map((day) => ({
    day: day.toLocaleDateString("en-US", { weekday: "short" }),
    commits: 0,
  }));
  const end = addLocalDays(today, 1);
  for (const seconds of response.commitTimes) {
    const at = new Date(seconds * 1000);
    if (at < days[0] || at >= end) continue;
    let i = DAY_COUNT - 1;
    while (at < days[i]) i--;
    dailyCommits[i].commits++;
  }
  const latest = response.latestCommit;
  return {
    dailyCommits,
    latestCommit: latest && {
      repo: latest.repo,
      message: latest.message,
      date: localIsoDate(new Date(latest.at)),
      url: latest.url,
    },
  };
}

export async function loadStravaStats(): Promise<StravaStats> {
  const response = await fetchStats<StravaStatsResponse>(stravaStatsEndpoint);
  return response ? summarizeRuns(response) : STRAVA_STATS;
}

export async function loadAgenticCodingStats(): Promise<AgenticCodingStats | null> {
  const stats = await fetchStats<AgenticCodingStats>(agenticStatsEndpoint);
  if (stats || !import.meta.env.DEV) return stats;

  return fetchStats<AgenticCodingStats>(`${import.meta.env.BASE_URL}agentic-stats.json`);
}

export async function loadGitHubStats(): Promise<GitHubStats | null> {
  const response = await fetchStats<GitHubStatsResponse>(githubStatsEndpoint, 8000);
  return response && summarizeCommits(response);
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
