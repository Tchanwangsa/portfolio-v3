import { render } from "solid-js/web";
import { Navigate, Route, Router, useParams } from "@solidjs/router";
import "./index.css";
import App from "./App";
import { AboutPage } from "./pages/AboutPage";
import { ProjectsPage } from "./pages/ProjectsPage";
import { ProjectDetailPage } from "./pages/ProjectDetailPage";
import { BlogsPage } from "./pages/BlogsPage";
import { BlogDetailPage } from "./pages/BlogDetailPage";

function BlogSlugRedirect() {
  const params = useParams<{ slug: string }>();
  return <Navigate href={`/blogs/${params.slug}`} />;
}

const root = document.getElementById("root");

if (root) {
  render(
    () => (
      <Router root={App}>
        <Route path="/" component={AboutPage} />
        <Route path="/projects">
          <Route path="/" component={ProjectsPage} />
          <Route path="/:slug" component={ProjectDetailPage} />
        </Route>
        <Route path="/blogs">
          <Route path="/" component={BlogsPage} />
          <Route path="/:slug" component={BlogDetailPage} />
        </Route>
        <Route path="/blog" component={() => <Navigate href="/blogs" />} />
        <Route path="/blog/:slug" component={BlogSlugRedirect} />
        <Route path="*" component={() => <Navigate href="/" />} />
      </Router>
    ),
    root
  );
}
