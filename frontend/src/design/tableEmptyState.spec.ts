import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const styleCss = readFileSync(
  path.join(
    path.dirname(fileURLToPath(import.meta.url)),
    "../assets/style.css",
  ),
  "utf8",
);

describe("style.css table empty state", () => {
  it("uses 56px height for empty rows (8px grid)", () => {
    expect(styleCss).toMatch(
      /\.el-table__empty-block\s*\{[^}]*min-height:\s*56px/,
    );
    expect(styleCss).toMatch(
      /\.el-table__empty-text\s*\{[^}]*line-height:\s*56px/,
    );
  });
});
