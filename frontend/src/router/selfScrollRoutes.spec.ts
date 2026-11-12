import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

/**
 * main.ts の `meta.selfScroll` と E2E の SELF_SCROLL_ROUTE_NAMES を同期するガード。
 */
describe("selfScroll route meta", () => {
  it("marks only expected top-level routes in main.ts", () => {
    const mainTs = readFileSync(resolve(__dirname, "../main.ts"), "utf8");
    const expected = [
      "routes.gallery",
      "routes.activity",
      "routes.me",
      "routes.friends",
      "routes.user",
      "routes.automation",
      "routes.settings",
    ];
    for (const titleKey of expected) {
      const pattern = new RegExp(
        `titleKey:\\s*"${titleKey.replace(".", "\\.")}"[^\\n]*selfScroll:\\s*true`,
      );
      expect(mainTs, titleKey).toMatch(pattern);
    }
    const count = (mainTs.match(/selfScroll:\s*true/g) ?? []).length;
    expect(count).toBe(expected.length);
  });
});
