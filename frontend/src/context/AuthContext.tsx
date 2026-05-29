"use client";

import { createContext, useContext } from "react";
import type { User } from "@/lib/api";

const AuthContext = createContext<User | null>(null);

type Props = {
  user: User | null;
  children: React.ReactNode;
};

export function AuthProvider({ user, children }: Props) {
  return <AuthContext.Provider value={user}>{children}</AuthContext.Provider>;
}

export function useAuthUser(): User | null {
  return useContext(AuthContext);
}

export function useIsGuest(): boolean {
  return useContext(AuthContext) === null;
}
