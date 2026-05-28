export type PasswordRules = {
  minLength: boolean;
  lowercase: boolean;
  uppercase: boolean;
  digit: boolean;
  special: boolean;
};

export type PasswordStrength = 0 | 1 | 2 | 3 | 4;

const LOWER = /[a-z]/;
const UPPER = /[A-Z]/;
const DIGIT = /[0-9]/;
const SPECIAL = /[^a-zA-Z0-9]/;

export function checkPasswordRules(password: string): PasswordRules {
  return {
    minLength: password.length >= 8,
    lowercase: LOWER.test(password),
    uppercase: UPPER.test(password),
    digit: DIGIT.test(password),
    special: SPECIAL.test(password),
  };
}

export function isPasswordValid(rules: PasswordRules): boolean {
  return (
    rules.minLength &&
    rules.lowercase &&
    rules.uppercase &&
    rules.digit &&
    rules.special
  );
}

export function passwordStrength(password: string, rules: PasswordRules): PasswordStrength {
  if (!password) return 0;
  const met = [
    rules.minLength,
    rules.lowercase,
    rules.uppercase,
    rules.digit,
    rules.special,
  ].filter(Boolean).length;
  if (met <= 2) return 1;
  if (met === 3) return 2;
  if (met === 4) return 3;
  return password.length >= 12 ? 4 : 3;
}

export const PASSWORD_RULE_KEYS = [
  "minLength",
  "lowercase",
  "uppercase",
  "digit",
  "special",
] as const;

export type PasswordRuleKey = (typeof PASSWORD_RULE_KEYS)[number];
