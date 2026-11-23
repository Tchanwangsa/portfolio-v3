import { createResource, createSignal, For, Show } from "solid-js";
import {
  ArrowUpRightIcon,
  GitHubIcon,
  ModelIcon,
  RobotIcon,
  StravaIcon,
} from "./Icons";
import {
  MOCK_GITHUB,
  MOCK_STRAVA,
  formatModelName,
  formatTokens,
  loadAgenticCodingStats,
} from "../data/stats";

function StravaCard() {
  const [activeIdx, setActiveIdx] = createSignal<number | null>(null);
  const points = MOCK_STRAVA.weeklyKm;
  const totalKm = points.reduce((sum, p) => sum + p.km, 0);
  const maxKm = Math.max(...points.map((p) => p.km), 1);

  const width = 260;
  const height = 84;
  const padX = 10;
  const padTop = 10;
  const padBottom = 8;
  const plotW = width - padX * 2;
  const plotH = height - padTop - padBottom;

  const coords = points.map((p, i) => ({
    x: padX + (i / Math.max(points.length - 1, 1)) * plotW,
    y: padTop + plotH - (p.km / (maxKm * 1.12)) * plotH,
    ...p,
  }));

  const linePath = coords
    .map((c, i) => `${i === 0 ? "M" : "L"} ${c.x.toFixed(1)} ${c.y.toFixed(1)}`)
    .join(" ");

  const areaPath = `${linePath} L ${coords[coords.length - 1].x.toFixed(1)} ${(
    padTop + plotH
  ).toFixed(1)} L ${coords[0].x.toFixed(1)} ${(padTop + plotH).toFixed(1)} Z`;

  const headline = () => {
    const idx = activeIdx();
    if (idx !== null) {
      const p = points[idx];
      return `${p.km.toFixed(1)} km`;
    }
    return `${totalKm.toFixed(1)} km`;
  };

  return (
    <article class="flex flex-col rounded-2xl bg-[var(--surface)] p-5">
      <div class="mb-4 flex items-center justify-between gap-2">
        <div class="flex items-center gap-2 text-[var(--fg-muted)]">
          <StravaIcon class="h-3.5 w-3.5 shrink-0" />
          <span class="text-xs font-medium">Strava</span>
        </div>

        <div class="text-right text-sm font-medium text-[var(--fg)] tabular-nums">
          {headline()}
        </div>
      </div>

      {/* 7-week line/area graph */}
      <div
        class="relative flex flex-1 flex-col justify-end"
        onMouseLeave={() => setActiveIdx(null)}
        role="img"
        aria-label="Kilometers run over the last 7 weeks"
      >
        <svg
          viewBox={`0 0 ${width} ${height}`}
          class="h-[84px] w-full overflow-visible"
        >
          <defs>
            <linearGradient id="strava-fill" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="currentColor" stop-opacity="0.16" />
              <stop offset="100%" stop-color="currentColor" stop-opacity="0" />
            </linearGradient>
          </defs>

          {/* Subtle horizontal guide lines */}
          <line
            x1={padX}
            y1={padTop}
            x2={width - padX}
            y2={padTop}
            stroke="currentColor"
            stroke-opacity="0.08"
            stroke-dasharray="2 3"
          />
          <line
            x1={padX}
            y1={padTop + plotH / 2}
            x2={width - padX}
            y2={padTop + plotH / 2}
            stroke="currentColor"
            stroke-opacity="0.08"
            stroke-dasharray="2 3"
          />
          <line
            x1={padX}
            y1={padTop + plotH}
            x2={width - padX}
            y2={padTop + plotH}
            stroke="currentColor"
            stroke-opacity="0.12"
          />

          <path d={areaPath} fill="url(#strava-fill)" class="text-[var(--fg)]" />
          <path
            d={linePath}
            fill="none"
            stroke="currentColor"
            stroke-width="1.75"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="text-[var(--fg)]"
          />

          <For each={coords}>
            {(c, i) => {
              const isActive = () => activeIdx() === i();
              return (
                <g
                  class="cursor-pointer"
                  onMouseEnter={() => setActiveIdx(i())}
                >
                  {/* Invisible wider hit target */}
                  <rect
                    x={c.x - plotW / (points.length * 2)}
                    y={0}
                    width={plotW / points.length}
                    height={height}
                    fill="transparent"
                  />
                  {isActive() && (
                    <line
                      x1={c.x}
                      y1={padTop}
                      x2={c.x}
                      y2={padTop + plotH}
                      stroke="currentColor"
                      stroke-opacity="0.22"
                      stroke-width="1"
                    />
                  )}
                  <circle
                    cx={c.x}
                    cy={c.y}
                    r={isActive() ? 3.5 : 2.25}
                    class="fill-[var(--fg)] transition-all duration-150"
                    opacity={
                      activeIdx() === null || isActive() ? 1 : 0.35
                    }
                  />
                </g>
              );
            }}
          </For>
        </svg>

        <div class="mt-1.5 flex items-center justify-between px-1 text-[11px] text-[var(--fg-muted)]">
          <For each={points}>
            {(p, i) => (
              <span
                class={`transition-colors duration-150 ${
                  activeIdx() === i() ? "text-[var(--fg)]" : ""
                }`}
              >
                {p.label}
              </span>
            )}
          </For>
        </div>
      </div>

      {/* Latest run */}
      <div class="mt-4 border-t border-[var(--border)] pt-2.5">
        <a
          href={MOCK_STRAVA.latestRun.url}
          target="_blank"
          rel="noopener noreferrer"
          class="group -mx-2 block rounded-lg px-2 py-1.5 transition-colors duration-150 hover:bg-[var(--surface-hover)] focus-visible:bg-[var(--surface-hover)] focus-visible:outline-none"
        >
          <div class="flex items-center justify-between gap-2 text-xs text-[var(--fg-muted)]">
            <span>Latest run</span>
            <span>{MOCK_STRAVA.latestRun.when}</span>
          </div>
          <div
            class="mt-1 flex items-center gap-1 text-[13.5px] font-medium text-[var(--fg)]"
            title={MOCK_STRAVA.latestRun.name}
          >
            <span class="truncate">{MOCK_STRAVA.latestRun.name}</span>
            <ArrowUpRightIcon class="h-3 w-3 shrink-0 text-[var(--fg-muted)] transition-transform duration-150 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 group-hover:text-[var(--fg)]" />
          </div>
          <div class="mt-0.5 truncate text-xs text-[var(--fg-muted)] tabular-nums">
            {MOCK_STRAVA.latestRun.distanceKm} km · {MOCK_STRAVA.latestRun.duration} ·{" "}
            {MOCK_STRAVA.latestRun.pace}
          </div>
        </a>
      </div>
    </article>
  );
}

