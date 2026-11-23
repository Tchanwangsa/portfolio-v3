import { createEffect } from "solid-js";
import { AboutSection } from "../components/AboutSection";
import { TechnicalContributions } from "../components/TechnicalContributions";
import { StatsSection } from "../components/StatsSection";

export function AboutPage() {
  createEffect(() => {
    document.title = "Tanat Chanwangsa";
  });

  return (
    <div class="space-y-12">
      <AboutSection />
      <TechnicalContributions />
      <StatsSection />
    </div>
  );
}
