/** VRChat launcher `-screen-width` / `-screen-height` limits (GUI). */
export const VRCHAT_LAUNCHER_SCREEN_RES = {
  WIDTH_MIN: 1280,
  WIDTH_MAX: 7680,
  HEIGHT_MIN: 720,
  HEIGHT_MAX: 4320,
} as const;

/** VRChat config.json camera resolution limits (GUI). */
export const VRCHAT_CAMERA_RES = {
  WIDTH_MIN: 1280,
  WIDTH_MAX: 7680,
  HEIGHT_MIN: 720,
  HEIGHT_MAX: 4320,
} as const;

/** VRChat config.json screenshot resolution limits (GUI). */
export const VRCHAT_SCREENSHOT_RES = {
  WIDTH_MIN: 1280,
  WIDTH_MAX: 3840,
  HEIGHT_MIN: 720,
  HEIGHT_MAX: 2160,
} as const;
