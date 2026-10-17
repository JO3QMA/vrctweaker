# Privacy Policy (VRChat Tweaker)

Last updated: 2026-10-14

This policy describes what data **VRChat Tweaker** (the Windows desktop app) stores on your device and which network requests it makes, based on the **actual implementation** in this repository. VRChat Tweaker is not an official product of VRChat Inc.

## 1. Provider and contact

- The app is open source under the [MIT License](../../LICENSE).
- Privacy questions: use [GitHub Issues](https://github.com/JO3QMA/vrctweaker/issues). Do not paste VRChat display names or user IDs in issue bodies.

## 2. Overview

The app does **not** implement automatic telemetry or analytics uploads to a dedicated operator server. Data is stored mainly in **local files and a SQLite database**. Network access is limited to **explicit features** in code (VRChat integration, status page, yt-dlp updates, and resources loaded by the UI).

## 3. Data stored on your device

### 3.1 Application data directory

On Windows, the app data directory is `vrchat-tweaker` under `os.UserConfigDir()` (typically `%AppData%\Roaming\vrchat-tweaker\`, see `internal/platform/paths.AppDataDir`).

| Path | Contents |
|------|----------|
| `vrchat-tweaker.db` | SQLite database (see below) |
| `tray.ico` | System tray icon |
| `.vrchat-auth-token` | **Fallback auth file** (mode 0600) when the OS credential store is unavailable, or when `VRCHAT_TWEAKER_USE_FILE_CREDENTIALS=1` |

### 3.2 SQLite (`vrchat-tweaker.db`)

Schema: `internal/infrastructure/sqlite/db.go`.

| Table / settings | Summary |
|------------------|---------|
| `app_settings` | UI language, VRChat/Steam/output_log paths, log retention days, gallery auto-enrich, close-to-tray, sleep suppression, last launch profile, activity log checkpoint (`activity_log_checkpoint`), yt-dlp settings, friends sync time (`vrchat_friends_synced_at`), status description history (`presence_description_history`), etc. |
| `launch_profiles` | Launch profile names and arguments |
| `screenshots` | Screenshot **file paths**, timestamps, world IDs, file sizes, author VRChat user IDs when available |
| `screenshot_thumbnails` | JPEG thumbnail blobs for the gallery |
| `screenshot_enrichment` | Metadata enrichment from Activity logs (instance ID, participants JSON, etc.); may embed XMP into image files |
| `play_sessions` | Play sessions from output_log |
| `user_encounters` | **Encounter records** from output_log (user IDs, display names, instances, worlds, join/leave times, log source paths) |
| `friend_avatar_observations` | Friend avatar observations from output_log |
| `video_playback_history` | Video playback attempts parsed from output_log (URLs, outcomes, resolved URLs) |
| `users_cache` | Cached user/friend profiles from the VRChat API |
| `world_info` | World IDs, display names, last visited |
| `avatar_cache` | Avatar IDs and names |
| `automation_rules` / `automation_items` | Automation configuration |

Clearing the gallery index removes DB rows only; **it does not delete image files** on disk.

### 3.3 VRChat authentication

| Location | Contents |
|----------|----------|
| Windows Credential Manager (preferred) | Service `vrchat-tweaker`, user `auth-token`: wrapped auth blob or legacy token |
| `.vrchat-auth-token` file | Fallback when keyring is unavailable |
| WebView IndexedDB `vrctwk-credentials` | Non-extractable AES-GCM **wrapping key** (`frontend/src/services/credentialCrypto.ts`), local only |

Your **password is sent only during login** to the VRChat API and is **not** stored in plaintext in the database or credential store. Sessions use an auth token and HTTP `auth` cookie.

### 3.4 VRChat client files (read/write)

| Typical Windows location | App behavior |
|--------------------------|--------------|
| `%USERPROFILE%\AppData\LocalLow\VRChat\VRChat\` | Default `config.json` and `output_log*.txt`; logs are **read and watched** (path configurable). Settings may delete `config.json` |
| `%LocalAppData%\vrchat-tweaker\ytdlp\yt-dlp.exe` | Cached official yt-dlp binary |
| `%LocalLow%\VRChat\VRChat\Tools\yt-dlp.exe` | Symlink to the cache when maintain mode is on |
| `%AppData%\yt-dlp\config` or `config.txt` | Local yt-dlp user config written by Cookie linkage (`--cookies-from-browser` or `--cookies`) |

Cookie linkage writes **local yt-dlp options**; the app does not upload cookie contents to an operator server (there is no such server). Cookies are read by **yt-dlp when VRChat runs yt-dlp**.

### 3.5 Other

- Single-instance mutex/event and local activation socket (`internal/infrastructure/singleinstance`) are used only for on-device process coordination.

## 4. External communication

### 4.1 VRChat

| Endpoint | Purpose |
|----------|---------|
| `https://api.vrchat.cloud/api/1` | REST API (login, current user, friends, users, avatars, status, favorites, etc.) |
| `wss://pipeline.vrchat.cloud/?authToken=…` | Pipeline WebSocket while logged in |
| Image URLs in API responses (e.g. under `api.vrchat.cloud`) | Loaded by the WebView for avatars and similar UI |

VRChat requests use User-Agent `VRChat Tweaker/<version>` with contact URL `https://github.com/JO3QMA/vrctweaker/issues`.

### 4.2 VRChat status

| Endpoint | Purpose |
|----------|---------|
| `https://status.vrchat.com/api/v2/` | Dashboard server status |

### 4.3 GitHub (yt-dlp updates)

| Endpoint | Purpose |
|----------|---------|
| `https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest` | Latest release metadata |
| `https://github.com/…`, `https://objects.githubusercontent.com/…`, `https://release-assets.githubusercontent.com/…` | Download `yt-dlp.exe` (hosts validated in code) |

Used when you run an update from the Video settings or when yt-dlp maintain mode is enabled.

### 4.4 Video sites (indirect)

The app parses video URLs from output_log into the database. HTTP to streaming hosts happens when **VRChat runs yt-dlp**, not as a always-on direct feature of Tweaker.

### 4.5 Links you open explicitly

Documentation links (this policy, terms, Getting Started) open in your default browser.

## 5. What is not sent to the distributor

Not uploaded automatically by the app:

- Full SQLite contents (encounters, friends cache, screenshot index)
- Full output_log text
- Screenshot file bytes
- Auth tokens or wrapped blobs (except session use with VRChat API/Pipeline)
- Dedicated analytics or crash-report endpoints (not implemented)

If **you** post to GitHub Issues, that content is stored on GitHub.

## 6. Deletion and retention

| Action | Effect |
|--------|--------|
| **Logout** | Clears stored credentials, session, and related self cache rows |
| **Activity retention days** | Deletes old encounters, play sessions, video playback history, friend avatar observations |
| **Clear encounters** | Deletes all encounter and friend-avatar observation rows |
| **Clear screenshot index** | Deletes screenshot-related DB rows; files remain |
| **Clear friends cache** | Deletes all `users_cache` rows and clears friends sync timestamp |
| **VACUUM** | SQLite maintenance only |
| **Clear VRChat asset cache** | Deletes files in VRChat’s asset cache directory when used |
| **Delete VRChat config.json** | Removes that VRChat config file when used |
| **Remove app data folders** | Manually delete `%AppData%\Roaming\vrchat-tweaker\`, `%LocalAppData%\vrchat-tweaker\`, etc. |

Clear WebView site data to remove IndexedDB wrapping keys.

## 7. Changes

Updates are published in `docs/legal/privacy-policy-en.md` in the repository.

## 8. Japanese version

[プライバシーポリシー（日本語）](./privacy-policy-ja.md)
