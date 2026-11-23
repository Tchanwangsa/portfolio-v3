import { createEffect } from "solid-js";
import { EntryList } from "../components/EntryList";
import { BLOG_POSTS } from "../data/blogs";

export function BlogsPage() {
  createEffect(() => {
    document.title = "Blog — Tanat Chanwangsa";
  });

  const items = BLOG_POSTS.map((post) => ({
    href: `/blogs/${post.slug}`,
    title: post.title,
    date: post.date,
  }));

  return (
    <section aria-labelledby="blogs-heading">
      <h1 id="blogs-heading" class="sr-only">
        Blog
      </h1>

      <EntryList
        items={items}
        ariaLabel="Blog posts"
        emptyMessage="No posts published yet."
      />
    </section>
  );
}
