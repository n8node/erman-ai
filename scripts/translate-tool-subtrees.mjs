/**
 * Fast parallel translation of tool i18n bundle into de/fr/es/zh.
 */
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const messagesDir = path.join(__dirname, "../frontend/src/messages");
const bundlePath = path.join(__dirname, "i18n-tools-en-bundle.json");
const cachePath = path.join(__dirname, ".i18n-subtree-cache.json");

const TARGETS = ["de", "fr", "es", "zh"];
const CONCURRENCY = 8;

function isPlainObject(v) {
  return typeof v === "object" && v !== null && !Array.isArray(v);
}

function flattenLeaves(obj, prefix = "") {
  const out = [];
  for (const [key, value] of Object.entries(obj || {})) {
    const pathKey = prefix ? `${prefix}.${key}` : key;
    if (typeof value === "string") out.push([pathKey, value]);
    else if (isPlainObject(value)) out.push(...flattenLeaves(value, pathKey));
  }
  return out;
}

function setPath(obj, parts, value) {
  let cur = obj;
  for (let i = 0; i < parts.length - 1; i++) {
    const p = parts[i];
    if (!isPlainObject(cur[p])) cur[p] = {};
    cur = cur[p];
  }
  cur[parts[parts.length - 1]] = value;
}

function unflattenLeaves(entries) {
  const out = {};
  for (const [pathKey, value] of entries) {
    setPath(out, pathKey.split("."), value);
  }
  return out;
}

function protectPlaceholders(text) {
  const tokens = [];
  const protectedText = text.replace(/\{[^}]+\}/g, (m) => {
    const token = `__PH${tokens.length}__`;
    tokens.push(m);
    return token;
  });
  return { protectedText, tokens };
}

function restorePlaceholders(text, tokens) {
  let out = text;
  for (let i = 0; i < tokens.length; i++) out = out.replaceAll(`__PH${i}__`, tokens[i]);
  return out;
}

function isBadTranslation(value) {
  return !value || /MYMEMORY|USAGELIMITS|QUERY LENGTH LIMIT/i.test(value);
}

let cache = {};
if (fs.existsSync(cachePath)) cache = JSON.parse(fs.readFileSync(cachePath, "utf8"));

function saveCache() {
  fs.writeFileSync(cachePath, JSON.stringify(cache, null, 2) + "\n");
}

async function translateOne(text, target) {
  const cacheKey = `${target}::${text}`;
  if (cache[cacheKey] && !isBadTranslation(cache[cacheKey])) return cache[cacheKey];

  const { protectedText, tokens } = protectPlaceholders(text);
  const url = new URL("https://api.mymemory.translated.net/get");
  url.searchParams.set("q", protectedText.slice(0, 480));
  url.searchParams.set("langpair", `en|${target}`);
  url.searchParams.set("de", "hello@erman.ai");

  const res = await fetch(url);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const data = await res.json();
  const translated = data?.responseData?.translatedText;
  if (!translated || isBadTranslation(translated)) throw new Error("bad");

  const result = restorePlaceholders(translated.trim(), tokens);
  cache[cacheKey] = result;
  return result;
}

async function translateBatch(items, target) {
  const results = [];
  for (let i = 0; i < items.length; i += CONCURRENCY) {
    const chunk = items.slice(i, i + CONCURRENCY);
    const chunkResults = await Promise.all(
      chunk.map(async ([pathKey, text]) => {
        try {
          const value = await translateOne(text, target);
          return [pathKey, value];
        } catch {
          return [pathKey, text];
        }
      })
    );
    results.push(...chunkResults);
    saveCache();
    if ((i + CONCURRENCY) % 40 === 0) {
      console.log(`    ${Math.min(i + CONCURRENCY, items.length)}/${items.length}`);
    }
    await new Promise((r) => setTimeout(r, 120));
  }
  return results;
}

function deepMerge(base, overlay) {
  const out = { ...base };
  for (const [key, value] of Object.entries(overlay)) {
    if (isPlainObject(value) && isPlainObject(out[key])) out[key] = deepMerge(out[key], value);
    else out[key] = value;
  }
  return out;
}

const bundle = JSON.parse(fs.readFileSync(bundlePath, "utf8"));
const leaves = flattenLeaves(bundle);
console.log(`Bundle: ${leaves.length} strings`);

for (const target of TARGETS) {
  console.log(`\n=== ${target} ===`);
  const translated = await translateBatch(leaves, target);
  const overlay = unflattenLeaves(translated);
  const localePath = path.join(messagesDir, `${target}.json`);
  const current = JSON.parse(fs.readFileSync(localePath, "utf8"));
  const merged = deepMerge(current, overlay);
  fs.writeFileSync(localePath, JSON.stringify(merged, null, 2) + "\n");
  console.log(`  wrote ${target}.json`);
}

console.log("\nDone.");
