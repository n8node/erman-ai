/**
 * Writes calculator template overlay JSON files from embedded translations.
 * Run: node scripts/write-template-overlays.mjs
 */
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const enStrings = JSON.parse(
  fs.readFileSync(path.join(__dirname, "template-en-strings.json"), "utf8")
);
const translations = JSON.parse(
  fs.readFileSync(path.join(__dirname, "template-translations.json"), "utf8")
);

const outDir = path.join(__dirname, "../frontend/src/lib/calculator-templates/overlays");

for (const locale of ["de", "es", "fr", "zh"]) {
  const map = translations[locale];
  if (!map) throw new Error(`Missing locale ${locale}`);
  const overlay = {};
  for (const en of enStrings) {
    if (!map[en]) {
      throw new Error(`Missing ${locale} translation for: ${en}`);
    }
    overlay[en] = map[en];
  }
  fs.writeFileSync(path.join(outDir, `${locale}.json`), JSON.stringify(overlay, null, 2) + "\n");
  console.log(`Wrote ${locale}.json (${Object.keys(overlay).length} keys)`);
}

console.log("Done.");
