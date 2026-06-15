"use client";

import { useEffect, useRef } from "react";

type Props = {
  counterCode: string;
};

export function YandexMetrikaInjector({ counterCode }: Props) {
  const injected = useRef(false);

  useEffect(() => {
    const code = counterCode.trim();
    if (!code || injected.current) return;

    const container = document.createElement("div");
    container.innerHTML = code;

    container.querySelectorAll("script").forEach((oldScript) => {
      const script = document.createElement("script");
      Array.from(oldScript.attributes).forEach((attr) => {
        script.setAttribute(attr.name, attr.value);
      });
      script.textContent = oldScript.textContent;
      document.body.appendChild(script);
    });

    container.querySelectorAll("noscript").forEach((ns) => {
      document.body.appendChild(ns.cloneNode(true));
    });

    injected.current = true;
  }, [counterCode]);

  return null;
}
