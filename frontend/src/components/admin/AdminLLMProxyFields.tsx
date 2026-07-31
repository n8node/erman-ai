"use client";

import { useTranslations } from "next-intl";
import { cn } from "@/lib/utils";
import type { LLMHTTPProxySettings } from "@/lib/api";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const templateClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm font-mono outline-none focus:border-accent focus:ring-1 focus:ring-accent";

type Props = {
  providerLabel: string;
  value: LLMHTTPProxySettings;
  onChange: (next: LLMHTTPProxySettings) => void;
};

export function AdminLLMProxyFields({ providerLabel, value, onChange }: Props) {
  const t = useTranslations("admin.llmProxy");

  function patch(partial: Partial<LLMHTTPProxySettings>) {
    onChange({ ...value, ...partial });
  }

  return (
    <div className="rounded-lg border border-border bg-bg2/40 p-4 space-y-3">
      <div>
        <p className="text-sm font-medium">{t("sectionTitle", { provider: providerLabel })}</p>
        <p className="mt-1 text-xs text-text3">{t("hint")}</p>
      </div>
      <label className="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          checked={value.enabled}
          onChange={(e) => patch({ enabled: e.target.checked })}
          className="rounded border-border2"
        />
        {t("enabled")}
      </label>
      <div>
        <label className="mb-1.5 block text-xs font-medium">{t("list")}</label>
        <textarea
          value={(value.proxy_urls || []).join("\n")}
          onChange={(e) => {
            const proxy_urls = e.target.value
              .split(/\r?\n/)
              .map((v) => v.trim())
              .filter(Boolean);
            patch({
              proxy_urls,
              enabled: proxy_urls.length > 0 ? true : value.enabled,
              proxy_active_url:
                value.proxy_active_url && proxy_urls.includes(value.proxy_active_url)
                  ? value.proxy_active_url
                  : proxy_urls[0] || "",
            });
          }}
          className={templateClass}
          rows={3}
          placeholder="http://user:pass@5.35.83.120:3128"
        />
        <p className="mt-1 text-xs text-text3">{t("listHint")}</p>
      </div>
      <div>
        <label className="mb-1.5 block text-xs font-medium">{t("active")}</label>
        <select
          value={value.proxy_active_url || ""}
          onChange={(e) => patch({ proxy_active_url: e.target.value })}
          className={fieldClass}
        >
          <option value="">{t("activeAuto")}</option>
          {(value.proxy_urls || []).map((proxyURL) => (
            <option key={proxyURL} value={proxyURL}>
              {proxyURL}
            </option>
          ))}
        </select>
      </div>
      <label className="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          checked={value.proxy_auto_failover}
          onChange={(e) => patch({ proxy_auto_failover: e.target.checked })}
          className="rounded border-border2"
        />
        {t("autoFailover")}
      </label>
    </div>
  );
}
