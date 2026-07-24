import type { LaunchArgsParsedDTO } from "../../wails/app";
import { PRIORITY_OMIT } from "../../wails/app";

export interface ValueOptionsEnabled {
  resolution: boolean;
  monitor: boolean;
  fps: boolean;
  processPriority: boolean;
  mainThreadPriority: boolean;
  profile: boolean;
  midi: boolean;
  ignoreTrackers: boolean;
  osc: boolean;
  affinity: boolean;
}

export interface LaunchProfileEditSnapshot {
  profileId: string;
  name: string;
  isDefault: boolean;
  launchArgs: LaunchArgsParsedDTO;
  valueOptionsEnabled: ValueOptionsEnabled;
}

export const LAUNCHER_SIDEBAR_OPEN_KEY = "vrctweaker.launcher.sidebarOpen";

export function defaultValueOptionsEnabled(): ValueOptionsEnabled {
  return {
    resolution: false,
    monitor: false,
    fps: false,
    processPriority: false,
    mainThreadPriority: false,
    profile: false,
    midi: false,
    ignoreTrackers: false,
    osc: false,
    affinity: false,
  };
}

export function syncValueOptionsEnabled(
  a: LaunchArgsParsedDTO,
): ValueOptionsEnabled {
  return {
    resolution: a.screenWidth > 0 || a.screenHeight > 0,
    monitor: a.monitor >= 1,
    fps: a.fps > 0,
    processPriority:
      a.processPriority !== PRIORITY_OMIT &&
      a.processPriority >= -2 &&
      a.processPriority <= 2,
    mainThreadPriority:
      a.mainThreadPriority !== PRIORITY_OMIT &&
      a.mainThreadPriority >= -2 &&
      a.mainThreadPriority <= 2,
    profile: a.profile >= 0,
    midi: a.midi !== "",
    ignoreTrackers: a.ignoreTrackers !== "",
    osc: a.osc !== "",
    affinity: a.affinity !== "",
  };
}

export function hasAdvancedLaunchOptionsEnabled(
  args: LaunchArgsParsedDTO,
  enabled: ValueOptionsEnabled,
): boolean {
  if (enabled.resolution && (args.screenWidth > 0 || args.screenHeight > 0)) {
    return true;
  }
  if (enabled.monitor && args.monitor >= 1) return true;
  if (enabled.fps && args.fps > 0) return true;
  if (args.skipRegistry) return true;
  if (
    enabled.processPriority &&
    args.processPriority !== PRIORITY_OMIT &&
    args.processPriority >= -2 &&
    args.processPriority <= 2
  ) {
    return true;
  }
  if (
    enabled.mainThreadPriority &&
    args.mainThreadPriority !== PRIORITY_OMIT &&
    args.mainThreadPriority >= -2 &&
    args.mainThreadPriority <= 2
  ) {
    return true;
  }
  if (enabled.profile && args.profile >= 0) return true;
  if (args.enableDebugGui) return true;
  if (args.enableSDKLogLevels) return true;
  if (args.enableUdonDebugLogging) return true;
  if (args.watchWorlds) return true;
  if (args.watchAvatars) return true;
  if (args.enforceWorldServerChecks) return true;
  if (enabled.midi && args.midi !== "") return true;
  if (enabled.ignoreTrackers && args.ignoreTrackers !== "") return true;
  if (args.videoDecoding !== "") return true;
  if (args.disableAMDStutterWorkaround) return true;
  if (enabled.osc && args.osc !== "") return true;
  if (enabled.affinity && args.affinity !== "") return true;
  return false;
}

function launchArgsEqual(
  a: LaunchArgsParsedDTO,
  b: LaunchArgsParsedDTO,
): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

function valueOptionsEqual(a: ValueOptionsEnabled, b: ValueOptionsEnabled) {
  return JSON.stringify(a) === JSON.stringify(b);
}

export function launchProfileEditsEqual(
  a: LaunchProfileEditSnapshot,
  b: LaunchProfileEditSnapshot,
): boolean {
  return (
    a.profileId === b.profileId &&
    a.name === b.name &&
    a.isDefault === b.isDefault &&
    launchArgsEqual(a.launchArgs, b.launchArgs) &&
    valueOptionsEqual(a.valueOptionsEnabled, b.valueOptionsEnabled)
  );
}

/** True only when the current editor state diverges from the saved snapshot for the same profile. */
export function isLaunchProfileEditDirty(
  saved: LaunchProfileEditSnapshot | null,
  current: LaunchProfileEditSnapshot | null,
): boolean {
  if (!saved || !current) return false;
  if (saved.profileId !== current.profileId) return false;
  return !launchProfileEditsEqual(saved, current);
}

export function readSidebarOpenPreference(): boolean {
  if (typeof localStorage === "undefined") return true;
  const raw = localStorage.getItem(LAUNCHER_SIDEBAR_OPEN_KEY);
  if (raw === null) return true;
  return raw === "1";
}

export function writeSidebarOpenPreference(open: boolean): void {
  if (typeof localStorage === "undefined") return;
  localStorage.setItem(LAUNCHER_SIDEBAR_OPEN_KEY, open ? "1" : "0");
}

/** Next free default display name: base, then `base 2`, `base 3`, … */
export function nextDefaultLaunchProfileName(
  base: string,
  existingNames: readonly string[],
): string {
  const taken = new Set(existingNames);
  if (!taken.has(base)) return base;
  for (let n = 2; ; n++) {
    const candidate = `${base} ${n}`;
    if (!taken.has(candidate)) return candidate;
  }
}
