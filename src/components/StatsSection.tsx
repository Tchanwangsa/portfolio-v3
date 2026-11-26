import {
  createMemo,
  createResource,
  createSignal,
  For,
  onCleanup,
  onMount,
  Show,
  type Component,
  type JSX,
} from "solid-js";
import {
  ArrowUpRightIcon,
  GitHubIcon,
  ModelIcon,
  RobotIcon,
  StravaIcon,
} from "./Icons";
import {
  formatModelName,
  formatTokens,
  loadAgenticCodingStats,
  loadGitHubStats,
  loadStravaStats,
  type StravaStats,
} from "../data/stats";
import { formatShortDate } from "../lib/date";

interface StatCardProps {
  icon: Component<{ class?: string }>;
  title: string;
  headline: string | null;
  loading: boolean;
  class?: string;
  children: JSX.Element;
}

function StatCard(props: StatCardProps) {
  return (
    <article class={`rounded-2xl bg-[var(--surface)] p-5 ${props.class ?? ""}`} aria-busy={props.loading}>
      <div class="mb-4 flex items-center justify-between gap-2">
        <div class="flex items-center gap-2 text-[var(--fg-muted)]">
          <props.icon class="h-3.5 w-3.5 shrink-0" />
          <span class="text-xs font-medium">{props.title}</span>
        </div>

        <div class="flex h-5 items-center text-right text-sm font-medium text-[var(--fg)] tabular-nums">
          <Show when={!props.loading} fallback={<Bone class="h-3 w-16" />}>
            {props.headline}
          </Show>
        </div>
      </div>

      <Show when={props.loading}>
        <p class="sr-only" role="status">Loading {props.title} stats…</p>
      </Show>
      {props.children}
    </article>
  );
}

function Unavailable(props: { title: string }) {
  return (
    <p class="text-sm text-[var(--fg-muted)]" role="status">
      {props.title} stats are unavailable.
    </p>
  );
}

// Skeletons reserve the same heights as the loaded content, so nothing below
// a card moves when its data arrives.

function Bone(props: { class: string }) {
  return (
    <span
      class={`block animate-pulse rounded bg-[var(--fg)]/10 motion-reduce:animate-none ${props.class}`}
      aria-hidden="true"
    />
  );
}

const SKELETON_BAR_HEIGHTS = [35, 60, 45, 80, 55, 70, 40];

function BarChartSkeleton(props: { class?: string; barsClass: string }) {
  return (
    <div class={props.class} aria-hidden="true">
      <div class={`flex items-end border-b border-[var(--border)] px-0.5 pb-1.5 ${props.barsClass}`}>
        <For each={SKELETON_BAR_HEIGHTS}>
          {(height) => (
            <div class="flex h-full flex-1 items-end">
              <div class="w-full" style={{ height: `${height}%` }}>
                <Bone class="h-full w-full rounded-b-none rounded-t-[4px]" />
              </div>
            </div>
          )}
        </For>
      </div>
      <div class="mt-1.5 h-[16.5px]" />
    </div>
  );
}

function LatestSkeleton() {
  return (
    <div class="mt-4 border-t border-[var(--border)] pt-2.5" aria-hidden="true">
      <div class="py-1.5">
        <div class="flex h-4 items-center justify-between">
          <Bone class="h-2.5 w-20" />
          <Bone class="h-2.5 w-10" />
        </div>
        <div class="mt-1 flex h-[20.25px] items-center">
          <Bone class="h-3 w-32" />
        </div>
        <div class="mt-0.5 flex h-4 items-center">
          <Bone class="h-2.5 w-44" />
        </div>
      </div>
    </div>
  );
}

interface Bar {
  label: string;
  value: number;
  description: string;
}

interface BarChartProps {
  bars: Bar[];
  active: number | null;
  onActive: (index: number | null) => void;
  ariaLabel: string;
  class?: string;
  barsClass: string;
  restingOpacity: string;
}

