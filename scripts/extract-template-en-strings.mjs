import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const content = fs.readFileSync(
  path.join(__dirname, "../frontend/src/lib/calculator-templates/data.ts"),
  "utf8"
);
const en = new Set();
for (const m of content.matchAll(/L\("[^"]+",\s*"([^"]+)"\)/g)) en.add(m[1]);
for (const m of content.matchAll(/S\("[^"]+",\s*"([^"]+)",\s*"[^"]+",\s*"([^"]+)"/g)) {
  en.add(m[1]);
  en.add(m[2]);
}
const sorted = [...en].sort();
fs.writeFileSync(
  path.join(__dirname, "template-en-strings.json"),
  JSON.stringify(sorted, null, 2) + "\n"
);
console.log(sorted.length, "strings written");
