import { createEffect, For } from "solid-js";
import { A, useLocation, type RouteSectionProps } from "@solidjs/router";
import { Logo } from "./components/Logo";
import { Footer } from "./components/Footer";

const NAV_ITEMS = [
  { label: "About", href: "/", end: true },
  { label: "Projects", href: "/projects", end: false },
  { label: "Blog", href: "/blogs", end: false },
] as const;

export default function App(props: RouteSectionProps) {
  const location = useLocation();

  createEffect(() => {
    void location.pathname;
    window.scrollTo({ top: 0, behavior: "auto" });
  });

  const isActive = (href: string, end: boolean) => {
    const pathname = location.pathname;
    if (end) {
      return pathname === href;
    }
    return pathname === href || pathname.startsWith(`${href}/`);
  };

  return (
    <div class="mx-auto flex min-h-dvh max-w-[680px] flex-col px-5 pt-8 pb-10 sm:px-6 sm:pt-12 sm:pb-12">
      <header class="mb-10 flex items-center justify-between gap-4 sm:mb-12">
        <A
          href="/"
          class="inline-flex items-center gap-2.5 text-[var(--fg)] transition-opacity duration-150 hover:opacity-75"
        >
          <Logo class="h-[24px] w-[24px] shrink-0" />
          <span class="font-sans text-xs font-bold leading-none tracking-[0.06em] uppercase sm:text-sm">
            Tanat Chanwangsa
          </span>
        </A>

        <nav class="flex items-center gap-4 sm:gap-6" aria-label="Primary">
          <For each={NAV_ITEMS}>
            {(item) => {
              const active = () => isActive(item.href, item.end);
              return (
                <A
                  href={item.href}
                  end={item.end}
                  aria-current={active() ? "page" : undefined}
                  class={`font-sans text-sm font-medium transition-colors duration-150 ${
                    active()
                      ? "text-[var(--fg)]"
                      : "text-[var(--fg-muted)] hover:text-[var(--fg)]"
                  }`}
                >
                  {item.label}
                </A>
              );
            }}
          </For>
        </nav>
      </header>

      <main class="flex-1">{props.children}</main>

      <Footer />
    </div>
  );
}
