import { createEffect } from "solid-js";
import { EntryList } from "../components/EntryList";
import { PROJECTS } from "../data/projects";

export function ProjectsPage() {
  createEffect(() => {
    document.title = "Projects — Tanat Chanwangsa";
  });

  const items = PROJECTS.map((project) => ({
    href: `/projects/${project.slug}`,
    title: project.title,
    date: project.date,
  }));

  return (
    <section aria-labelledby="projects-heading">
      <h1 id="projects-heading" class="sr-only">
        Projects
      </h1>

      <EntryList
        items={items}
        ariaLabel="Projects"
        emptyMessage="No projects published yet."
      />
    </section>
  );
}
