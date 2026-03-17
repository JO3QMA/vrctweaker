# docs/features 依存関係と実装順序

## 依存グラフ

```
Tier 0 (他に依存しない)
├── settings-paths-vrchat-and-outputlog.md
├── media-screenshot-metadata-extraction.md
├── identity-auth-os-keyring.md
├── automation-rule-eval-and-actions.md
└── settings-db-maintenance.md

Tier 1 (Tier 0 に依存)
├── activity-log-monitoring-output_log.md  ← settings-paths (output_log_path)
├── launcher-linux-proton-support.md        ← settings-paths (steam_path_linux)
├── activity-stats-aggregation.md           ← play_sessions (activity系)
└── ui-automation-view.md                   ← automation-rule-eval-and-actions

Tier 2 (Tier 1 に依存)
├── ui-gallery-view.md                      ← screenshots API (既存)
├── ui-activity-view.md                     ← activity-stats-aggregation
└── ui-friends-view.md                      ← identity-auth-os-keyring (RefreshFriends, IsLoggedIn)

Tier 3 (Tier 2 に依存)
├── media-world-join-from-screenshot.md    ← ui-gallery-view, launcher
└── identity-favorite-online-notifications.md ← ui-friends-view (SetFavorite, RefreshFriends)

Tier 4 (最後に実施)
└── testing-e2e-wails-mock-and-mcp-seeding.md ← 他機能のE2E検証のため
```

## 推奨実装順序（一括実行時）

1. settings-paths-vrchat-and-outputlog
2. media-screenshot-metadata-extraction
3. identity-auth-os-keyring
4. automation-rule-eval-and-actions
5. settings-db-maintenance
6. activity-log-monitoring-output_log
7. launcher-linux-proton-support
8. activity-stats-aggregation
9. ui-automation-view
10. ui-gallery-view
11. ui-activity-view
12. ui-friends-view
13. media-world-join-from-screenshot
14. identity-favorite-online-notifications
15. testing-e2e-wails-mock-and-mcp-seeding

##  explicit 依存一覧

| 機能 | 依存先 |
|------|--------|
| activity-log-monitoring-output_log | settings-paths-vrchat-and-outputlog |
| launcher-linux-proton-support | settings-paths-vrchat-and-outputlog |
| ui-automation-view | automation-rule-eval-and-actions |
| ui-activity-view | activity-stats-aggregation |
| ui-friends-view | identity-auth-os-keyring |
| media-world-join-from-screenshot | ui-gallery-view |
| identity-favorite-online-notifications | ui-friends-view |
| testing-e2e-wails-mock-and-mcp-seeding | 他UI機能（最後） |
