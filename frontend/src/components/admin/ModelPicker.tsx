"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { cn } from "@/lib/utils";

type Props = {
  label: string;
  value: string;
  models: string[];
  onChange: (value: string) => void;
  placeholder?: string;
};

export function ModelPicker({ label, value, models, onChange, placeholder }: Props) {
  const [open, setOpen] = useState(false);
  const [filter, setFilter] = useState("");
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function onDocClick(e: MouseEvent) {
      if (!rootRef.current?.contains(e.target as Node)) {
        setOpen(false);
        setFilter("");
      }
    }
    document.addEventListener("mousedown", onDocClick);
    return () => document.removeEventListener("mousedown", onDocClick);
  }, []);

  const filtered = useMemo(() => {
    const q = filter.trim().toLowerCase();
    if (!q) return models;
    return models.filter((m) => m.toLowerCase().includes(q));
  }, [models, filter]);

  function openList() {
    setOpen(true);
    setFilter("");
  }

  function closeList() {
    setOpen(false);
    setFilter("");
  }

  return (
    <div ref={rootRef} className="relative">
      <label className="mb-1.5 block text-xs font-medium">{label}</label>
      <input
        className="w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent"
        value={open ? filter : value}
        placeholder={placeholder}
        onChange={(e) => {
          setFilter(e.target.value);
          onChange(e.target.value);
          setOpen(true);
        }}
        onFocus={openList}
        onClick={openList}
      />
      {open && (
        <ul className="absolute z-20 mt-1 max-h-52 w-full overflow-auto rounded-lg border border-border bg-bg py-1 shadow-lg">
          {models.length === 0 ? (
            <li className="px-3 py-2 text-xs text-text3">{placeholder}</li>
          ) : filtered.length === 0 ? (
            <li className="px-3 py-2 text-xs text-text3">—</li>
          ) : (
            filtered.slice(0, 100).map((model) => (
              <li key={model}>
                <button
                  type="button"
                  className={cn(
                    "w-full px-3 py-1.5 text-left text-xs hover:bg-bg2",
                    model === value && "bg-bg2 font-medium"
                  )}
                  onMouseDown={(e) => e.preventDefault()}
                  onClick={() => {
                    onChange(model);
                    closeList();
                  }}
                >
                  {model}
                </button>
              </li>
            ))
          )}
          {filtered.length > 100 && (
            <li className="border-t border-border px-3 py-2 text-[10px] text-text3">
              +{filtered.length - 100} …
            </li>
          )}
        </ul>
      )}
    </div>
  );
}