function GitHubCard() {
  const [activeIdx, setActiveIdx] = createSignal<number | null>(null);
  const days = MOCK_GITHUB.dailyCommits;
  const totalCommits = days.reduce((sum, d) => sum + d.commits, 0);
  const maxCommits = Math.max(...days.map((d) => d.commits), 1);

  const headline = () => {
    const idx = activeIdx();
    if (idx !== null) {
      const count = days[idx].commits;
      return `${count} ${count === 1 ? "commit" : "commits"}`;
    }
    return `${totalCommits} commits`;
  };

  return (
    <article class="flex flex-col rounded-2xl bg-[var(--surface)] p-5">
      <div class="mb-4 flex items-center justify-between gap-2">
        <div class="flex items-center gap-2 text-[var(--fg-muted)]">
          <GitHubIcon class="h-3.5 w-3.5 shrink-0" />
          <span class="text-xs font-medium">GitHub</span>
        </div>

        <div class="text-right text-sm font-medium text-[var(--fg)] tabular-nums">
          {headline()}
        </div>
      </div>

      {/* 7-day commit bar chart */}
      <div
        class="flex flex-1 flex-col justify-end"
        onMouseLeave={() => setActiveIdx(null)}
        role="img"
        aria-label="GitHub commits over the last 7 days"
      >
        <div class="flex h-[84px] items-end gap-2 border-b border-[var(--border)] px-0.5 pb-1.5">
          <For each={days}>
            {(d, i) => {
              const pct = Math.max((d.commits / maxCommits) * 100, 10);
              const isActive = () => activeIdx() === i();
              const isDimmed = () =>
                activeIdx() !== null && activeIdx() !== i();
              return (
                <button
                  type="button"
                  onMouseEnter={() => setActiveIdx(i())}
                  onFocus={() => setActiveIdx(i())}
                  onBlur={() => setActiveIdx(null)}
                  class="flex h-full flex-1 cursor-pointer items-end focus:outline-none"
                  aria-label={`${d.day}: ${d.commits} commits`}
                >
                  <span
                    class="block w-full rounded-t-[4px] bg-[var(--fg)] transition-all duration-150"
                    style={{
                      height: `${pct}%`,
                      opacity: isActive()
                        ? "1"
                        : isDimmed()
                        ? "0.22"
                        : i() === days.length - 1
                        ? "0.85"
                        : "0.42",
                    }}
                  />
                </button>
              );
            }}
          </For>
        </div>

        <div class="mt-1.5 flex items-center justify-between px-0.5 text-[11px] text-[var(--fg-muted)]">
          <For each={days}>
            {(d, i) => (
              <span
                class={`flex-1 text-center transition-colors duration-150 ${
                  activeIdx() === i() ? "text-[var(--fg)]" : ""
                }`}
              >
                {d.day.slice(0, 1)}
              </span>
            )}
          </For>
        </div>
      </div>

      {/* Latest commit */}
      <div class="mt-4 border-t border-[var(--border)] pt-2.5">
        <a
          href={MOCK_GITHUB.latestProject.url}
          target="_blank"
          rel="noopener noreferrer"
          class="group -mx-2 block rounded-lg px-2 py-1.5 transition-colors duration-150 hover:bg-[var(--surface-hover)] focus-visible:bg-[var(--surface-hover)] focus-visible:outline-none"
        >
          <div class="flex items-center justify-between gap-2 text-xs text-[var(--fg-muted)]">
            <span>Latest commit</span>
            <span>{MOCK_GITHUB.latestProject.when}</span>
          </div>
          <div
            class="mt-1 flex items-center gap-1 text-[13.5px] font-medium text-[var(--fg)]"
            title={MOCK_GITHUB.latestProject.repo}
          >
            <span class="truncate">{MOCK_GITHUB.latestProject.repo}</span>
            <ArrowUpRightIcon class="h-3 w-3 shrink-0 text-[var(--fg-muted)] transition-transform duration-150 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 group-hover:text-[var(--fg)]" />
          </div>
          <div
            class="mt-0.5 truncate text-xs text-[var(--fg-muted)]"
            title={MOCK_GITHUB.latestProject.message}
          >
            {MOCK_GITHUB.latestProject.message}
          </div>
        </a>
      </div>
    </article>
  );
}

