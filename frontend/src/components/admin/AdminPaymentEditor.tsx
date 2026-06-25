"use client";

import { Copy, Eye, EyeOff, Zap } from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  fetchAdminPaymentSettings,
  testAdminPaymentConnection,
  updateAdminPaymentSettings,
  type PaymentAdminView,
  type PaymentProvider,
  type RobokassaAdminSettings,
  type YookassaAdminSettings,
} from "@/lib/api";
import { cn } from "@/lib/utils";

const fieldClass =
  "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

function SecretInput({
  value,
  onChange,
  placeholder,
}: {
  value: string;
  onChange: (v: string) => void;
  placeholder: string;
}) {
  const [visible, setVisible] = useState(false);
  return (
    <div className="relative">
      <input
        type={visible ? "text" : "password"}
        autoComplete="new-password"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className={cn(fieldClass, "pr-10")}
      />
      <button
        type="button"
        onClick={() => setVisible((v) => !v)}
        className="absolute right-2 top-1/2 -translate-y-1/2 text-text3 hover:text-text"
        aria-label={visible ? "Hide" : "Show"}
      >
        {visible ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
      </button>
    </div>
  );
}

function CopyField({ value, label }: { value: string; label: string }) {
  const t = useTranslations("admin.payments");
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      /* ignore */
    }
  }

  return (
    <div className="flex gap-2">
      <input type="text" readOnly value={value} className={cn(fieldClass, "font-mono text-xs")} />
      <button
        type="button"
        onClick={handleCopy}
        className="inline-flex shrink-0 items-center gap-1.5 rounded-lg border border-border2 px-3 py-2 text-sm hover:bg-bg2"
      >
        <Copy className="h-4 w-4" />
        {copied ? t("copied") : t("copy")}
      </button>
    </div>
  );
}

