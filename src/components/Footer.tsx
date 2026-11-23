import { UpArrowIcon } from "./Icons";

export function Footer() {
  const scrollToTop = () => {
    const prefersReducedMotion = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;
    window.scrollTo({
      top: 0,
      behavior: prefersReducedMotion ? "auto" : "smooth",
    });
  };

  return (
    <footer class="mt-16 flex flex-wrap items-center justify-between gap-4 border-t border-[var(--border)] pt-6 font-mono text-xs text-[var(--fg-muted)]">
      <p>© 2026 Tanat Chanwangsa.</p>

      <button
        type="button"
        onClick={scrollToTop}
        class="group inline-flex cursor-pointer items-center gap-1.5 text-[var(--fg-muted)] transition-colors duration-150 hover:text-[var(--fg)] focus-visible:text-[var(--fg)] focus-visible:outline-none"
      >
        <UpArrowIcon class="h-3.5 w-3.5 shrink-0 transition-transform duration-150 group-hover:-translate-y-0.5" />
        <span>Back to Top</span>
      </button>
    </footer>
  );
}
