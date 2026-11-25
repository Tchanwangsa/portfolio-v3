import { For } from "solid-js";
import aboutMarkdown from "../content/about.md?raw";
import { GitHubIcon, LinkedInIcon, MailIcon } from "./Icons";
import { Markdown } from "./Markdown";

const CONTACTS = [
  { label: "GitHub", href: "https://github.com/Tchanwangsa", icon: GitHubIcon },
  { label: "Email", href: "mailto:tchanwangsa@gmail.com", icon: MailIcon },
  { label: "LinkedIn", href: "https://www.linkedin.com/in/tanat-chanwangsa", icon: LinkedInIcon },
];

export function AboutSection() {
  return (
    <section
      id="about"
      aria-labelledby="about-heading"
      class="px-1 sm:px-2"
    >
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
            class="font-sans text-[24px] leading-tight font-semibold tracking-[-0.02em] text-[var(--fg)] sm:text-[32px]"
          >
            Hey, I'm Tanat
          </h1>

          <div
            class="-ml-[8px] mt-2 flex items-center gap-1"
            aria-label="Contact links"
          >
            <For each={CONTACTS}>
              {(contact) => {
                const external = contact.href.startsWith("https://");
                return (
                  <a
                    href={contact.href}
                    target={external ? "_blank" : undefined}
                    rel={external ? "noopener noreferrer" : undefined}
                    class="inline-flex h-9 w-9 items-center justify-center rounded-md text-[var(--fg-muted)] transition-colors duration-150 hover:bg-[var(--surface-hover)] hover:text-[var(--fg)] focus-visible:bg-[var(--surface-hover)] focus-visible:text-[var(--fg)] focus-visible:outline-none"
                    aria-label={contact.label}
                    title={contact.label}
                  >
                    <contact.icon class="block h-5 w-5 sm:h-6 sm:w-6" />
                  </a>
                );
              }}
            </For>
          </div>
        </div>
      </div>

      <Markdown
        content={aboutMarkdown}
        class="space-y-4 font-sans text-[15px] leading-[1.75] text-[var(--fg-body)] sm:text-[15.5px]"
      />
    </section>
  );
}