export function AdminPaymentEditor() {
  const t = useTranslations("admin.payments");
  const [activeProvider, setActiveProvider] = useState<PaymentProvider>("yookassa");
  const [yookassa, setYookassa] = useState<YookassaAdminSettings>({
    shop_id: "",
    return_url: "",
    enabled: false,
  });
  const [robokassa, setRobokassa] = useState<RobokassaAdminSettings>({
    merchant_login: "",
    test_mode: false,
    enabled: false,
  });
  const [yookassaSecret, setYookassaSecret] = useState("");
  const [robokassaPassword1, setRobokassaPassword1] = useState("");
  const [robokassaPassword2, setRobokassaPassword2] = useState("");
  const [yookassaSecretSet, setYookassaSecretSet] = useState(false);
  const [yookassaSecretHint, setYookassaSecretHint] = useState("");
  const [robokassaPassword1Set, setRobokassaPassword1Set] = useState(false);
  const [robokassaPassword1Hint, setRobokassaPassword1Hint] = useState("");
  const [robokassaPassword2Set, setRobokassaPassword2Set] = useState(false);
  const [robokassaPassword2Hint, setRobokassaPassword2Hint] = useState("");
  const [yookassaWebhookURL, setYookassaWebhookURL] = useState("");
  const [robokassaResultURL, setRobokassaResultURL] = useState("");
  const [robokassaResult2URL, setRobokassaResult2URL] = useState("");
  const [defaultReturnURL, setDefaultReturnURL] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [testMessage, setTestMessage] = useState("");

  function applyView(data: PaymentAdminView) {
    setActiveProvider(data.active_provider);
    setYookassa(data.yookassa);
    setRobokassa(data.robokassa);
    setYookassaSecretSet(data.yookassa_secret_set);
    setYookassaSecretHint(data.yookassa_secret_hint || "");
    setRobokassaPassword1Set(data.robokassa_password1_set);
    setRobokassaPassword1Hint(data.robokassa_password1_hint || "");
    setRobokassaPassword2Set(data.robokassa_password2_set);
    setRobokassaPassword2Hint(data.robokassa_password2_hint || "");
    setYookassaWebhookURL(data.yookassa_webhook_url);
    setRobokassaResultURL(data.robokassa_result_url);
    setRobokassaResult2URL(data.robokassa_result2_url);
    setDefaultReturnURL(data.default_return_url);
  }

  useEffect(() => {
    fetchAdminPaymentSettings()
      .then(applyView)
      .catch((err) =>
        setError(err instanceof Error ? err.message : t("loadFailed"))
      )
      .finally(() => setLoading(false));
  }, [t]);

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess("");
    try {
      const data = await updateAdminPaymentSettings({
        active_provider: activeProvider,
        yookassa: {
          ...yookassa,
          return_url: yookassa.return_url || defaultReturnURL,
        },
        robokassa,
        ...(yookassaSecret.trim() ? { yookassa_secret_key: yookassaSecret.trim() } : {}),
        ...(robokassaPassword1.trim() ? { robokassa_password1: robokassaPassword1.trim() } : {}),
        ...(robokassaPassword2.trim() ? { robokassa_password2: robokassaPassword2.trim() } : {}),
      });
      applyView(data);
      setYookassaSecret("");
      setRobokassaPassword1("");
      setRobokassaPassword2("");
      setSuccess(t("saved"));
    } catch (err) {
      setError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setSaving(false);
    }
  }

  async function handleTest(provider: PaymentProvider) {
    setTesting(true);
    setTestMessage("");
    setError("");
    try {
      if (yookassaSecret.trim() || robokassaPassword1.trim() || robokassaPassword2.trim()) {
        const saved = await updateAdminPaymentSettings({
          active_provider: activeProvider,
          yookassa: {
            ...yookassa,
            return_url: yookassa.return_url || defaultReturnURL,
          },
          robokassa,
          ...(yookassaSecret.trim() ? { yookassa_secret_key: yookassaSecret.trim() } : {}),
          ...(robokassaPassword1.trim() ? { robokassa_password1: robokassaPassword1.trim() } : {}),
          ...(robokassaPassword2.trim() ? { robokassa_password2: robokassaPassword2.trim() } : {}),
        });
        applyView(saved);
        setYookassaSecret("");
        setRobokassaPassword1("");
        setRobokassaPassword2("");
      }
      const result = await testAdminPaymentConnection(provider);
      setTestMessage(result.message);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("testFailed"));
    } finally {
      setTesting(false);
    }
  }

  if (loading) {
    return <p className="text-sm text-text2">{t("loading")}</p>;
  }

  const providers: { id: PaymentProvider; label: string }[] = [
    { id: "yookassa", label: t("providerYookassa") },
    { id: "robokassa", label: t("providerRobokassa") },
  ];

  return (
    <div className="space-y-6">
      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </div>
      )}
      {success && (
        <div className="rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-900">
          {success}
        </div>
      )}

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <div>
          <h2 className="text-sm font-medium">{t("activeProviderTitle")}</h2>
          <p className="mt-1 text-sm text-text2">{t("activeProviderHint")}</p>
        </div>
        <div className="flex flex-wrap gap-2">
          {providers.map(({ id, label }) => (
            <button
              key={id}
              type="button"
              onClick={() => {
                setActiveProvider(id);
                if (id === "robokassa") {
                  setRobokassa((p) => ({ ...p, enabled: true }));
                }
                setSuccess("");
              }}
              className={cn(
                "rounded-lg border px-5 py-2.5 text-sm font-medium transition-colors",
                activeProvider === id
                  ? "border-accent bg-accent/10 text-accent"
                  : "border-border2 text-text2 hover:bg-bg2"
              )}
            >
              {label}
            </button>
          ))}
        </div>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <div>
          <h2 className="text-sm font-medium">{t("yookassaTitle")}</h2>
          <p className="mt-1 text-sm text-text2">{t("yookassaSubtitle")}</p>
        </div>

        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("shopId")}</label>
          <input
            type="text"
            value={yookassa.shop_id}
            onChange={(e) => setYookassa((p) => ({ ...p, shop_id: e.target.value }))}
            className={fieldClass}
          />
          <p className="mt-1 text-xs text-text3">{t("shopIdHint")}</p>
        </div>

        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("secretKey")}</label>
          <SecretInput
            value={yookassaSecret}
            onChange={setYookassaSecret}
            placeholder={
              yookassaSecretSet
                ? t("secretPlaceholderExisting", { hint: yookassaSecretHint || "••••" })
                : t("secretPlaceholder")
            }
          />
          <p className="mt-1 text-xs text-text3">{t("secretHint")}</p>
        </div>

        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("returnUrl")}</label>
          <input
            type="url"
            value={yookassa.return_url || defaultReturnURL}
            onChange={(e) => setYookassa((p) => ({ ...p, return_url: e.target.value }))}
            className={fieldClass}
          />
          <p className="mt-1 text-xs text-text3">{t("returnUrlHint")}</p>
        </div>

        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={yookassa.enabled}
            onChange={(e) => setYookassa((p) => ({ ...p, enabled: e.target.checked }))}
            className="rounded border-border2"
          />
          {t("yookassaEnabled")}
        </label>

        <button
          type="button"
          disabled={saving}
          onClick={handleSave}
          className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:opacity-60"
        >
          {saving ? t("saving") : t("saveYookassa")}
        </button>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-3">
        <h2 className="text-xs font-medium uppercase tracking-wide text-text3">
          {t("testSection")}
        </h2>
        <p className="text-sm text-text2">{t("yookassaTestHint")}</p>
        <button
          type="button"
          disabled={testing}
          onClick={() => handleTest("yookassa")}
          className="inline-flex items-center gap-2 rounded-lg border border-border2 px-4 py-2 text-sm hover:bg-bg2 disabled:opacity-60"
        >
          <Zap className="h-4 w-4" />
          {testing ? t("testing") : t("testConnection")}
        </button>
        {testMessage && activeProvider === "yookassa" && (
          <p className={cn("text-sm", testMessage.includes("успеш") || testMessage.toLowerCase().includes("success") ? "text-green-800" : "text-text2")}>
            {testMessage}
          </p>
        )}
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <h2 className="text-xs font-medium uppercase tracking-wide text-text3">
          {t("yookassaInstructions")}
        </h2>
        <ol className="list-decimal space-y-2 pl-5 text-sm text-text2">
          <li>{t("yookassaStep1")}</li>
          <li>{t("yookassaStep2")}</li>
          <li>{t("yookassaStep3")}</li>
          <li>{t("yookassaStep4")}</li>
          <li>
            <span className="block mb-2">{t("yookassaStep5")}</span>
            <CopyField value={yookassaWebhookURL} label="webhook" />
          </li>
          <li>{t("yookassaStep6")}</li>
          <li>{t("yookassaStep7")}</li>
        </ol>
      </section>

      <section className="rounded-xl border border-border bg-bg p-5 space-y-4">
        <div>
          <h2 className="text-sm font-medium">{t("robokassaTitle")}</h2>
          <p className="mt-1 text-sm text-text2">{t("robokassaSubtitle")}</p>
        </div>

        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("merchantLogin")}</label>
          <input
            type="text"
            value={robokassa.merchant_login}
            onChange={(e) => setRobokassa((p) => ({ ...p, merchant_login: e.target.value }))}
            className={fieldClass}
          />
          <p className="mt-1 text-xs text-text3">{t("merchantLoginHint")}</p>
        </div>

        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("password1")}</label>
          <SecretInput
            value={robokassaPassword1}
            onChange={setRobokassaPassword1}
            placeholder={
              robokassaPassword1Set
                ? t("secretPlaceholderExisting", { hint: robokassaPassword1Hint || "••••" })
                : t("password1Placeholder")
            }
          />
          <p className="mt-1 text-xs text-text3">{t("password1Hint")}</p>
        </div>

        <div>
          <label className="mb-1.5 block text-xs font-medium">{t("password2")}</label>
          <SecretInput
            value={robokassaPassword2}
            onChange={setRobokassaPassword2}
            placeholder={
              robokassaPassword2Set
                ? t("secretPlaceholderExisting", { hint: robokassaPassword2Hint || "••••" })
                : t("password2Placeholder")
            }
          />
          <p className="mt-1 text-xs text-text3">{t("password2Hint")}</p>
        </div>

        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={robokassa.test_mode}
            onChange={(e) => setRobokassa((p) => ({ ...p, test_mode: e.target.checked }))}
            className="rounded border-border2"
          />
          {t("testMode")}
        </label>

        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={robokassa.enabled}
            onChange={(e) => setRobokassa((p) => ({ ...p, enabled: e.target.checked }))}
            className="rounded border-border2"
          />
          {t("robokassaEnabled")}
        </label>

        <button
          type="button"
          disabled={saving}
          onClick={handleSave}
          className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:opacity-60"
        >
          {saving ? t("saving") : t("saveRobokassa")}
        </button>
      </section>

      <section className="rounded-xl border border-bg2 bg-bg2 p-5 space-y-3">
        <h3 className="text-sm font-medium">{t("robokassaResultTitle")}</h3>
        <p className="text-sm text-text2">{t("robokassaResultHint")}</p>
        <CopyField value={robokassaResultURL} label="result-wp" />
        <ul className="list-disc space-y-1 pl-5 text-xs text-text3">
          <li>{t("robokassaNoteMd5")}</li>
          <li>{t("robokassaNoteShp")}</li>
          <li>{t("robokassaNoteWordPress")}</li>
        </ul>
      </section>

      <section className="rounded-xl border border-bg2 bg-bg2 p-5 space-y-3">
        <h3 className="text-sm font-medium">{t("robokassaResult2Title")}</h3>
        <p className="text-sm text-text2">{t("robokassaResult2Hint")}</p>
        <CopyField value={robokassaResult2URL} label="result2" />
        <ul className="list-disc space-y-1 pl-5 text-xs text-text3">
          <li>{t("robokassaNoteResult2Auto")}</li>
          <li>{t("robokassaNoteResult2Test")}</li>
        </ul>
        <a
          href="https://docs.robokassa.ru/ru/notifications-and-redirects"
          target="_blank"
          rel="noopener noreferrer"
          className="inline-block text-sm text-accent hover:underline"
        >
          {t("robokassaDocs")}
        </a>
      </section>

      {activeProvider === "robokassa" && (
        <section className="rounded-xl border border-border bg-bg p-5 space-y-3">
          <p className="text-sm text-text2">{t("robokassaTestHint")}</p>
          <button
            type="button"
            disabled={testing}
            onClick={() => handleTest("robokassa")}
            className="inline-flex items-center gap-2 rounded-lg border border-border2 px-4 py-2 text-sm hover:bg-bg2 disabled:opacity-60"
          >
            <Zap className="h-4 w-4" />
            {testing ? t("testing") : t("testConnection")}
          </button>
          {testMessage && (
            <p className={cn("text-sm", testMessage.includes("заполн") || testMessage.toLowerCase().includes("success") ? "text-green-800" : "text-text2")}>
              {testMessage}
            </p>
          )}
        </section>
      )}
    </div>
  );
}
