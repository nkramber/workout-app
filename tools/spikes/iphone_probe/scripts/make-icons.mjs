// Renders the PNG icons of the probe from public/favicon.svg. It runs by
// hand with `npm run icons`, and its output is committed, so no build
// needs a browser. Chromium comes from Playwright, which the browser
// tests install.
import { readFileSync, writeFileSync } from "node:fs";
import path from "node:path";

import { chromium } from "@playwright/test";

const pub = path.join(import.meta.dirname, "..", "public");
const mark = readFileSync(path.join(pub, "favicon.svg"), "utf8");

// Android crops a maskable icon to a circle, so the mark sits in the
// middle 80 percent of the square.
const maskable = mark
  .replace('viewBox="0 0 32 32"', 'viewBox="-4 -4 40 40"')
  .replace('<rect width="32" height="32" rx="7"', '<rect x="-4" y="-4" width="40" height="40" rx="0"');

const icons = [
  { file: "icon-192.png", size: 192, svg: mark },
  { file: "icon-512.png", size: 512, svg: mark },
  { file: "icon-maskable-512.png", size: 512, svg: maskable },
  // iOS reads no SVG for a Home Screen icon, and 180 is its size.
  { file: "apple-touch-icon.png", size: 180, svg: mark },
];

const browser = await chromium.launch();
for (const { file, size, svg } of icons) {
  const page = await browser.newPage({ viewport: { width: size, height: size }, deviceScaleFactor: 1 });
  const data = `data:image/svg+xml;base64,${Buffer.from(svg).toString("base64")}`;
  await page.setContent(
    `<style>html,body{margin:0;background:transparent}img{display:block;width:${size}px;height:${size}px}</style><img src="${data}">`,
  );
  await page.locator("img").waitFor();
  writeFileSync(path.join(pub, file), await page.screenshot({ omitBackground: true }));
  await page.close();
}
await browser.close();
console.log(`wrote ${icons.length} icons to ${pub}`);
