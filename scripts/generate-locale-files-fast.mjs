/**
 * Fast parallel locale generator (MyMemory API, concurrency pool).
 */
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const messagesDir = path.join(__dirname, "../frontend/src/messages");
const en = JSON.parse(fs.readFileSync(path.join(messagesDir, "en.json"), "utf8"));

const TARGETS = { de: "de", es: "es", fr: "fr", zh: "zh-CN" };
const CONCURRENCY = 12;
const cachePath = path.join(__dirname, ".translation-cache.json");
let cache = fs.existsSync(cachePath) ? JSON.parse(fs.readFileSync(cachePath, "utf8")) : {};

function flatten(obj, prefix = "") {
  const out = [];
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? `${prefix}.${k}` : k;
    if (typeof v === "string") out.push([key, v]);
    else out.push(...flatten(v, key));
  }
  return out;
}

function unflatten(entries) {
  const out = {};
  for (const [key, value] of entries) {
    const parts = key.split(".");
    let cur = out;
    for (let i = 0; i < parts.length - 1; i++) {
      if (!cur[parts[i]]) cur[parts[i]] = {};
      cur = cur[parts[i]];
    }
    cur[parts[parts.length - 1]] = value;
  }
  return out;
}

async function translateText(text, targetLang) {
  if (!text?.trim()) return text;
  const cacheKey = `${targetLang}::${text}`;
  if (cache[cacheKey]) return cache[cacheKey];
  const url = `https://api.mymemory.translated.net/get?q=${encodeURIComponent(text)}&langpair=en|${targetLang}`;
  const res = await fetch(url);
  const data = await res.json();
  const translated = data?.responseData?.translatedText ?? text;
  cache[cacheKey] = translated;
  return translated;
}

async function poolMap(items, worker) {
  const results = new Array(items.length);
  let idx = 0;
  async function run() {
    while (idx < items.length) {
      const i = idx++;
      results[i] = await worker(items[i], i);
    }
  }
  await Promise.all(Array.from({ length: CONCURRENCY }, run));
  return results;
}

async function main() {
  const flat = flatten(en);
  console.log(`Source keys: ${flat.length}`);

  for (const [fileLocale, apiLang] of Object.entries(TARGETS)) {
    console.log(`\n==> ${fileLocale}`);
    const translated = await poolMap(flat, async ([key, text]) => {
      const value = await translateText(text, apiLang);
      return [key, value];
    });
    fs.writeFileSync(
      path.join(messagesDir, `${fileLocale}.json`),
      JSON.stringify(unflatten(translated), null, 2) + "\n",
      "utf8"
    );
    fs.writeFileSync(cachePath, JSON.stringify(cache, null, 2), "utf8");
    console.log(`Wrote ${fileLocale}.json`);
  }
  console.log("Done.");
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
