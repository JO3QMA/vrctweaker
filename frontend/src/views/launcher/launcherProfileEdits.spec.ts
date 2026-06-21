import { describe, it, expect } from "vitest";
import type { LaunchArgsParsedDTO } from "../../wails/app";
import { PRIORITY_OMIT } from "../../wails/app";
import {
  defaultValueOptionsEnabled,
  hasAdvancedLaunchOptionsEnabled,
  launchProfileEditsEqual,
  syncValueOptionsEnabled,
  type LaunchProfileEditSnapshot,
} from "./launcherProfileEdits";

const emptyArgs = (): LaunchArgsParsedDTO => ({
  noVr: false,
  screenMode: "",
  screenWidth: 0,
  screenHeight: 0,
  fps: 0,
  skipRegistry: false,
  processPriority: PRIORITY_OMIT,
  mainThreadPriority: PRIORITY_OMIT,
  monitor: 0,
  profile: -1,
  enableDebugGui: false,
  enableSDKLogLevels: false,
  enableUdonDebugLogging: false,
  midi: "",
  watchWorlds: false,
  watchAvatars: false,
  ignoreTrackers: "",
  videoDecoding: "",
  disableAMDStutterWorkaround: false,
  osc: "",
  affinity: "",
  enforceWorldServerChecks: false,
  custom: "",
});

function snapshot(
  overrides: Partial<LaunchProfileEditSnapshot> = {},
): LaunchProfileEditSnapshot {
  return {
    profileId: "1",
    name: "Default",
    isDefault: true,
    launchArgs: emptyArgs(),
    valueOptionsEnabled: defaultValueOptionsEnabled(),
    ...overrides,
  };
}

describe("launcherProfileEdits", () => {
  it("detects advanced options when resolution is enabled", () => {
    const args = { ...emptyArgs(), screenWidth: 1920, screenHeight: 1080 };
    const enabled = syncValueOptionsEnabled(args);
    expect(hasAdvancedLaunchOptionsEnabled(args, enabled)).toBe(true);
  });

  it("does not treat primary-only custom args as advanced", () => {
    const args = { ...emptyArgs(), custom: "-batchmode" };
    const enabled = defaultValueOptionsEnabled();
    expect(hasAdvancedLaunchOptionsEnabled(args, enabled)).toBe(false);
  });

  it("detects dirty when profile name changes", () => {
    const base = snapshot();
    const edited = snapshot({ name: "Renamed" });
    expect(launchProfileEditsEqual(base, edited)).toBe(false);
  });

  it("treats identical snapshots as clean", () => {
    const a = snapshot();
    const b = snapshot();
    expect(launchProfileEditsEqual(a, b)).toBe(true);
  });
});
