/**
 * Must stay in sync with MsgAssetCache* in internal/usecase/vrchat_asset_cache.go.
 * Go sentinel Error() strings are matched exactly in ConfigView.
 */
export const AssetCacheErr = {
  VRCHAT_RUNNING: "vrchat is running",
  VOLUME_ROOT: "cache path is volume root",
  NOT_DIRECTORY: "cache path is not a directory",
  PATH_MISSING: "cache path does not exist",
  EQUALS_PICTURE_FOLDER: "cache path equals picture folder",
  EQUALS_VRCHAT_DATA_DIR: "cache path equals vrchat data directory",
  EMPTY_PATH: "cache path is empty",
  REMOVE_FAILED: "cache remove failed",
  FAILED: "asset cache clear failed",
} as const;