function AgenticCodingCard() {
  const [activeIdx, setActiveIdx] = createSignal<number | null>(null);
  const [stats] = createResource(loadAgenticCodingStats);

  const headline = () => {
    const snapshot = stats();
    if (!snapshot) return null;

    const idx = activeIdx();
    const tokens = idx === null ? snapshot.totalTokens : snapshot.dailyTokens[idx].tokens;
    return `${formatTokens(tokens)} tokens`;
  };

  return (
    <article class="rounded-2xl bg-[var(--surface)] p-5 sm:col-span-2">
      <div class="mb-4 flex items-center justify-between gap-2">
        <div class="flex items-center gap-2 text-[var(--fg-muted)]">
          <RobotIcon class="h-3.5 w-3.5 shrink-0" />
          <span class="text-xs font-medium">Agentic Coding</span>
        </div>

        <div class="text-right text-sm font-medium text-[var(--fg)] tabular-nums">
          {headline()}
        </div>
      </div>

      <Show
        when={stats()}
        keyed
        fallback={
          <p class="text-sm text-[var(--fg-muted)]" role="status">
            {stats.loading ? "Loading usage stats…" : "Usage stats are unavailable."}
          </p>
        }
      >
        {(snapshot) => {
          const days = snapshot.dailyTokens;
          const maxTokens = Math.max(...days.map((day) => day.tokens), 1);
          return (
            <>
              <div
                onMouseLeave={() => setActiveIdx(null)}
                role="group"
                aria-label={`Daily agentic coding tokens for the seven days through ${snapshot.asOf}`}
              >
                <div class="flex h-[88px] items-end gap-2.5 border-b border-[var(--border)] px-0.5 pb-1.5 sm:gap-3">
                  <For each={days}>
                    {(d, i) => {
                      const pct = d.tokens === 0 ? 0 : Math.max((d.tokens / maxTokens) * 100, 10);
                      const isActive = () => activeIdx() === i();
                      const isDimmed = () =>
                        activeIdx() !== null && activeIdx() !== i();
                      return (
                        <button
                          type="button"
                          onMouseEnter={() => setActiveIdx(i())}
                          onFocus={() => setActiveIdx(i())}
                          onBlur={() => setActiveIdx(null)}
                          class="flex h-full flex-1 cursor-pointer items-end focus:outline-none"
                          aria-label={`${d.day}: ${formatTokens(d.tokens)} tokens`}
                        >
                          <span
                            class="block w-full rounded-t-[4px] bg-[var(--fg)] transition-all duration-150"
                            style={{
                              height: `${pct}%`,
                              opacity: isActive()
                                ? "1"
                                : isDimmed()
                                ? "0.22"
                                : i() === days.length - 1
                                ? "0.85"
                                : "0.38",
                            }}
                          />
                        </button>
                      );
                    }}
                  </For>
                </div>

                <div class="mt-1.5 flex items-center justify-between px-0.5 text-[11px] text-[var(--fg-muted)]">
                  <For each={days}>
                    {(d, i) => (
                      <span
                        class={`flex-1 text-center transition-colors duration-150 ${
                          activeIdx() === i() ? "text-[var(--fg)]" : ""
                        }`}
                      >
                        {d.day}
                      </span>
                    )}
                  </For>
                </div>
              </div>

              <div class="mt-4 flex items-start justify-between gap-3 border-t border-[var(--border)] pt-3.5 sm:gap-4">
                <div class="min-w-0">
                  <div class="text-xs text-[var(--fg-muted)]">Most used model (7d)</div>
                  <div class="mt-1 flex flex-wrap items-center gap-x-1.5 gap-y-0.5 text-[13.5px] font-medium text-[var(--fg)]">
                    <Show when={snapshot.mostUsedModel} fallback="—" keyed>
                      {(model) => {
                        const name = formatModelName(model);
                        return (
                          <>
                            <span class="inline-flex min-w-0 items-center gap-1.5">
                              <ModelIcon model={model} class="h-3.5 w-3.5 shrink-0 text-[var(--fg)]" />
                              <span class="min-w-0 [overflow-wrap:anywhere]" title={name}>{name}</span>
                            </span>
                            <Show when={snapshot.mostUsedModelTokens} keyed>
                              {(tokens) => (
                                <span class="text-xs tabular-nums" title={`${tokens.toLocaleString()} tokens`}>
                                  ({formatTokens(tokens)} tokens)
                                </span>
                              )}
                            </Show>
                          </>
                        );
                      }}
                    </Show>
                  </div>
                </div>

                <dl class="flex shrink-0 gap-3 text-right sm:gap-6" aria-label="Totals across all tracked agents over 7 days">
                  <div>
                    <dt class="text-xs text-[var(--fg-muted)]">Tokens</dt>
                    <dd class="mt-1 text-[13.5px] font-medium text-[var(--fg)] tabular-nums">
                      {snapshot.totalTokens.toLocaleString()}
                    </dd>
                  </div>
                  <div>
                    <dt class="text-xs text-[var(--fg-muted)]">Sessions</dt>
                    <dd class="mt-1 text-[13.5px] font-medium text-[var(--fg)] tabular-nums">
                      {snapshot.sessions.toLocaleString()}
                    </dd>
                  </div>
                </dl>
              </div>
            </>
          );
        }}
      </Show>
    </article>
  );
}

export function StatsSection() {
  return (
    <section id="stats" aria-labelledby="stats-heading">
      <h2
        id="stats-heading"
        class="mb-5 px-1 font-serif text-2xl font-bold tracking-[-0.02em] text-[var(--fg)] sm:px-2"
      >
        Stats
      </h2>

      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <StravaCard />
        <GitHubCard />
        <AgenticCodingCard />
      </div>
    </section>
  );
}