function BarChart(props: BarChartProps) {
  const max = () => Math.max(...props.bars.map((bar) => bar.value), 1);

  return (
    <div
      class={props.class}
      onMouseLeave={() => props.onActive(null)}
      role="group"
      aria-label={props.ariaLabel}
    >
      <div class={`flex items-end border-b border-[var(--border)] px-0.5 pb-1.5 ${props.barsClass}`}>
        <For each={props.bars}>
          {(bar, i) => {
            // Short bars keep a visible stub so a quiet day still reads as activity.
            const pct = () => (bar.value === 0 ? 0 : Math.max((bar.value / max()) * 100, 10));
            const opacity = () => {
              if (props.active === i()) return "1";
              if (props.active !== null) return "0.22";
              return i() === props.bars.length - 1 ? "0.85" : props.restingOpacity;
            };
            return (
              <button
                type="button"
                onMouseEnter={() => props.onActive(i())}
                onFocus={() => props.onActive(i())}
                onBlur={() => props.onActive(null)}
                class="flex h-full flex-1 cursor-pointer items-end focus:outline-none"
                aria-label={bar.description}
              >
                <span
                  class="block w-full rounded-t-[4px] bg-[var(--fg)] transition-all duration-150"
                  style={{ height: `${pct()}%`, opacity: opacity() }}
                />
              </button>
            );
          }}
        </For>
      </div>

      <div class="mt-1.5 flex items-center justify-between px-0.5 text-[11px] text-[var(--fg-muted)]">
        <For each={props.bars}>
          {(bar, i) => (
            <span
              class={`flex-1 text-center transition-colors duration-150 ${
                props.active === i() ? "text-[var(--fg)]" : ""
              }`}
            >
              {bar.label}
            </span>
          )}
        </For>
      </div>
    </div>
  );
}

interface LatestLinkProps {
  href: string;
  label: string;
  when: JSX.Element;
  title: string;
  children: JSX.Element;
}

function LatestLink(props: LatestLinkProps) {
  return (
    <a
      href={props.href}
      target="_blank"
      rel="noopener noreferrer"
      class="group -mx-2 block rounded-lg px-2 py-1.5 transition-colors duration-150 hover:bg-[var(--surface-hover)] focus-visible:bg-[var(--surface-hover)] focus-visible:outline-none"
    >
      <div class="flex items-center justify-between gap-2 text-xs text-[var(--fg-muted)]">
        <span>{props.label}</span>
        {props.when}
      </div>
      <div
        class="mt-1 flex items-center gap-1 text-[13.5px] font-medium text-[var(--fg)]"
        title={props.title}
      >
        <span class="truncate">{props.title}</span>
        <ArrowUpRightIcon class="h-3 w-3 shrink-0 text-[var(--fg-muted)] transition-transform duration-150 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 group-hover:text-[var(--fg)]" />
      </div>
      {props.children}
    </a>
  );
}

interface WeeklyRunChartProps {
  weeks: StravaStats["weeklyKm"];
  active: number | null;
  onActive: (index: number | null) => void;
}

