import { createEffect, For, Show } from "solid-js";
import { A, useParams } from "@solidjs/router";
import { ArrowLeftIcon } from "../components/Icons";
import { getBlogPostBySlug } from "../data/blogs";
import { formatFullDate } from "../lib/date";

export function BlogDetailPage() {
  const params = useParams<{ slug: string }>();
  const post = () => getBlogPostBySlug(params.slug);

  createEffect(() => {
    const current = post();
    document.title = current
      ? `${current.title} — Tanat Chanwangsa`
      : "Post Not Found — Tanat Chanwangsa";
  });

  return (
    <Show
      when={post()}
      fallback={
        <section class="py-4">
          <A
            href="/blogs"
            class="group inline-flex items-center gap-1.5 font-sans text-sm text-[var(--fg-muted)] transition-colors duration-150 hover:text-[var(--fg)]"
          >
            <ArrowLeftIcon class="h-3.5 w-3.5 transition-transform duration-150 group-hover:-translate-x-0.5" />
            <span>Blog</span>
          </A>
          <h1 class="mt-6 font-serif text-2xl font-bold tracking-[-0.02em] text-[var(--fg)]">
            Post not found
          </h1>
          <p class="mt-2 font-sans text-[15px] text-[var(--fg-muted)]">
            The post you're looking for doesn't exist or has been moved.
          </p>
        </section>
      }
    >
      {(current) => (
        <article>
          <A
            href="/blogs"
            class="group inline-flex items-center gap-1.5 font-sans text-sm text-[var(--fg-muted)] transition-colors duration-150 hover:text-[var(--fg)]"
          >
            <ArrowLeftIcon class="h-3.5 w-3.5 transition-transform duration-150 group-hover:-translate-x-0.5" />
            <span>Blog</span>
          </A>

          <header class="mt-6 border-b border-[var(--border)] pb-6">
            <h1 class="font-serif text-2xl font-bold tracking-[-0.02em] text-[var(--fg)] sm:text-[28px]">
              {current().title}
            </h1>
            <time
              dateTime={current().date}
              class="mt-2.5 block font-sans text-sm text-[var(--fg-muted)] tabular-nums"
            >
              {formatFullDate(current().date)}
            </time>
          </header>

          <div class="mt-6 space-y-5 font-sans text-[15px] leading-[1.75] text-[var(--fg-body)] sm:text-[15.5px]">
            <For each={current().body}>
              {(paragraph) => <p class="text-pretty">{paragraph}</p>}
            </For>
          </div>
        </article>
      )}
    </Show>
  );
}
