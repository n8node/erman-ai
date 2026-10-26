import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "DrillAI — 3D цифровой двойник буровой площадки",
  description:
    "Интерактивная 3D-сцена кустовой площадки с классической и роботизированной буровыми установками.",
};

export default function DrillAiPage() {
  return (
    <main className="fixed inset-0 bg-[#0f141a]">
      <iframe
        title="DrillAI — 3D цифровой двойник буровой площадки"
        src="/dashboard/drillai/index.html"
        className="block h-full w-full border-0"
        allow="autoplay"
      />
    </main>
  );
}