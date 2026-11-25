import { For } from "solid-js";

const CONTRIBUTIONS = [
  "Founded Project Oculus, a macOS app that brings a student's Canvas, Ed and Echo360 coursework to the coding agent they already use.",
  "Replaced KaTeX and MathLive in Oculus with a Rust maths engine compiled to WebAssembly, including a visual equation editor.",
  "Led and designed Connect3, a campus app for event discovery, AI search, club memberships and ticketing across four universities.",
  "Built scrapers that keep Connect3's club directory, membership products and prices up to date.",
  "Redesigned the DSCubed website's homepage and moved it to Next.js 16, React 19 and Tailwind 4.",
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
        What I've built
      </h2>

      <ul class="space-y-4 font-sans text-[15px] leading-[1.75] text-[var(--fg-body)] sm:text-[15.5px]">
        <For each={CONTRIBUTIONS}>
          {(item) => (
            <li class="flex gap-3 text-pretty">
              <span
                class="select-none text-[var(--fg-muted)]"
                aria-hidden="true"
              >
                –
              </span>
              <span>{item}</span>
            </li>
          )}
        </For>
      </ul>
    </section>
  );
}
