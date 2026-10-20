import { readFileSync } from "node:fs";
import path from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import {
  elementForCssSelector,
  parseStyleSheet,
  readComputedCustomProperty,
  resolveCustomProperties,
  usedLengthIfComputed,
  valueReferencesVar,
} from "./cssCustomProperties";

const fixture = `
/* Banner */
:root {
  /* Brand: リンク・フォーカス・Primary button の塗り */
  --color-brand: #5c9ce6; /* リンク・フォーカス・Primary button 塗り */
  --color-brand-hover: #7ab3f0;

  /* Neutral: 表面・文字・枠線 */
  --color-text-primary: #e8eaed; /* 本文 */
  --color-danger-skip: #000;

  /* Semantic */
  --color-danger: #e57373; /* 危険・破壊 */
  --color-status-unknown: var(--color-text-secondary);
}

/* Theme block */
html.dark {
  --el-color-primary: var(--color-brand);
  --el-font-size-base: var(--font-size-14);
  --el-font-size-small: 13px; /* スケール外の derivative */
  --el-font-size-extra-small: var(--font-size-12);
}

/* Text style catalog (ADR 0020) */
.text-h1 {
  font-size: var(--font-size-h1);
  line-height: var(--line-height-tight);
  font-weight: var(--font-weight-600);
}
`;

describe("parseStyleSheet", () => {
  const parsed = parseStyleSheet(fixture);

  it("keeps group order and does not let an inline note swallow the next token", () => {
    const names = parsed.customProperties.map((prop) => prop.name);
    expect(names).toEqual([
      "--color-brand",
      "--color-brand-hover",
      "--color-text-primary",
      "--color-danger-skip",
      "--color-danger",
      "--color-status-unknown",
      "--el-color-primary",
      "--el-font-size-base",
      "--el-font-size-small",
      "--el-font-size-extra-small",
    ]);
    expect(parsed.customProperties[0]?.group).toBe(
      "Brand: リンク・フォーカス・Primary button の塗り",
    );
    expect(parsed.customProperties[1]?.group).toBe(
      "Brand: リンク・フォーカス・Primary button の塗り",
    );
    expect(parsed.customProperties[2]?.group).toBe("Neutral: 表面・文字・枠線");
    expect(parsed.customProperties[4]?.description).toBe("危険・破壊");
    expect(parsed.customProperties[8]?.description).toContain("derivative");
    expect(parsed.customProperties[9]?.group).not.toContain("derivative");
    expect(parsed.customProperties[9]?.description).toBe("");
  });

  it("strips box-drawing rules from section comments and tightens calc()", () => {
    const sheet = parseStyleSheet(`
      /* ─── Element Plus dark theme CSS variable overrides ───────────── */
      html.dark {
        --font-size-h1: calc(
          var(--font-size-14) * 1.4
        );
      }
    `);
    expect(sheet.customProperties[0]?.group).toBe(
      "Element Plus dark theme CSS variable overrides",
    );
    expect(sheet.customProperties[0]?.value).toBe(
      "calc(var(--font-size-14) * 1.4)",
    );
  });

  it("collapses whitespace outside quotes and keeps it inside them", () => {
    const sheet = parseStyleSheet(`
      :root {
        --font-family-ui:
          "Segoe  UI",   sans-serif;
        --quoted-single: 'a  b',   sans-serif;
      }
    `);
    expect(
      sheet.customProperties.find((prop) => prop.name === "--font-family-ui")
        ?.value,
    ).toBe('"Segoe  UI", sans-serif');
    expect(
      sheet.customProperties.find((prop) => prop.name === "--quoted-single")
        ?.value,
    ).toBe("'a  b', sans-serif");
  });

  it("keeps var() source text and the declaring selector", () => {
    const unknown = parsed.customProperties.find(
      (prop) => prop.name === "--color-status-unknown",
    );
    expect(unknown?.value).toBe("var(--color-text-secondary)");
    expect(unknown?.selector).toBe(":root");
    const primary = parsed.customProperties.find(
      (prop) => prop.name === "--el-color-primary",
    );
    expect(primary?.selector).toBe("html.dark");
    expect(primary?.value).toBe("var(--color-brand)");
    expect(valueReferencesVar(primary?.value ?? "")).toBe(true);
    expect(valueReferencesVar("#5c9ce6")).toBe(false);
  });

  it("keeps class declaration order for text styles", () => {
    const text = parsed.declarations.filter(
      (decl) => decl.selector === ".text-h1",
    );
    expect(text.map((decl) => decl.property)).toEqual([
      "font-size",
      "line-height",
      "font-weight",
    ]);
    expect(text[0]?.value).toBe("var(--font-size-h1)");
    expect(text[0]?.group).toBe("Text style catalog (ADR 0020)");
  });
});

