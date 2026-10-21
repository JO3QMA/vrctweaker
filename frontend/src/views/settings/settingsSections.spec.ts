import { describe, expect, it } from "vitest";
import {
  settingsSectionFromPath,
  settingsChildRoutes,
} from "./settingsSections";

describe("settingsSectionFromPath", () => {
  it("maps each settings subpath", () => {
    expect(settingsSectionFromPath("/settings/account")).toBe("account");
    expect(settingsSectionFromPath("/settings/general")).toBe("general");
    expect(settingsSectionFromPath("/settings/data")).toBe("data");
    expect(settingsSectionFromPath("/settings/about")).toBe("about");
    expect(settingsSectionFromPath("/settings/reset")).toBe("reset");
  });

  it("defaults unknown paths to account", () => {
    expect(settingsSectionFromPath("/settings")).toBe("account");
    expect(settingsSectionFromPath("/")).toBe("account");
    expect(settingsSectionFromPath("/settings/unknown")).toBe("account");
  });
});

describe("settingsChildRoutes", () => {
  it("redirects the settings index to account and names each section", () => {
    const routes = settingsChildRoutes();
    expect(routes[0]).toMatchObject({
      path: "",
      redirect: { name: "settings-account" },
    });
    expect(routes.slice(1).map((route) => route.name)).toEqual([
      "settings-account",
      "settings-general",
      "settings-data",
      "settings-about",
      "settings-reset",
    ]);
  });
});
