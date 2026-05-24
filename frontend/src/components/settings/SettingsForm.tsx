"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import type { User } from "@/lib/api";
import { changePassword, updateMe } from "@/lib/api";

type Props = {
  user: User;
};

export function SettingsForm({ user }: Props) {
  const t = useTranslations("settings");
  const [email, setEmail] = useState(user.email);
  const [locale, setLocale] = useState(user.locale);
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [profileMsg, setProfileMsg] = useState("");
  const [passwordMsg, setPasswordMsg] = useState("");
  const [profileError, setProfileError] = useState("");
  const [passwordError, setPasswordError] = useState("");
  const [profileLoading, setProfileLoading] = useState(false);
  const [passwordLoading, setPasswordLoading] = useState(false);

  async function handleProfile(e: React.FormEvent) {
    e.preventDefault();
    setProfileMsg("");
    setProfileError("");
    setProfileLoading(true);
    try {
      await updateMe(email, locale);
      setProfileMsg(t("profileSaved"));
    } catch (err) {
      setProfileError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setProfileLoading(false);
    }
  }

  async function handlePassword(e: React.FormEvent) {
    e.preventDefault();
    setPasswordMsg("");
    setPasswordError("");
    setPasswordLoading(true);
    try {
      await changePassword(currentPassword, newPassword);
      setPasswordMsg(t("passwordSaved"));
      setCurrentPassword("");
      setNewPassword("");
    } catch (err) {
      setPasswordError(err instanceof Error ? err.message : t("saveFailed"));
    } finally {
      setPasswordLoading(false);
    }
  }

  return (
    <div className="mx-auto max-w-xl space-y-6">
      <div>
        <h1 className="text-base font-medium">{t("title")}</h1>
        <p className="mt-1 text-sm text-text2">{t("subtitle")}</p>
      </div>

      <section className="rounded-xl border border-border bg-bg p-6">
        <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
          {t("profileSection")}
        </h2>
        <form onSubmit={handleProfile} className="mt-4 space-y-4">
          {profileError && (
            <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
              {profileError}
            </div>
          )}
          {profileMsg && (
            <div className="rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-800">
              {profileMsg}
            </div>
          )}
          <div>
            <label htmlFor="email" className="mb-1.5 block text-xs font-medium">
              {t("email")}
            </label>
            <input
              id="email"
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent"
            />
          </div>
          <div>
            <label htmlFor="locale" className="mb-1.5 block text-xs font-medium">
              {t("locale")}
            </label>
            <select
              id="locale"
              value={locale}
              onChange={(e) => setLocale(e.target.value)}
              className="w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent"
            >
              <option value="ru">Русский</option>
              <option value="en">English</option>
            </select>
          </div>
          <button
            type="submit"
            disabled={profileLoading}
            className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:opacity-60"
          >
            {profileLoading ? t("saving") : t("saveProfile")}
          </button>
        </form>
      </section>

      <section className="rounded-xl border border-border bg-bg p-6">
        <h2 className="text-[10px] font-medium uppercase tracking-wider text-text3">
          {t("passwordSection")}
        </h2>
        <form onSubmit={handlePassword} className="mt-4 space-y-4">
          {passwordError && (
            <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
              {passwordError}
            </div>
          )}
          {passwordMsg && (
            <div className="rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-800">
              {passwordMsg}
            </div>
          )}
          <div>
            <label
              htmlFor="currentPassword"
              className="mb-1.5 block text-xs font-medium"
            >
              {t("currentPassword")}
            </label>
            <input
              id="currentPassword"
              type="password"
              required
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              className="w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent"
            />
          </div>
          <div>
            <label
              htmlFor="newPassword"
              className="mb-1.5 block text-xs font-medium"
            >
              {t("newPassword")}
            </label>
            <input
              id="newPassword"
              type="password"
              required
              minLength={8}
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              className="w-full rounded-lg border border-border2 px-3 py-2 text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent"
            />
          </div>
          <button
            type="submit"
            disabled={passwordLoading}
            className="rounded-lg bg-text px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:opacity-60"
          >
            {passwordLoading ? t("saving") : t("savePassword")}
          </button>
        </form>
      </section>
    </div>
  );
}
