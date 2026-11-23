import { For } from "solid-js";
import { GitHubIcon, LinkedInIcon, MailIcon } from "./Icons";

const ABOUT_PARAGRAPHS = [
  "Lorem ipsum dolor sit amet consectetur adipisicing elit. Pariatur possimus aliquip ut distinctio ducimus dolor illum deserunt deleniti.",
  "Laborum ipsum nihil animi vero quis id sed maxime et exercitation irure deserunt pariatur ullamco est assumenda atque veniam.",
  "Consequatur laboris nulla cumque est cillum similique mollit in quis tempor id illum dolor consequat in incididunt enim.",
];

export function AboutSection() {
  return (
    <section
      id="about"
      aria-labelledby="about-heading"
      class="rounded-2xl bg-[var(--surface)] p-6 sm:p-8"
    >
      {/* Top row: portrait + greeting & contacts */}
      <div class="mb-6 flex items-center gap-4 sm:gap-5">
        <figure class="h-[72px] w-[72px] shrink-0 overflow-hidden rounded-2xl border border-[var(--border)] bg-[var(--bg)] sm:h-[80px] sm:w-[80px]">
          <img
            src="/portrait.jpg"
            alt="Portrait of Tanat Chanwangsa"
            class="block h-full w-full object-cover"
            width={80}
            height={80}
          />
        </figure>

        <div class="min-w-0">
          <h1
            id="about-heading"
            class="font-sans text-lg font-semibold tracking-[-0.015em] text-[var(--fg)] sm:text-xl"
          >
            Hey, I'm Tanat
          </h1>

          <div
            class="-ml-1.5 mt-1.5 flex items-center gap-1"
            aria-label="Contact links"
          >
            <a
              href="https://github.com/Tchanwangsa"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex h-7 w-7 items-center justify-center rounded-md text-[var(--fg-muted)] transition-colors duration-150 hover:bg-[var(--surface-hover)] hover:text-[var(--fg)] focus-visible:bg-[var(--surface-hover)] focus-visible:text-[var(--fg)] focus-visible:outline-none"
              aria-label="GitHub"
              title="GitHub"
            >
              <GitHubIcon class="block h-[16px] w-[16px]" />
            </a>
            <a
              href="mailto:tchanwangsa@gmail.com"
              class="inline-flex h-7 w-7 items-center justify-center rounded-md text-[var(--fg-muted)] transition-colors duration-150 hover:bg-[var(--surface-hover)] hover:text-[var(--fg)] focus-visible:bg-[var(--surface-hover)] focus-visible:text-[var(--fg)] focus-visible:outline-none"
              aria-label="Email"
              title="Email"
            >
              <MailIcon class="block h-[16px] w-[16px]" />
            </a>
            <a
              href="https://www.linkedin.com/in/tanat-chanwangsa"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex h-7 w-7 items-center justify-center rounded-md text-[var(--fg-muted)] transition-colors duration-150 hover:bg-[var(--surface-hover)] hover:text-[var(--fg)] focus-visible:bg-[var(--surface-hover)] focus-visible:text-[var(--fg)] focus-visible:outline-none"
              aria-label="LinkedIn"
              title="LinkedIn"
            >
              <LinkedInIcon class="block h-[16px] w-[16px]" />
            </a>
          </div>
        </div>
      </div>

      {/* Body copy */}
      <div class="space-y-4 font-sans text-[15px] leading-[1.75] text-[var(--fg-body)] sm:text-[15.5px]">
        <For each={ABOUT_PARAGRAPHS}>
          {(paragraph) => <p class="text-pretty">{paragraph}</p>}
        </For>
      </div>
    </section>
  );
}
