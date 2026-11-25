import { For, createMemo } from "solid-js";

type InlineNode =
  | { type: "text"; value: string }
  | { type: "link"; label: string; href: string }
  | { type: "strong"; value: string }
  | { type: "em"; value: string }
  | { type: "code"; value: string };

const INLINE_PATTERN =
  /\[([^\]]+)\]\(([^)\s]+)\)|\*\*([^*]+)\*\*|\*([^*]+)\*|`([^`]+)`/g;

function isSafeHref(href: string): boolean {
  return (
    href.startsWith("https://") ||
    href.startsWith("http://") ||
    href.startsWith("mailto:") ||
    href.startsWith("/") ||
    href.startsWith("#")
  );
}

function isExternalHref(href: string): boolean {
  return href.startsWith("https://") || href.startsWith("http://");
}

function parseInline(text: string): InlineNode[] {
  const nodes: InlineNode[] = [];
  let lastIndex = 0;

  for (const match of text.matchAll(INLINE_PATTERN)) {
    const index = match.index ?? 0;
    if (index > lastIndex) {
      nodes.push({ type: "text", value: text.slice(lastIndex, index) });
    }

    if (match[1] !== undefined && match[2] !== undefined) {
      if (isSafeHref(match[2])) {
        nodes.push({ type: "link", label: match[1], href: match[2] });
      } else {
        nodes.push({ type: "text", value: match[1] });
      }
    } else if (match[3] !== undefined) {
      nodes.push({ type: "strong", value: match[3] });
    } else if (match[4] !== undefined) {
      nodes.push({ type: "em", value: match[4] });
    } else if (match[5] !== undefined) {
      nodes.push({ type: "code", value: match[5] });
    }

    lastIndex = index + match[0].length;
  }

  if (lastIndex < text.length) {
    nodes.push({ type: "text", value: text.slice(lastIndex) });
  }

  return nodes;
}

function parseParagraphs(markdown: string): InlineNode[][] {
  return markdown
    .replace(/\r\n/g, "\n")
    .trim()
    .split(/\n\s*\n+/)
    .map((block) => block.replace(/\n+/g, " ").trim())
    .filter(Boolean)
    .map(parseInline);
}

function Inline(props: { nodes: InlineNode[] }) {
  return (
    <For each={props.nodes}>
      {(node) => {
        switch (node.type) {
          case "link": {
            const external = isExternalHref(node.href);
            return (
              <a
                href={node.href}
                target={external ? "_blank" : undefined}
                rel={external ? "noopener noreferrer" : undefined}
                class="text-[var(--fg)] underline decoration-[var(--fg-muted)]/50 underline-offset-[3px] transition-colors duration-150 hover:decoration-[var(--fg)] focus-visible:decoration-[var(--fg)] focus-visible:outline-none"
              >
                {node.label}
              </a>
            );
          }
          case "strong":
            return (
              <strong class="font-medium text-[var(--fg)]">
                {node.value}
              </strong>
            );
          case "em":
            return <em>{node.value}</em>;
          case "code":
            return (
              <code class="rounded bg-[var(--bg)] px-1.5 py-0.5 font-mono text-[0.9em] text-[var(--fg)]">
                {node.value}
              </code>
            );
          case "text":
            return node.value;
        }
      }}
    </For>
  );
}

/** Renders one line of inline markdown: links, bold, italics and code. */
export function InlineMarkdown(props: { content: string }) {
  const nodes = createMemo(() => parseInline(props.content));
  return <Inline nodes={nodes()} />;
}

interface MarkdownProps {
  content: string;
  class?: string;
}

export function Markdown(props: MarkdownProps) {
  const paragraphs = createMemo(() => parseParagraphs(props.content));

  return (
    <div class={props.class}>
      <For each={paragraphs()}>
        {(nodes) => (
          <p class="text-pretty">
            <Inline nodes={nodes} />
          </p>
        )}
      </For>
    </div>
  );
}
