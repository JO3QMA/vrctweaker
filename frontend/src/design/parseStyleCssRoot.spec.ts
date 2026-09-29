import { describe, expect, it } from "vitest";
import {
  parseCssPxLength,
  parseStyleCssRootDeclarations,
  resolveCssVarInMap,
  rootNumericScalePx,
} from "./parseStyleCssRoot";

describe("parseStyleCssRoot", () => {
  it("parses declarations from :root", () => {
    const vars = parseStyleCssRootDeclarations(`
      :root {
        /* comment */
        --color-brand: #abc;
        --space-8: 8px;
      }
    `);
    expect(vars["--color-brand"]).toBe("#abc");
    expect(vars["--space-8"]).toBe("8px");
  });

  it("resolves var() chains in a parsed map", () => {
    const vars = parseStyleCssRootDeclarations(`
      :root {
        --space-4: 4px;
        --space-inline-tight: var(--space-4);
      }
    `);
    expect(resolveCssVarInMap(vars, "--space-inline-tight")).toBe("4px");
  });

  it("parses px lengths", () => {
    expect(parseCssPxLength("16px")).toBe(16);
    expect(parseCssPxLength("calc(1px)")).toBeNull();
  });

  it("parses :root when a comment inside the block contains }", () => {
    const vars = parseStyleCssRootDeclarations(`
      :root {
        /* legacy } marker */
        --color-brand: #abc;
      }
    `);
    expect(vars["--color-brand"]).toBe("#abc");
  });

  it("collects numeric scale steps from a parsed map", () => {
    const vars = parseStyleCssRootDeclarations(`
      :root {
        --space-4: 4px;
        --space-inline-tight: var(--space-4);
        --space-16: 16px;
      }
    `);
    expect(rootNumericScalePx(vars, "--space-")).toEqual([4, 16]);
  });
});