function WeeklyRunChart(props: WeeklyRunChartProps) {
  // Draw in CSS pixels so the points line up with the HTML week labels at any
  // card width. A fixed viewBox would letterbox the plot on wide cards.
  let chartEl: HTMLDivElement | undefined;
  const [width, setWidth] = createSignal(260);
  onMount(() => {
    if (!chartEl) return;
    const observer = new ResizeObserver(([entry]) => {
      if (entry.contentRect.width > 0) setWidth(entry.contentRect.width);
    });
    observer.observe(chartEl);
    onCleanup(() => observer.disconnect());
  });

  const height = 84;
  const padX = 14;
  const padTop = 10;
  const padBottom = 8;
  const plotW = () => width() - padX * 2;
  const plotH = height - padTop - padBottom;
  const baseline = padTop + plotH;

  const coords = createMemo(() => {
    const maxKm = Math.max(...props.weeks.map((p) => p.km), 1);
    return props.weeks.map((p, i) => ({
      x: padX + (i / Math.max(props.weeks.length - 1, 1)) * plotW(),
      y: baseline - (p.km / (maxKm * 1.12)) * plotH,
      ...p,
    }));
  });

  const linePath = () => coords()
    .map((c, i) => `${i === 0 ? "M" : "L"} ${c.x.toFixed(1)} ${c.y.toFixed(1)}`)
    .join(" ");

  const areaPath = () => {
    const last = coords()[coords().length - 1];
    const first = coords()[0];
    return `${linePath()} L ${last.x.toFixed(1)} ${baseline.toFixed(1)} L ${first.x.toFixed(1)} ${baseline.toFixed(1)} Z`;
  };

  const guide = (y: number, opacity: string, dashed = true) => (
    <line
      x1={padX}
      y1={y}
      x2={width() - padX}
      y2={y}
      stroke="currentColor"
      stroke-opacity={opacity}
      stroke-dasharray={dashed ? "2 3" : undefined}
    />
  );

  return (
    <div
      ref={chartEl}
      class="relative flex flex-1 flex-col justify-end"
      onMouseLeave={() => props.onActive(null)}
      role="img"
      aria-label="Kilometers run over the last 7 weeks"
    >
      <svg
        viewBox={`0 0 ${width()} ${height}`}
        class="h-[84px] w-full overflow-visible"
      >
        <defs>
          <linearGradient id="strava-fill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="currentColor" stop-opacity="0.16" />
            <stop offset="100%" stop-color="currentColor" stop-opacity="0" />
          </linearGradient>
        </defs>

        {guide(padTop, "0.08")}
        {guide(padTop + plotH / 2, "0.08")}
        {guide(baseline, "0.12", false)}

        <path d={areaPath()} fill="url(#strava-fill)" class="text-[var(--fg)]" />
        <path
          d={linePath()}
          fill="none"
          stroke="currentColor"
          stroke-width="1.75"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="text-[var(--fg)]"
        />

        <For each={coords()}>
          {(c, i) => {
            const isActive = () => props.active === i();
            return (
              <g
                class="cursor-pointer"
                onMouseEnter={() => props.onActive(i())}
              >
                {/* Invisible hit target spanning the week's column */}
                <rect
                  x={c.x - plotW() / (props.weeks.length * 2)}
                  y={0}
                  width={plotW() / props.weeks.length}
                  height={height}
                  fill="transparent"
                />
                {isActive() && (
                  <line
                    x1={c.x}
                    y1={padTop}
                    x2={c.x}
                    y2={baseline}
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
                    props.active === null || isActive() ? 1 : 0.35
                  }
                />
              </g>
            );
          }}
        </For>
      </svg>

      <div class="relative mt-1.5 h-4 text-[11px] leading-4 text-[var(--fg-muted)]">
        <For each={coords()}>
          {(c, i) => (
            <span
              class={`absolute top-0 -translate-x-1/2 whitespace-nowrap transition-colors duration-150 ${
                props.active === i() ? "text-[var(--fg)]" : ""
              }`}
              style={{ left: `${c.x}px` }}
            >
              {c.label}
            </span>
          )}
        </For>
      </div>
    </div>
  );
}

function StravaCard() {
  const [activeIdx, setActiveIdx] = createSignal<number | null>(null);
  const [stats] = createResource(loadStravaStats);

  const headline = () => {
    const weeks = stats()?.weeklyKm;
    if (!weeks) return null;
    const idx = activeIdx();
    const km = idx === null ? weeks.reduce((sum, p) => sum + p.km, 0) : weeks[idx].km;
    return `${km.toFixed(1)} km`;
  };

  return (
    <StatCard icon={StravaIcon} title="Strava" headline={headline()} loading={stats.loading} class="flex flex-col">
      <Show
        when={!stats.loading}
        fallback={
          <>
            <div class="flex flex-1 flex-col justify-end" aria-hidden="true">
              <Bone class="h-[84px] w-full rounded-lg" />
              <div class="mt-1.5 h-4" />
            </div>
            <LatestSkeleton />
          </>
        }
      >
        <Show when={stats()} keyed fallback={<Unavailable title="Strava" />}>
          {(snapshot) => (
            <>
              <WeeklyRunChart weeks={snapshot.weeklyKm} active={activeIdx()} onActive={setActiveIdx} />

              <div class="mt-4 border-t border-[var(--border)] pt-2.5">
                <Show when={snapshot.latestRun} fallback={<p class="py-1.5 text-xs text-[var(--fg-muted)]">No runs in the last seven weeks.</p>} keyed>
                  {(run) => (
                    <LatestLink href={run.url} label="Latest run" when={<span>{run.when}</span>} title={run.name}>
                      <div class="mt-0.5 truncate text-xs text-[var(--fg-muted)] tabular-nums">
                        {run.distanceKm} km · {run.duration} ·{" "}
                        {run.pace}
                      </div>
                    </LatestLink>
                  )}
                </Show>
              </div>
            </>
          )}
        </Show>
      </Show>
    </StatCard>
  );
}

function GitHubCard() {
  const [activeIdx, setActiveIdx] = createSignal<number | null>(null);
  const [stats] = createResource(loadGitHubStats);

  const commits = (count: number) => `${count} ${count === 1 ? "commit" : "commits"}`;

  const headline = () => {
    const days = stats()?.dailyCommits;
    if (!days) return null;
    const idx = activeIdx();
    return commits(idx === null ? days.reduce((sum, day) => sum + day.commits, 0) : days[idx].commits);
  };

  return (
    <StatCard icon={GitHubIcon} title="GitHub" headline={headline()} loading={stats.loading} class="flex flex-col">
      <Show
        when={!stats.loading}
        fallback={
          <>
            <BarChartSkeleton class="flex flex-1 flex-col justify-end" barsClass="h-[84px] gap-2" />
            <LatestSkeleton />
          </>
        }
      >
        <Show when={stats()} keyed fallback={<Unavailable title="GitHub" />}>
          {(snapshot) => (
            <>
              <BarChart
                bars={snapshot.dailyCommits.map((day) => ({
                  label: day.day.slice(0, 1),
                  value: day.commits,
                  description: `${day.day}: ${commits(day.commits)}`,
                }))}
                active={activeIdx()}
                onActive={setActiveIdx}
                ariaLabel="GitHub commits over the last seven days"
                class="flex flex-1 flex-col justify-end"
                barsClass="h-[84px] gap-2"
                restingOpacity="0.42"
              />

              <div class="mt-4 border-t border-[var(--border)] pt-2.5">
                <Show
                  when={snapshot.latestCommit}
                  keyed
                  fallback={<p class="text-xs text-[var(--fg-muted)]">No public commits found.</p>}
                >
                  {(commit) => (
                    <LatestLink
                      href={commit.url}
                      label="Latest public commit"
                      when={<time dateTime={commit.date}>{formatShortDate(commit.date)}</time>}
                      title={commit.repo}
                    >
                      <div class="mt-0.5 truncate text-xs text-[var(--fg-muted)]" title={commit.message}>
                        {commit.message}
                      </div>
                    </LatestLink>
                  )}
                </Show>
              </div>
            </>
          )}
        </Show>
      </Show>
    </StatCard>
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

  // No snapshot means the collector is switched off, so the card steps aside.
  return (
    <Show when={stats.loading || stats()}>
      <StatCard icon={RobotIcon} title="Agentic Coding" headline={headline()} loading={stats.loading} class="sm:col-span-2">
        <Show
          when={stats()}
          keyed
          fallback={
            <>
              <BarChartSkeleton barsClass="h-[88px] gap-2.5 sm:gap-3" />
              <div class="mt-4 flex items-start justify-between gap-3 border-t border-[var(--border)] pt-3.5 sm:gap-4" aria-hidden="true">
                <div>
                  <div class="flex h-4 items-center"><Bone class="h-2.5 w-28" /></div>
                  <div class="mt-1 flex h-[20.25px] items-center"><Bone class="h-3 w-40" /></div>
                </div>
                <div class="flex gap-3 sm:gap-6">
                  <For each={["w-20", "w-10"]}>
                    {(width) => (
                      <div class="flex flex-col items-end">
                        <div class="flex h-4 items-center"><Bone class="h-2.5 w-10" /></div>
                        <div class="mt-1 flex h-[20.25px] items-center"><Bone class={`h-3 ${width}`} /></div>
                      </div>
                    )}
                  </For>
                </div>
              </div>
            </>
          }
        >
          {(snapshot) => (
            <>
              <BarChart
                bars={snapshot.dailyTokens.map((day) => ({
                  label: day.day,
                  value: day.tokens,
                  description: `${day.day}: ${formatTokens(day.tokens)} tokens`,
                }))}
                active={activeIdx()}
                onActive={setActiveIdx}
                ariaLabel={`Daily agentic coding tokens for the seven days through ${snapshot.asOf}`}
                barsClass="h-[88px] gap-2.5 sm:gap-3"
                restingOpacity="0.38"
              />

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
          )}
        </Show>
      </StatCard>
    </Show>
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

      <div class="mt-4 flex items-center justify-between gap-3 px-1 text-xs text-[var(--fg-muted)] sm:px-2">
        <span>Check out how I get my stats</span>
        <a
          href="https://github.com/Tchanwangsa/portfolio-api"
          target="_blank"
          rel="noopener noreferrer"
          class="group -mx-1.5 inline-flex shrink-0 items-center gap-1.5 rounded-md px-1.5 py-1 transition-colors duration-150 hover:text-[var(--fg)] focus-visible:text-[var(--fg)] focus-visible:outline-none"
        >
          <GitHubIcon class="h-3.5 w-3.5" />
          <span>Read the source</span>
          <ArrowUpRightIcon class="h-3 w-3 transition-transform duration-150 group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
        </a>
      </div>
    </section>
  );
}
