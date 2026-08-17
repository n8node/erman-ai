const TELEGRAM_BOT_TOKEN = /^\d{8,12}:[A-Za-z0-9_-]{20,}$/;
const MAX_BOT_TOKEN = /^[A-Za-z0-9]{32,128}$/;

export function validateTelegramBotToken(token: string): string | null {
  const value = token.trim();
  if (!value) return null;
  if (/\s/.test(value)) {
    return "invalidTelegramToken";
  }
  if (!TELEGRAM_BOT_TOKEN.test(value)) {
    return "invalidTelegramToken";
  }
  return null;
}

export function validateMaxBotToken(token: string): string | null {
  const value = token.trim();
  if (!value) return null;
  if (/\s/.test(value)) {
    return "invalidMaxToken";
  }
  if (!MAX_BOT_TOKEN.test(value)) {
    return "invalidMaxToken";
  }
  return null;
}
