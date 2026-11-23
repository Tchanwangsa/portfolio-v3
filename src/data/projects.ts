import { compareDatesDesc } from "../lib/date";

export interface Project {
  slug: string;
  title: string;
  date: string; // ISO YYYY-MM-DD
  summary: string;
  stack?: string[];
  repoUrl?: string;
  liveUrl?: string;
  body: string[];
}

const RAW_PROJECTS: Project[] = [
  {
    slug: "portfolio-v3",
    title: "Portfolio v3",
    date: "2026-11-18",
    summary:
      "A minimal personal site built with SolidJS and Tailwind CSS, focused on clean typography, fast navigation, and live personal telemetry.",
    stack: ["SolidJS", "TypeScript", "Tailwind CSS", "Vite"],
    repoUrl: "https://github.com/Tchanwangsa/portfolio-v3",
    body: [
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.",
      "Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
    ],
  },
  {
    slug: "agent-telemetry-gateway",
    title: "Agent Telemetry Gateway",
    date: "2026-08-14",
    summary:
      "Lightweight local-to-web collector for tracking daily token usage, session counts, and model distribution across coding agents.",
    stack: ["TypeScript", "SQLite", "Cloudflare Workers"],
    body: [
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Quisque vel justo nec turpis cursus pretium. Integer in mi quis neque pretium aliquet.",
      "Vivamus sit amet facilisis nunc. Curabitur feugiat, erat nec dignissim efficitur, justo risus eleifend elit, ac laoreet arcu ligula ac odio.",
    ],
  },
  {
    slug: "runlog-sync",
    title: "Runlog Sync",
    date: "2026-03-22",
    summary:
      "Small background service that mirrors Strava training logs into a static JSON feed for embedding personal mileage graphs.",
    stack: ["Go", "Strava API"],
    body: [
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas.",
      "Donec vitae libero non enim placerat eleifend aliquam erat volutpat. Curabitur diam ex, dapibus purus sapien, cursus sed nisl in, commodo gravida.",
    ],
  },
  {
    slug: "fast-diff-viewer",
    title: "Fast Diff Viewer",
    date: "2025-10-05",
    summary:
      "Keyboard-first side-by-side diff explorer designed to stay responsive on multi-thousand-line patches.",
    stack: ["Rust", "WebAssembly", "SolidJS"],
    body: [
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Praesent vulputate sem vel libero vehicula, in porta velit suscipit.",
      "Aliquam erat volutpat. Nullam efficitur mauris at mauris dapibus, non malesuada ex rhoncus.",
    ],
  },
  {
    slug: "tiny-kv",
    title: "Tiny KV",
    date: "2025-04-11",
    summary:
      "An embeddable log-structured key-value store written from scratch to explore compaction, write-ahead logs, and crash recovery.",
    stack: ["Rust", "Systems"],
    body: [
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi tincidunt, erat a dignissim malesuada, sapien eros cursus orci, eget interdum erat sem eu lorem.",
      "Suspendisse potenti. Ut ac dui vel leo malesuada sodales non ac dolor.",
    ],
  },
  {
    slug: "campus-transit-map",
    title: "Campus Transit Map",
    date: "2024-09-19",
    summary:
      "Real-time shuttle arrival tracker with live stop predictions and low-bandwidth mobile support.",
    stack: ["TypeScript", "MapLibre", "Node.js"],
    body: [
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae.",
      "Etiam pulvinar, lacus sed porta vulputate, est sem tincidunt ipsum, eget dictum lacus sapien et odio.",
    ],
  },
];

export const PROJECTS: Project[] = [...RAW_PROJECTS].sort((a, b) =>
  compareDatesDesc(a.date, b.date)
);

export function getProjectBySlug(slug: string): Project | undefined {
  return PROJECTS.find((project) => project.slug === slug);
}
