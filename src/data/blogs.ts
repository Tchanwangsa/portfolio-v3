import { compareDatesDesc } from "../lib/date";

export interface BlogPost {
  slug: string;
  title: string;
  date: string; // ISO YYYY-MM-DD
  body: string[];
}

const RAW_BLOG_POSTS: BlogPost[] = [
  {
    slug: "building-a-quieter-portfolio",
    title: "Building a Quieter Portfolio",
    date: "2026-11-12",
    body: [
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
      "Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
      "Curabitur pretium tincidunt lacus. Nulla gravida orci a odio. Nullam varius, turpis et commodo pharetra, est eros bibendum elit, nec luctus magna felis sollicitudin mauris.",
    ],
  },
  {
    slug: "notes-on-agentic-coding-workflows",
    title: "Notes on Agentic Coding Workflows",
    date: "2026-09-28",
    body: [
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Integer nec odio. Praesent libero. Sed cursus ante dapibus diam. Sed nisi.",
      "Nulla quis sem at nibh elementum imperdiet. Duis sagittis ipsum. Praesent mauris. Fusce nec tellus sed augue semper porta.",
    ],
  },
  {
    slug: "why-i-reach-for-solidjs",
    title: "Why I Reach for SolidJS",
    date: "2026-05-09",
    body: [
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Mauris massa. Vestibulum lacinia arcu eget nulla. Class aptent taciti sociosqu ad litora torquent per conubia nostra.",
      "Curabitur sodales ligula in libero. Sed dignissim lacinia nunc. Curabitur tortor. Pellentesque nibh. Aenean quam.",
    ],
  },
  {
    slug: "measuring-what-matters-in-side-projects",
    title: "Measuring What Matters in Side Projects",
    date: "2025-12-03",
    body: [
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. In scelerisque sem at dolor. Maecenas mattis. Sed convallis tristique sem.",
      "Proin ut ligula vel nunc egestas porttitor. Morbi lectus risus, iaculis vel, suscipit quis, luctus non, massa.",
    ],
  },
  {
    slug: "running-every-week-for-a-year",
    title: "Running Every Week for a Year",
    date: "2025-07-17",
    body: [
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce ac turpis quis ligula lacinia aliquet. Mauris ipsum. Nulla metus metus, ullamcorper vel, tincidunt sed, euismod in, nibh.",
      "Quisque volutpat condimentum velit. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos.",
    ],
  },
  {
    slug: "first-principles-of-fast-interfaces",
    title: "First Principles of Fast Interfaces",
    date: "2024-10-21",
    body: [
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam nec ante. Sed lacinia, urna non tincidunt mattis, tortor neque adipiscing diam, a cursus ipsum ante quis turpis.",
      "Nulla facilisi. Ut fringilla. Suspendisse potenti. Nunc feugiat mi a tellus consequat imperdiet.",
    ],
  },
];

export const BLOG_POSTS: BlogPost[] = [...RAW_BLOG_POSTS].sort((a, b) =>
  compareDatesDesc(a.date, b.date)
);

export function getBlogPostBySlug(slug: string): BlogPost | undefined {
  return BLOG_POSTS.find((post) => post.slug === slug);
}
