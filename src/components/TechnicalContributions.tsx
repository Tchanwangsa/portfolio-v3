import { For } from "solid-js";

const CONTRIBUTIONS = [
  "Lorem ipsum dolor sit amet consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
  "Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
  "Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.",
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
        Technical Contributions
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