describe("resolveCustomProperties", () => {
  afterEach(() => {
    document.documentElement.className = "";
    document.documentElement.removeAttribute("style");
  });

  it("reads a themed custom property from the matching element", () => {
    expect(elementForCssSelector(document, "html.dark")).toBeNull();
    document.documentElement.classList.add("dark");
    document.documentElement.style.setProperty("--color-brand", "#5c9ce6");
    document.documentElement.style.setProperty("--el-color-primary", "#5c9ce6");
    expect(elementForCssSelector(document, "html.dark")).toBe(
      document.documentElement,
    );
    expect(elementForCssSelector(document, ":root")).toBe(
      document.documentElement,
    );
    expect(
      readComputedCustomProperty(document, "html.dark", "--el-color-primary"),
    ).toBe("#5c9ce6");
    expect(
      readComputedCustomProperty(
        document,
        "html.missing",
        "--el-color-primary",
      ),
    ).toBe("");

    const [row] = resolveCustomProperties(document, [
      {
        selector: "html.dark",
        name: "--el-color-primary",
        value: "var(--color-brand)",
        group: "mapping",
        description: "",
      },
    ]);
    expect(row?.referencesVar).toBe(true);
    expect(row?.resolved).toBe("#5c9ce6");
  });

  it("replaces an unevaluated calc() with the used px length", () => {
    expect(usedLengthIfComputed("calc(14px * 1.4)", "19.6px")).toBe("19.6px");
    expect(usedLengthIfComputed("#9aa0a6", "19.6px")).toBe("#9aa0a6");
    expect(
      usedLengthIfComputed("calc(14px * 1.4)", "var(--font-size-h1)"),
    ).toBe("calc(14px * 1.4)");
  });
});

describe("style.css catalog", () => {
  // Vitest's CSS pipeline drops `?raw` imports. Read the stylesheet from disk.
  const styleCss = readFileSync(
    path.resolve(process.cwd(), "src/assets/style.css"),
    "utf8",
  );
  const parsed = parseStyleSheet(styleCss);
  const root = parsed.customProperties.filter(
    (prop) => prop.selector === ":root",
  );
  const dark = parsed.customProperties.filter(
    (prop) => prop.selector === "html.dark",
  );

  function groupsFor(
    props: { name: string; group: string }[],
    prefix: string,
  ): string[] {
    const groups: string[] = [];
    for (const prop of props) {
      if (!prop.name.startsWith(prefix)) continue;
      const last = groups[groups.length - 1];
      if (last !== prop.group) groups.push(prop.group);
    }
    return groups;
  }

  it("keeps color groups, usage notes, and var() aliases in source order", () => {
    expect(groupsFor(root, "--color-")).toEqual([
      "Brand: Main accent. Primary button fill and links use --color-brand.",
      "Neutral: Surfaces, text, and border. Prefer these over legacy --bg-* / --text-* aliases.",
      "Semantic: タグ・メッセージ・検証のフィードバック。Server status と Presence には使わない",
      "Server status (domain): Dashboard のサーバー状態。Semantic には流用しない",
      "Presence (domain): Presence change の色。背景と枠は別トークン",
    ]);
    expect(
      root.find((prop) => prop.name === "--color-text-primary")?.description,
    ).toBe("本文");
    expect(
      root.find((prop) => prop.name === "--color-danger")?.description,
    ).toBe("危険・破壊");
    expect(
      root.find((prop) => prop.name === "--color-status-unknown")?.value,
    ).toBe("var(--color-text-secondary)");
    const legacy = root.filter((prop) => prop.group.startsWith("Legacy"));
    expect(legacy.map((prop) => prop.name)).toEqual([
      "--bg-primary",
      "--bg-secondary",
      "--bg-tertiary",
      "--text-primary",
      "--text-secondary",
      "--accent",
      "--accent-hover",
      "--danger",
      "--success",
      "--border",
    ]);
    expect(legacy.every((prop) => prop.value.startsWith("var(--color-"))).toBe(
      true,
    );
  });

  it("keeps spacing and typography notes that used to live in the TS catalogs", () => {
    expect(
      root.find((prop) => prop.name === "--space-form-field"),
    ).toMatchObject({
      value: "var(--space-12)",
      description: "フォーム項目間",
    });
    expect(root.find((prop) => prop.name === "--font-size-h1")?.value).toBe(
      "calc(var(--font-size-14) * 1.4)",
    );
    expect(
      dark
        .filter((prop) => prop.name.startsWith("--el-font-size-"))
        .map((prop) => prop.name),
    ).toEqual([
      "--el-font-size-extra-large",
      "--el-font-size-large",
      "--el-font-size-medium",
      "--el-font-size-base",
      "--el-font-size-extra-small",
      "--el-font-size-small",
    ]);
    expect(
      root.find((prop) => prop.name === "--line-height-normal")?.description,
    ).toBe("本文既定");
    expect(
      dark.find((prop) => prop.name === "--el-font-size-base")?.value,
    ).toBe("var(--font-size-14)");
    expect(
      dark.find((prop) => prop.name === "--el-font-size-small")?.description,
    ).toContain("derivative");
    const textH1 = parsed.declarations.filter(
      (decl) => decl.selector === ".text-h1",
    );
    expect(textH1.map((decl) => `${decl.property}:${decl.value}`)).toEqual([
      "font-size:var(--font-size-h1)",
      "line-height:var(--line-height-tight)",
      "font-weight:var(--font-weight-600)",
    ]);
    expect(textH1[0]?.group).toBe("Text style catalog (ADR 0020)");
    expect(root.find((prop) => prop.name === "--font-size-h1")?.group).toBe(
      "Font size derivative",
    );
  });
});
