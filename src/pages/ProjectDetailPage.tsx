import { createEffect, For, Show } from "solid-js";
import { A, useParams } from "@solidjs/router";
import { ArrowLeftIcon, ArrowUpRightIcon } from "../components/Icons";
import { getProjectBySlug } from "../data/projects";
import { formatFullDate } from "../lib/date";

export function ProjectDetailPage() {
  const params = useParams<{ slug: string }>();
  const project = () => getProjectBySlug(params.slug);

  createEffect(() => {
    const current = project();
    document.title = current
      ? `${current.title} — Tanat Chanwangsa`
      : "Project Not Found — Tanat Chanwangsa";
  });

  return (
    <Show
      when={project()}
      fallback={
        <section class="py-4">
          <A
            href="/projects"
            class="group inline-flex items-center gap-1.5 font-sans text-sm text-[var(--fg-muted)] transition-colors duration-150 hover:text-[var(--fg)]"
          >
            <ArrowLeftIcon class="h-3.5 w-3.5 transition-transform duration-150 group-hover:-translate-x-0.5" />
            <span>Projects</span>
          </A>
          <h1 class="mt-6 font-serif text-2xl font-bold tracking-[-0.02em] text-[var(--fg)]">
            Project not found
          </h1>
          <p class="mt-2 font-sans text-[15px] text-[var(--fg-muted)]">
            The project you're looking for doesn't exist or has been moved.
          </p>
        </section>
      }
    >
      {(current) => (
        <article>
          <A
            href="/projects"
            class="group inline-flex items-center gap-1.5 font-sans text-sm text-[var(--fg-muted)] transition-colors duration-150 hover:text-[var(--fg)]"
          >
            <ArrowLeftIcon class="h-3.5 w-3.5 transition-transform duration-150 group-hover:-translate-x-0.5" />
            <span>Projects</span>
          </A>

          <header class="mt-6 border-b border-[var(--border)] pb-6">
            <h1 class="font-serif text-2xl font-bold tracking-[-0.02em] text-[var(--fg)] sm:text-[28px]">
              {current().title}
            </h1>

            <div class="mt-2.5 flex flex-wrap items-center justify-between gap-3">
              <time
                dateTime={current().date}
                class="font-sans text-sm text-[var(--fg-muted)] tabular-nums"
              >
                {formatFullDate(current().date)}
              </time>

              <Show when={current().repoUrl || current().liveUrl}>
                <div class="flex items-center gap-4 font-sans text-sm">
                  <Show when={current().liveUrl}>
                    {(url) => (
                      <a
                        href={url()}
                        target="_blank"
                        rel="noopener noreferrer"
                        class="group inline-flex items-center gap-1 text-[var(--fg-muted)] transition-colors duration-150 hover:text-[var(--fg)]"
                      >
                        <span>Visit</span>
                        <ArrowUpRightIcon class="h-3.5 w-3.5 transition-transform duration-150 group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
                      </a>
                    )}
                  </Show>
                  <Show when={current().repoUrl}>
                    {(url) => (
                      <a
                        href={url()}
                        target="_blank"
                        rel="noopener noreferrer"
                        class="group inline-flex items-center gap-1 text-[var(--fg-muted)] transition-colors duration-150 hover:text-[var(--fg)]"
                      >
                        <span>Source</span>
                        <ArrowUpRightIcon class="h-3.5 w-3.5 transition-transform duration-150 group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
                      </a>
                    )}
                  </Show>
                </div>
              </Show>
            </div>

            <Show when={current().stack && current().stack!.length > 0}>
              <ul
                aria-label="Technologies"
                class="mt-4 flex flex-wrap gap-1.5"
              >
                <For each={current().stack}>
                  {(tech) => (
                    <li class="rounded-md bg-[var(--surface)] px-2 py-0.5 font-mono text-xs text-[var(--fg-muted)]">
                      {tech}
                    </li>
                  )}
                </For>
              </ul>
            </Show>
          </header>

          <div class="mt-6 space-y-5 font-sans text-[15px] leading-[1.75] text-[var(--fg-body)] sm:text-[15.5px]">
            <p class="text-pretty text-[var(--fg)]">{current().summary}</p>
            <For each={current().body}>
              {(paragraph) => <p class="text-pretty">{paragraph}</p>}
            </For>
          </div>
        </article>
      )}
    </Show>
  );
}
