import { A } from "@solidjs/router";
import { ArrowLeftIcon } from "./Icons";

interface BackLinkProps {
  href: string;
  label: string;
}

export function BackLink(props: BackLinkProps) {
  return (
    <A
      href={props.href}
      class="group inline-flex items-center gap-1.5 font-sans text-sm text-[var(--fg-muted)] transition-colors duration-150 hover:text-[var(--fg)]"
    >
      <ArrowLeftIcon class="h-3.5 w-3.5 transition-transform duration-150 group-hover:-translate-x-0.5" />
      <span>{props.label}</span>
    </A>
  );
}

interface NotFoundProps extends BackLinkProps {
  title: string;
  message: string;
}

export function NotFound(props: NotFoundProps) {
  return (
    <section class="py-4">
      <BackLink href={props.href} label={props.label} />
      <h1 class="mt-6 font-serif text-2xl font-bold tracking-[-0.02em] text-[var(--fg)]">
        {props.title}
      </h1>
      <p class="mt-2 font-sans text-[15px] text-[var(--fg-muted)]">
        {props.message}
      </p>
    </section>
  );
}
