import { readFileSync } from "node:fs";
import path from "node:path";
import { describe, expect, it } from "vitest";
import { parseStyleSheet } from "./cssCustomProperties";
import { WCAG_AA_NORMAL_TEXT, wcagContrastRatio } from "./colorContrast";

const DARK_ELEVATED = "#243447";
const LIGHT_BASE = "#ffffff";
const LIGHT_ELEVATED = "#f5f7fa";

function hexFromStyleCss(
  name: string,
  selector: ":root" | "html.light",
): string {
  const styleCss = readFileSync(
    path.resolve(process.cwd(), "src/assets/style.css"),
    "utf8",
  );
  const row = parseStyleSheet(styleCss).customProperties.find(
    (prop) => prop.name === name && prop.selector === selector,
  );
  expect(row?.value).toMatch(/^#[0-9a-f]{6}$/i);
  return row!.value;
}

describe("link color contrast (WCAG AA normal text)", () => {
  it("keeps dark-theme link tokens on elevated surface", () => {
    const link = hexFromStyleCss("--color-link", ":root");
    const hover = hexFromStyleCss("--color-link-hover", ":root");
    expect(wcagContrastRatio(link, DARK_ELEVATED)).toBeGreaterThanOrEqual(
      WCAG_AA_NORMAL_TEXT,
    );
    expect(wcagContrastRatio(hover, DARK_ELEVATED)).toBeGreaterThanOrEqual(
      WCAG_AA_NORMAL_TEXT,
    );
    expect(wcagContrastRatio("#5c9ce6", DARK_ELEVATED)).toBeLessThan(
      WCAG_AA_NORMAL_TEXT,
    );
  });

  it("keeps reserved light-theme link tokens on light surfaces", () => {
    const link = hexFromStyleCss("--color-link", "html.light");
    const hover = hexFromStyleCss("--color-link-hover", "html.light");
    expect(wcagContrastRatio(link, LIGHT_BASE)).toBeGreaterThanOrEqual(
      WCAG_AA_NORMAL_TEXT,
    );
    expect(wcagContrastRatio(link, LIGHT_ELEVATED)).toBeGreaterThanOrEqual(
      WCAG_AA_NORMAL_TEXT,
    );
    expect(wcagContrastRatio(hover, LIGHT_BASE)).toBeGreaterThanOrEqual(
      WCAG_AA_NORMAL_TEXT,
    );
  });
});
