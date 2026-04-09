import { describe, it, expect } from "vitest";
import { appRoutes } from "./routes";

describe("appRoutes", () => {
  it("defines the expected top-level routes for Wails (no lazy chunks)", () => {
    expect(appRoutes.length).toBe(11);
    const paths = appRoutes.map((r) => r.path);
    expect(paths).toContain("/");
    expect(paths).toContain("/activity/encounter-history");
    expect(paths).toContain("/settings");
  });

  it("uses object components, not dynamic import loaders", () => {
    for (const r of appRoutes) {
      expect(r.component).toBeDefined();
      expect(typeof r.component).toBe("object");
    }
  });
});
