"use client";

import { useCallback, useEffect, useRef } from "react";

export const WORKSPACE_SSO_PATH =
  "/workspace-app/login?provider=OIDC&flow=redirect";

interface WorkspaceFrameProps {
  title: string;
}

export function WorkspaceFrame({ title }: WorkspaceFrameProps) {
  const frameRef = useRef<HTMLIFrameElement>(null);

  const ensureSSO = useCallback(() => {
    const frameWindow = frameRef.current?.contentWindow;
    if (!frameWindow) return;

    try {
      const currentURL = new URL(frameWindow.location.href);
      const { pathname, searchParams } = currentURL;
      const isRootLogin = pathname === "/login";
      const isWorkspaceLogin = pathname === "/workspace-app/login";
      const isOIDCRedirect =
        searchParams.get("provider") === "OIDC" &&
        searchParams.get("flow") === "redirect";

      if (isRootLogin || (isWorkspaceLogin && !isOIDCRedirect)) {
        frameWindow.location.replace(WORKSPACE_SSO_PATH);
      }
    } catch {
      // The OIDC round trip can briefly navigate outside the iframe origin.
    }
  }, []);

  useEffect(() => {
    const interval = window.setInterval(ensureSSO, 750);
    return () => window.clearInterval(interval);
  }, [ensureSSO]);

  return (
    <iframe
      ref={frameRef}
      src={WORKSPACE_SSO_PATH}
      title={title}
      className="min-h-0 flex-1 border-0 bg-white"
      allow="clipboard-read; clipboard-write; fullscreen"
      onLoad={ensureSSO}
    />
  );
}
