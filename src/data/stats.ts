export interface WeeklyRunPoint {
  label: string;
  km: number;
}

export interface StravaStats {
  weeklyKm: WeeklyRunPoint[];
  latestRun: {
    name: string;
    distanceKm: number;
    duration: string;
    pace: string;
    when: string;
    url: string;
  };
}

export interface DailyCommitPoint {
  day: string;
  commits: number;
}

export interface GitHubStats {
  dailyCommits: DailyCommitPoint[];
  latestProject: {
    repo: string;
    message: string;
    when: string;
    url: string;
  };
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

// Placeholder stats for the Strava and GitHub cards.
export const MOCK_STRAVA: StravaStats = {
  weeklyKm: [
    { label: "W-6", km: 18.4 },
    { label: "W-5", km: 24.2 },
    { label: "W-4", km: 14.8 },
    { label: "W-3", km: 29.5 },
    { label: "W-2", km: 22.1 },
    { label: "W-1", km: 31.6 },
    { label: "Now", km: 26.8 },
  ],
  latestRun: {
    name: "Evening River Loop",
    distanceKm: 7.4,
    duration: "36m 45s",
    pace: "4:58 /km",
    when: "2d ago",
    url: "https://www.strava.com",
  },
};

export const MOCK_GITHUB: GitHubStats = {
  dailyCommits: [
    { day: "Tue", commits: 4 },
    { day: "Wed", commits: 9 },
    { day: "Thu", commits: 3 },
    { day: "Fri", commits: 12 },
    { day: "Sat", commits: 7 },
    { day: "Sun", commits: 5 },
    { day: "Mon", commits: 8 },
  ],
  latestProject: {
    repo: "Tchanwangsa/portfolio-v3",
    message: "feat: add footer and stats section",
    when: "today",
    url: "https://github.com/Tchanwangsa/portfolio-v3",
  },
};

const agenticStatsEndpoint = import.meta.env.VITE_AGENTIC_STATS_API_URL ||
  (import.meta.env.DEV
    ? "http://127.0.0.1:8787/api/agentic-stats"
    : "https://api.tchawangsa.xyz/api/agentic-stats");

async function fetchAgenticStats(url: string): Promise<AgenticCodingStats | null> {
  try {
    const response = await fetch(url, { cache: "no-cache", signal: AbortSignal.timeout(4000) });
    return response.ok ? await response.json() : null;
  } catch {
    return null;
  }
}

export async function loadAgenticCodingStats(): Promise<AgenticCodingStats | null> {
  const stats = await fetchAgenticStats(agenticStatsEndpoint);
  if (stats || !import.meta.env.DEV) return stats;

  return fetchAgenticStats(`${import.meta.env.BASE_URL}agentic-stats.json`);
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
