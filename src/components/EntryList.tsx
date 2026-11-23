import { For, Show } from "solid-js";
import { A } from "@solidjs/router";
import { formatShortDate, getYear } from "../lib/date";

export interface EntryListItem {
  href: string;
  title: string;
  date: string; // ISO YYYY-MM-DD
}

interface EntryListProps {
  items: EntryListItem[];
  ariaLabel: string;
  emptyMessage?: string;
}

export function EntryList(props: EntryListProps) {
  return (
    <Show
      when={props.items.length > 0}
      fallback={
        <p class="py-8 font-sans text-sm text-[var(--fg-muted)]">
          {props.emptyMessage ?? "Nothing here yet."}
        </p>
      }
    >
      <ul aria-label={props.ariaLabel} class="space-y-0.5">
        <For each={props.items}>
          {(item, index) => {
            const year = getYear(item.date);
            const showYear = () => {
              const i = index();
              return i === 0 || getYear(props.items[i - 1].date) !== year;
            };

            return (
              <li>
                <A
                  href={item.href}
                  class="group -mx-3 flex items-baseline gap-3 rounded-lg px-3 py-3 transition-colors duration-150 hover:bg-[var(--surface)] focus-visible:bg-[var(--surface)] focus-visible:outline-none sm:gap-5"
                >
                  <span
                    class="w-12 shrink-0 font-sans text-[13.5px] text-[var(--fg-muted)] tabular-nums sm:w-16"
                    aria-hidden={!showYear()}
                  >
                    {showYear() ? year : ""}
                  </span>

                  <span class="min-w-0 flex-1 font-sans text-[15px] font-medium leading-snug text-[var(--fg)] sm:text-[15.5px]">
                    {item.title}
                  </span>

                  <time
                    dateTime={item.date}
                    class="shrink-0 text-right font-sans text-[13.5px] text-[var(--fg-muted)] tabular-nums transition-colors duration-150 group-hover:text-[var(--fg)]"
                  >
                    {formatShortDate(item.date)}
                  </time>
                </A>
              </li>
            );
          }}
        </For>
      </ul>
    </Show>
  );
}
