import { describe, it, expect } from "vitest";
import { NAV_MAIN_MENU_ITEMS, NAV_SETTINGS_MENU_ITEM } from "../navMenuItems";

describe("navMenuItems", () => {
  it("defines unique paths for main nav items", () => {
    const paths = NAV_MAIN_MENU_ITEMS.map((item) => item.path);
    expect(new Set(paths).size).toBe(paths.length);
  });

  it("marks video as Windows-only", () => {
    const video = NAV_MAIN_MENU_ITEMS.find((item) => item.path === "/video");
    expect(video?.windowsOnly).toBe(true);
  });

  it("keeps settings in a separate footer item", () => {
    expect(NAV_SETTINGS_MENU_ITEM.path).toBe("/settings");
    expect(
      NAV_MAIN_MENU_ITEMS.some(
        (item) => item.path === NAV_SETTINGS_MENU_ITEM.path,
      ),
    ).toBe(false);
  });
});
