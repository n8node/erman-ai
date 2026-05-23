#!/usr/bin/env bash
# Safe .env reader (handles values with spaces). Usage: env_get KEY [default]
env_get() {
  local key="$1"
  local default="${2:-}"
  local line val
  line=$(grep -E "^${key}=" .env 2>/dev/null | head -1 || true)
  if [[ -z "$line" ]]; then
    echo "$default"
    return
  fi
  val="${line#*=}"
  val="${val%$'\r'}"
  val="${val#\"}"
  val="${val%\"}"
  val="${val#\'}"
  val="${val%\'}"
  echo "${val:-$default}"
}
