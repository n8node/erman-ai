"use client";

import { useEffect, useId, useRef, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { intlLocale } from "@/i18n/intl-locale";
import type { StrategyDiagram, StrategyPriorityRow, StrategyROISummary } from "@/lib/api-strategy";

function formatRub(n: number, locale: string) {
  return new Intl.NumberFormat(intlLocale(locale), { maximumFractionDigits: 0 }).format(n) + " ₽";
}

export function StrategyROIBarChart({ summary }: { summary: StrategyROISummary }) {
  const t = useTranslations("strategy.result");
  const locale = useLocale();
  const max = Math.max(...summary.lines.map((l) => l.net_benefit_monthly_rub), 1);
  return (
    <div className="space-y-3">
      {summary.lines.map((line) => {
        const pct = Math.max(8, (line.net_benefit_monthly_rub / max) * 100);
        return (
          <div key={line.run_id}>
            <div className="mb-1 flex justify-between gap-2 text-xs">
              <span className="font-medium text-text">{line.process_name}</span>
              <span className="text-text2">
                {formatRub(line.net_benefit_monthly_rub, locale)}/{t("months")}
              </span>
            </div>
            <div className="h-2 overflow-hidden rounded-full bg-bg2">
              <div className="h-full rounded-full bg-accent transition-all" style={{ width: `${pct}%` }} />
            </div>
          </div>
        );
      })}
      <div className="grid gap-3 sm:grid-cols-2 pt-2">
        <div className="rounded-lg bg-accent-bg px-4 py-3">
          <p className="text-[10px] uppercase tracking-wider text-accent">Σ / {t("months")}</p>
          <p className="mt-1 text-lg font-medium text-accent">{formatRub(summary.total_monthly_benefit_rub, locale)}</p>
        </div>
        <div className="rounded-lg bg-accent-bg px-4 py-3">
          <p className="text-[10px] uppercase tracking-wider text-accent">Σ NPV</p>
          <p className="mt-1 text-lg font-medium text-accent">{formatRub(summary.total_npv_rub, locale)}</p>
        </div>
      </div>
    </div>
  );
}

export function StrategyPriorityMatrix({ rows }: { rows: StrategyPriorityRow[] }) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[640px] text-left text-xs">
        <thead>
          <tr className="border-b border-border text-[10px] uppercase tracking-wider text-text3">
            <th className="pb-2 pr-3">#</th>
            <th className="pb-2 pr-3">Use case</th>
            <th className="pb-2 pr-3">Impact</th>
            <th className="pb-2 pr-3">Effort</th>
            <th className="pb-2">Score</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((r) => (
            <tr key={r.priority} className="border-b border-border">
              <td className="py-2 pr-3 font-mono text-text3">{r.priority}</td>
              <td className="py-2 pr-3 text-text">{r.use_case}</td>
              <td className="py-2 pr-3">
                <ImpactBar value={r.impact} color="bg-ai" />
              </td>
              <td className="py-2 pr-3">
                <ImpactBar value={r.effort} color="bg-amber-500" />
              </td>
              <td className="py-2 font-medium text-text">{r.score.toFixed(1)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function ImpactBar({ value, color }: { value: number; color: string }) {
  const pct = Math.min(100, (value / 5) * 100);
  return (
    <div className="flex items-center gap-2">
      <div className="h-1.5 w-16 overflow-hidden rounded-full bg-bg2">
        <div className={`h-full rounded-full ${color}`} style={{ width: `${pct}%` }} />
      </div>
      <span className="text-text3">{value}</span>
    </div>
  );
}

function isMermaidErrorSvg(svg: string): boolean {
  return svg.includes('class="error-text"') || svg.includes("Syntax error in text");
}

export function StrategyMermaidDiagram({ diagram }: { diagram: StrategyDiagram }) {
  const t = useTranslations("strategy.result");
  const ref = useRef<HTMLDivElement>(null);
  const uid = useId().replace(/:/g, "");
  const [renderFailed, setRenderFailed] = useState(false);

  useEffect(() => {
    let cancelled = false;
    async function render() {
      if (!ref.current || !diagram.mermaid?.trim()) return;
      setRenderFailed(false);
      try {
        const mermaid = (await import("mermaid")).default;
        mermaid.initialize({ startOnLoad: false, theme: "neutral", securityLevel: "strict" });
        const { svg } = await mermaid.render(`strategy-diagram-${uid}`, diagram.mermaid);
        if (cancelled || !ref.current) return;
        if (isMermaidErrorSvg(svg)) {
          ref.current.innerHTML = "";
          setRenderFailed(true);
          return;
        }
        ref.current.innerHTML = svg;
      } catch {
        if (!cancelled && ref.current) {
          ref.current.innerHTML = "";
          setRenderFailed(true);
        }
      }
    }
    void render();
    return () => {
      cancelled = true;
    };
  }, [diagram.mermaid, uid]);

  return (
    <div className="rounded-lg border border-border bg-bg2/30 p-4">
      <p className="mb-2 text-sm font-medium text-text">{diagram.title}</p>
      {renderFailed ? (
        <div className="space-y-2">
          <p className="text-xs text-warning">{t("diagramRenderFailed")}</p>
          <details className="rounded-lg border border-border bg-bg px-3 py-2">
            <summary className="cursor-pointer text-[11px] text-text3">{t("diagramSource")}</summary>
            <pre className="mt-2 overflow-x-auto whitespace-pre-wrap font-mono text-[11px] text-text2">
              {diagram.mermaid}
            </pre>
          </details>
        </div>
      ) : (
        <div ref={ref} className="overflow-x-auto text-xs text-text2 [&_svg]:max-w-full" />
      )}
    </div>
  );
}
