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
      <div class="mb-6 flex items-center gap-4 sm:gap-[30px]">
        <figure class="h-[100px] w-[100px] shrink-0 overflow-hidden rounded-2xl border border-[var(--border)] bg-[var(--bg)] sm:h-[132px] sm:w-[132px]">
          <img
            src="/portrait.jpg"
            alt="Portrait of Tanat Chanwangsa"
            class="block h-full w-full origin-[50%_35%] scale-[1.22] object-cover"
            width={132}
            height={132}
          />
        </figure>

        <div class="min-w-0">
          <h1
            id="about-heading"
            class="font-sans text-[22px] leading-tight font-semibold tracking-[-0.02em] text-[var(--fg)] sm:text-[30px]"
          >
            Hey, I'm Tanat
          </h1>

          <div
            class="-ml-[9px] mt-2 flex items-center gap-1.5"
            aria-label="Contact links"
          >
            <a
              href="https://github.com/Tchanwangsa"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex h-[42px] w-[42px] items-center justify-center rounded-md text-[var(--fg-muted)] transition-colors duration-150 hover:bg-[var(--surface-hover)] hover:text-[var(--fg)] focus-visible:bg-[var(--surface-hover)] focus-visible:text-[var(--fg)] focus-visible:outline-none"
              aria-label="GitHub"
              title="GitHub"
            >
              <GitHubIcon class="block h-6 w-6" />
            </a>
            <a
              href="mailto:tchanwangsa@gmail.com"
              class="inline-flex h-[42px] w-[42px] items-center justify-center rounded-md text-[var(--fg-muted)] transition-colors duration-150 hover:bg-[var(--surface-hover)] hover:text-[var(--fg)] focus-visible:bg-[var(--surface-hover)] focus-visible:text-[var(--fg)] focus-visible:outline-none"
              aria-label="Email"
              title="Email"
            >
              <MailIcon class="block h-6 w-6" />
            </a>
            <a
              href="https://www.linkedin.com/in/tanat-chanwangsa"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex h-[42px] w-[42px] items-center justify-center rounded-md text-[var(--fg-muted)] transition-colors duration-150 hover:bg-[var(--surface-hover)] hover:text-[var(--fg)] focus-visible:bg-[var(--surface-hover)] focus-visible:text-[var(--fg)] focus-visible:outline-none"
              aria-label="LinkedIn"
              title="LinkedIn"
            >
              <LinkedInIcon class="block h-6 w-6" />
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
