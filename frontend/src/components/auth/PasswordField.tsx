"use client";

import { Eye, EyeOff } from "lucide-react";
import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import {
  checkPasswordRules,
  isPasswordValid,
  passwordStrength,
  PASSWORD_RULE_KEYS,
  type PasswordRuleKey,
  type PasswordStrength,
} from "@/lib/password-policy";
import { cn } from "@/lib/utils";

const inputClass =
  "w-full rounded-lg border border-border2 py-2 pl-3 pr-10 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

const STRENGTH_COLORS: Record<PasswordStrength, string> = {
  0: "bg-border",
  1: "bg-[#e8a0a0]",
  2: "bg-[#e8c47a]",
  3: "bg-[#7eb3e8]",
  4: "bg-[#8bc96a]",
};

type PasswordFieldProps = {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  autoComplete?: string;
  showStrength?: boolean;
  showRequirements?: boolean;
};

export function PasswordField({
  id,
  label,
  value,
  onChange,
  autoComplete,
  showStrength = false,
  showRequirements = false,
}: PasswordFieldProps) {
  const t = useTranslations("auth.passwordPolicy");
  const [visible, setVisible] = useState(false);

  const rules = useMemo(() => checkPasswordRules(value), [value]);
  const strength = useMemo(() => passwordStrength(value, rules), [value, rules]);
  const valid = isPasswordValid(rules);

  const strengthLabel =
    strength === 0
      ? ""
      : strength === 1
        ? t("weak")
        : strength === 2
          ? t("fair")
          : strength === 3
            ? t("good")
            : t("strong");

  return (
    <div>
      <label htmlFor={id} className="mb-1.5 block text-xs font-medium">
        {label}
      </label>
      <div className="relative">
        <input
          id={id}
          type={visible ? "text" : "password"}
          autoComplete={autoComplete}
          required
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className={inputClass}
          aria-describedby={
            showStrength || showRequirements ? `${id}-hints` : undefined
          }
        />
        <button
          type="button"
          tabIndex={-1}
          onClick={() => setVisible((v) => !v)}
          className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 text-text3 hover:text-text2"
          aria-label={visible ? t("hide") : t("show")}
        >
          {visible ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
        </button>
      </div>

      {(showStrength || showRequirements) && (
        <div id={`${id}-hints`} className="mt-2 space-y-2">
          {showStrength && value.length > 0 && (
            <div className="space-y-1">
              <div className="flex gap-1" role="meter" aria-valuenow={strength} aria-valuemin={0} aria-valuemax={4}>
                {[1, 2, 3, 4].map((level) => (
                  <div
                    key={level}
                    className={cn(
                      "h-1 flex-1 rounded-full transition-colors duration-200",
                      strength >= level ? STRENGTH_COLORS[strength] : "bg-border2"
                    )}
                  />
                ))}
              </div>
              <p
                className={cn(
                  "text-[11px] font-medium",
                  strength <= 1 && "text-[#a32d2d]",
                  strength === 2 && "text-[#ba7517]",
                  strength === 3 && "text-accent",
                  strength >= 4 && "text-[#3b6d11]"
                )}
              >
                {strengthLabel}
              </p>
            </div>
          )}

          {showRequirements && (
            <ul className="space-y-1 text-[11px] text-text2">
              {PASSWORD_RULE_KEYS.map((key) => (
                <RuleItem key={key} ruleKey={key} met={rules[key]} label={t(key)} />
              ))}
            </ul>
          )}
        </div>
      )}

      {showRequirements && value.length > 0 && !valid && (
        <span className="sr-only">{t("invalid")}</span>
      )}
    </div>
  );
}

function RuleItem({
  ruleKey,
  met,
  label,
}: {
  ruleKey: PasswordRuleKey;
  met: boolean;
  label: string;
}) {
  return (
    <li
      className={cn(
        "flex items-center gap-2 transition-colors",
        met ? "text-[#3b6d11]" : "text-text3"
      )}
    >
      <span
        className={cn(
          "inline-flex h-3.5 w-3.5 shrink-0 items-center justify-center rounded-full border text-[9px]",
          met ? "border-[#3b6d11] bg-[#eaf3de]" : "border-border2 bg-bg2"
        )}
        aria-hidden
      >
        {met ? "✓" : ""}
      </span>
      <span data-rule={ruleKey}>{label}</span>
    </li>
  );
}

export { checkPasswordRules, isPasswordValid };
