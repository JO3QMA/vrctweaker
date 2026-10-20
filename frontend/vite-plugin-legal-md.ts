import fs from "node:fs";
import path from "node:path";
import type { Plugin } from "vite";

const PREFIX = "virtual:legal-md/";

/** Bundles `docs/legal/*.md` from the repo root without importing outside Vite root. */
export function legalMarkdownPlugin(repoRoot: string): Plugin {
  const legalDir = path.join(repoRoot, "docs", "legal");

  return {
    name: "legal-markdown-raw",
    resolveId(source) {
      if (source.startsWith(PREFIX)) {
        return `\0${source}`;
      }
      return undefined;
    },
    load(id) {
      if (!id.startsWith(`\0${PREFIX}`)) {
        return undefined;
      }
      const fileName = id.slice(`\0${PREFIX}`.length);
      if (
        fileName.includes("..") ||
        fileName.includes("/") ||
        fileName.includes("\\")
      ) {
        return undefined;
      }
      const fullPath = path.join(legalDir, fileName);
      const text = fs.readFileSync(fullPath, "utf8");
      return `export default ${JSON.stringify(text)};`;
    },
  };
}

export function legalMarkdownImport(fileName: string): string {
  return `${PREFIX}${fileName}`;
}
