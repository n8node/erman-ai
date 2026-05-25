/**
 * Generates de.json, es.json, fr.json, zh.json from en.json via MyMemory API.
 * Usage: node scripts/generate-locale-files.mjs
 */
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const messagesDir = path.join(__dirname, "../frontend/src/messages");
const en = JSON.parse(fs.readFileSync(path.join(messagesDir, "en.json"), "utf8"));

const TARGETS = {
  de: "de",
  es: "es",
  fr: "fr",
  zh: "zh-CN",
};

const cachePath = path.join(__dirname, ".translation-cache.json");
let cache = {};
if (fs.existsSync(cachePath)) {
  cache = JSON.parse(fs.readFileSync(cachePath, "utf8"));
}

function sleep(ms) {
  return new Promise((r) => setTimeout(r, ms));
}

async function translateText(text, targetLang) {
  if (!text?.trim()) return text;
  const cacheKey = `${targetLang}::${text}`;
  if (cache[cacheKey]) return cache[cacheKey];

  const url = `https://api.mymemory.translated.net/get?q=${encodeURIComponent(text)}&langpair=en|${targetLang}`;
  for (let attempt = 0; attempt < 3; attempt++) {
    try {
      const res = await fetch(url);
      const data = await res.json();
      const translated = data?.responseData?.translatedText;
      if (
        translated &&
        typeof translated === "string" &&
        !translated.includes("MYMEMORY") &&
        !translated.includes("USAGELIMITS.PHP")
      ) {
        cache[cacheKey] = translated;
        return translated;
      }
    } catch {
      await sleep(500);
    }
    await sleep(300);
  }
  return text;
}

async function walk(obj, targetLang) {
  if (typeof obj === "string") {
    await sleep(120);
    return translateText(obj, targetLang);
  }
  const out = {};
  for (const [k, v] of Object.entries(obj)) {
    out[k] = await walk(v, targetLang);
  }
  return out;
}

async function main() {
  for (const [fileLocale, apiLang] of Object.entries(TARGETS)) {
    const outPath = path.join(messagesDir, `${fileLocale}.json`);
    console.log(`Translating -> ${fileLocale} (${apiLang})`);
    const translated = await walk(en, apiLang);
    fs.writeFileSync(outPath, JSON.stringify(translated, null, 2) + "\n", "utf8");
    fs.writeFileSync(cachePath, JSON.stringify(cache, null, 2), "utf8");
    console.log(`Wrote ${outPath}`);
  }
  console.log("Done.");
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
