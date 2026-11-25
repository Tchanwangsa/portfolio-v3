import { For } from "solid-js";
import { InlineMarkdown } from "./Markdown";

const CONTRIBUTIONS = [
  "Founded [Project Oculus](https://github.com/oculus-app/oculus), a macOS app that brings a student's Canvas, Ed and Echo360 coursework to the coding agent they already use.",
  "Replaced KaTeX and MathLive in Oculus with a [Rust maths engine](https://github.com/oculus-app/oculus/tree/master/app/math-core) compiled to WebAssembly, including a visual equation editor.",
  "Led and designed [Connect3](https://connect3.app/), a campus app for event discovery, AI search, club memberships and ticketing across four universities.",
  "Built scrapers that keep Connect3's club directory, membership products and prices up to date.",
  "Redesigned the homepage of the [DSCubed website](https://www.dscubed.org.au/) and moved the site to Next.js 16, React 19 and Tailwind 4.",
];

export function TechnicalContributions() {
  return (
    <section
      id="contributions"
      aria-labelledby="contributions-heading"
      class="px-1 sm:px-2"
    >
      <h2
        id="contributions-heading"
        class="mb-5 font-serif text-2xl font-bold tracking-[-0.02em] text-[var(--fg)]"
      >
        My work
      </h2>

      <ul class="space-y-1.5 font-sans text-[15px] leading-[1.75] text-[var(--fg-body)] sm:text-[15.5px]">
        <For each={CONTRIBUTIONS}>
          {(item) => (
            <li class="flex gap-3 text-pretty">
              <span
                class="select-none text-[var(--fg-muted)]"
                aria-hidden="true"
              >
                –
              </span>
              <span>
                <InlineMarkdown content={item} />
              </span>
            </li>
          )}
        </For>
      </ul>
    </section>
  );
}
