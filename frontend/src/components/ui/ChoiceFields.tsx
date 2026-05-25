"use client";

import { useEffect, useMemo, useState } from "react";
import { cn } from "@/lib/utils";

export type ChoiceOption = {
  value: string;
  label: string;
};

const CUSTOM = "__custom__";

type Props = {
  value: string;
  onChange: (value: string) => void;
  options: ChoiceOption[];
  customLabel: string;
  placeholder?: string;
  className?: string;
  selectClassName?: string;
};

export function ChoiceOrCustom({
  value,
  onChange,
  options,
  customLabel,
  placeholder,
  className,
  selectClassName,
}: Props) {
  const presetValues = useMemo(() => new Set(options.map((o) => o.value)), [options]);
  const isCustomMode = value !== "" && !presetValues.has(value);

  const [selectVal, setSelectVal] = useState(() => {
    if (!value) return "";
    return presetValues.has(value) ? value : CUSTOM;
  });
  const [customText, setCustomText] = useState(isCustomMode ? value : "");

  useEffect(() => {
    if (!value) {
      setSelectVal("");
      setCustomText("");
      return;
    }
    if (presetValues.has(value)) {
      setSelectVal(value);
      setCustomText("");
    } else {
      setSelectVal(CUSTOM);
      setCustomText(value);
    }
  }, [value, presetValues]);

  const fieldClass =
    selectClassName ??
    "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";

  function handleSelectChange(next: string) {
    setSelectVal(next);
    if (next === CUSTOM) {
      onChange(customText.trim());
    } else if (next === "") {
      onChange("");
    } else {
      onChange(next);
    }
  }

  function handleCustomChange(text: string) {
    setCustomText(text);
    if (selectVal === CUSTOM) {
      onChange(text.trim());
    }
  }

  return (
    <div className={cn("space-y-2", className)}>
      <select className={fieldClass} value={selectVal} onChange={(e) => handleSelectChange(e.target.value)}>
        <option value="">{customLabel}</option>
        {options.map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
        <option value={CUSTOM}>{placeholder ?? "…"}</option>
      </select>
      {selectVal === CUSTOM && (
        <input
          className={fieldClass}
          value={customText}
          placeholder={placeholder}
          onChange={(e) => handleCustomChange(e.target.value)}
        />
      )}
    </div>
  );
}

type MultiProps = {
  values: string[];
  onChange: (values: string[]) => void;
  options: ChoiceOption[];
  otherLabel: string;
  otherPlaceholder?: string;
  className?: string;
};

export function MultiChoiceWithOther({
  values,
  onChange,
  options,
  otherLabel,
  otherPlaceholder,
  className,
}: MultiProps) {
  const optionValues = new Set(options.map((o) => o.value));
  const selectedPresets = values.filter((v) => optionValues.has(v));
  const otherText = values.find((v) => !optionValues.has(v)) ?? "";

  function togglePreset(val: string) {
    const has = selectedPresets.includes(val);
    const nextPresets = has ? selectedPresets.filter((v) => v !== val) : [...selectedPresets, val];
    onChange(otherText ? [...nextPresets, otherText] : nextPresets);
  }

  function setOther(text: string) {
    const trimmed = text.trim();
    if (trimmed) {
      onChange([...selectedPresets, trimmed]);
    } else {
      onChange(selectedPresets);
    }
  }

  return (
    <div className={cn("space-y-3", className)}>
      <div className="flex flex-wrap gap-2">
        {options.map((o) => (
          <Chip key={o.value} active={selectedPresets.includes(o.value)} onClick={() => togglePreset(o.value)}>
            {o.label}
          </Chip>
        ))}
      </div>
      <input
        className="w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent"
        value={otherText}
        placeholder={otherPlaceholder ?? otherLabel}
        onChange={(e) => setOther(e.target.value)}
      />
    </div>
  );
}

function Chip({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "rounded-full border px-3 py-1 text-xs transition-colors",
        active ? "border-accent bg-accent/10 text-accent font-medium" : "border-border2 text-text2 hover:border-border"
      )}
    >
      {children}
    </button>
  );
}

export function RangeSelect({
  value,
  onChange,
  options,
  className,
}: {
  value: string;
  onChange: (value: string) => void;
  options: ChoiceOption[];
  className?: string;
}) {
  const fieldClass =
    "w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent";
  return (
    <select className={cn(fieldClass, className)} value={value} onChange={(e) => onChange(e.target.value)}>
      {options.map((o) => (
        <option key={o.value} value={o.value}>
          {o.label}
        </option>
      ))}
    </select>
  );
}
