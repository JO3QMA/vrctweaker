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

const elementPlusTableCss = readFileSync(
  path.join(
    path.dirname(fileURLToPath(import.meta.url)),
    "../../node_modules/element-plus/theme-chalk/el-table.css",
  ),
  "utf8",
);

function extractTableEmptyOverrideRules(css: string): string {
  const blockRe = /\.el-table\s+\.el-table__empty-(?:block|text)\s*\{[^}]+\}/g;
  const rules = css.match(blockRe);
  if (!rules?.length) {
    throw new Error("vt table empty overrides not found in style.css");
  }
  return rules.join("\n");
}

describe("style.css table empty state", () => {
  it("scopes empty row overrides under .el-table (beats lazy EP chunk)", () => {
    expect(styleCss).toMatch(
      /\.el-table\s+\.el-table__empty-block\s*\{[^}]*min-height:\s*56px/,
    );
    expect(styleCss).toMatch(
      /\.el-table\s+\.el-table__empty-text\s*\{[^}]*line-height:\s*56px/,
    );
  });

  it("keeps 56px after Element Plus table CSS loads later", () => {
    const overrides = extractTableEmptyOverrideRules(styleCss);
    const epStyle = document.createElement("style");
    epStyle.textContent = elementPlusTableCss;
    const appStyle = document.createElement("style");
    appStyle.textContent = overrides;
    document.head.appendChild(epStyle);
    document.head.appendChild(appStyle);

    const table = document.createElement("div");
    table.className = "el-table";
    const emptyBlock = document.createElement("div");
    emptyBlock.className = "el-table__empty-block";
    const emptyText = document.createElement("span");
    emptyText.className = "el-table__empty-text";
    emptyText.textContent = "—";
    emptyBlock.appendChild(emptyText);
    table.appendChild(emptyBlock);
    document.body.appendChild(table);

    expect(getComputedStyle(emptyBlock).minHeight).toBe("56px");
    expect(getComputedStyle(emptyText).lineHeight).toBe("56px");

    document.body.removeChild(table);
    document.head.removeChild(appStyle);
    document.head.removeChild(epStyle);
  });
});
