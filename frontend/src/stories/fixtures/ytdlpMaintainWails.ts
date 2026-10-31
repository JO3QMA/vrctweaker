import type { Decorator } from "@storybook/vue3-vite";
import { withWailsApp } from "../wailsDecorator";
import type {
  YTDLPMaintainActionResultDTO,
  YTDLPMaintainStatusDTO,
} from "../../wails/app";
import { YTDLP_MAINTAIN_OUTCOME } from "../../utils/ytdlpMaintainOutcome";

export const ytdlpBaseStatus: YTDLPMaintainStatusDTO = {
  supported: true,
  unsupportedReason: "",
  maintainDesired: true,
  riskAcknowledged: true,
  effectiveOfficial: true,
  cachePresent: true,
  cacheVersion: "2026.07.04",
  toolsPath:
    "C:\\Users\\example\\AppData\\LocalLow\\VRChat\\VRChat\\Tools\\yt-dlp.exe",
  cachePath:
    "C:\\Users\\example\\AppData\\Local\\vrchat-tweaker\\ytdlp\\yt-dlp.exe",
  pendingError: "",
  latestVersion: "2026.07.04",
  latestTag: "2026.07.04",
  latestDownloadUrl: "https://example.com/yt-dlp.exe",
  latestError: "",
  latestCheckedAt: "2026-10-31T00:14:00.000Z",
};

export function upToDateStatus(
  overrides: Partial<YTDLPMaintainStatusDTO> = {},
): YTDLPMaintainStatusDTO {
  return { ...ytdlpBaseStatus, ...overrides };
}

export function updateAvailableStatus(
  overrides: Partial<YTDLPMaintainStatusDTO> = {},
): YTDLPMaintainStatusDTO {
  return {
    ...ytdlpBaseStatus,
    cacheVersion: "2026.07.04",
    latestVersion: "2026.10.26",
    latestTag: "2026.10.26",
    ...overrides,
  };
}

export function notDownloadedStatus(
  overrides: Partial<YTDLPMaintainStatusDTO> = {},
): YTDLPMaintainStatusDTO {
  return {
    ...ytdlpBaseStatus,
    maintainDesired: false,
    effectiveOfficial: false,
    cachePresent: false,
    cacheVersion: "",
    latestVersion: "2026.10.26",
    latestCheckedAt: "2026-10-31T09:14:00.000Z",
    ...overrides,
  };
}

export function backgroundCheckFailedStatus(
  overrides: Partial<YTDLPMaintainStatusDTO> = {},
): YTDLPMaintainStatusDTO {
  return {
    ...ytdlpBaseStatus,
    latestError:
      'github api: 403 Forbidden: {"message":"API rate limit exceeded"}',
    latestCheckedAt: "2026-10-31T09:14:00.000Z",
    ...overrides,
  };
}

export function mockCheckAndUpdateResult(
  status: YTDLPMaintainStatusDTO,
  outcome: YTDLPMaintainActionResultDTO["outcome"] = YTDLP_MAINTAIN_OUTCOME.unchanged,
  extra?: Partial<YTDLPMaintainActionResultDTO>,
): YTDLPMaintainActionResultDTO {
  return {
    status,
    outcome,
    previousVersion: status.cacheVersion,
    newVersion: status.latestVersion,
    detailError: "",
    ...extra,
  };
}

export function ytdlpWailsMocks(status: YTDLPMaintainStatusDTO) {
  return {
    GetYTDLPMaintainStatus: () => Promise.resolve(status),
    RefreshYTDLPLatestRelease: () => Promise.resolve(status),
    AcknowledgeYTDLPToolsReplaceRisk: () => Promise.resolve(),
    SetYTDLPToolsReplaceMaintain: () => Promise.resolve(),
    CheckAndUpdateOfficialYTDLPCache: () =>
      Promise.resolve(mockCheckAndUpdateResult(status)),
    OpenYTDLPCacheFolder: () => Promise.resolve(),
    OpenYTDLPToolsFolder: () => Promise.resolve(),
    RuntimeIsWindows: () => Promise.resolve(true),
  };
}

export function wailsForYtdlpStatus(status: YTDLPMaintainStatusDTO): Decorator {
  return withWailsApp(ytdlpWailsMocks(status));
}
