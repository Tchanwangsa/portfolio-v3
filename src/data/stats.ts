export interface WeeklyRunPoint {
  label: string;
  km: number;
}

export interface StravaStats {
  weeklyKm: WeeklyRunPoint[];
  latestRun: {
    name: string;
    distanceKm: number;
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
  dailyTokens: DailyTokenPoint[];
  mostUsedModel: string;
  totalTokens: number;
  sessions: number;
}

/**
 * Placeholder stats — swap these out with API fetches (Strava, GitHub,
 * and a future local-to-web gateway for agentic coding telemetry).
 */
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

export const MOCK_AGENTIC: AgenticCodingStats = {
  dailyTokens: [
    { day: "Tue", tokens: 2_450_000 },
    { day: "Wed", tokens: 4_120_000 },
    { day: "Thu", tokens: 1_890_000 },
    { day: "Fri", tokens: 5_640_000 },
    { day: "Sat", tokens: 3_280_000 },
    { day: "Sun", tokens: 2_910_000 },
    { day: "Mon", tokens: 4_530_000 },
  ],
  mostUsedModel: "claude-opus-4",
  totalTokens: 24_820_000,
  sessions: 38,
};

export function formatTokens(count: number): string {
  if (count >= 1_000_000) {
    return `${(count / 1_000_000).toFixed(1)}M`;
  }
  if (count >= 1_000) {
    return `${Math.round(count / 1_000)}k`;
  }
  return String(count);
}
