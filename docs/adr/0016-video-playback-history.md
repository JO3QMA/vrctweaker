# ADR 0016: Video playback history は SessionCorrelator で消費する

## Status

Accepted（grill-with-docs / Issue #176）

## Context

- ADR 0002 は `VideoPlayback` をパースのみ・SessionCorrelator 未消費（スコープ外）としていた
- Issue #176 で **Video playback attempt** を永続化し、**Video playback history** を動画タブに出す。試行と結果の相関、および **Video playback world context**（試行時点の Open play session）が必要
- 用語の正本は [`CONTEXT.md`](../../CONTEXT.md)（Video playback attempt / outcome / correlation / history など）

## Decision

1. **SessionCorrelator を拡張**して Video playback を消費する（専用 correlator や adapter 内 pending マップは採らない）
2. 同一 Log source の session／world 状態と **Video playback correlation**（URL・FIFO）を同居させる。ファイル境界の `Reset`・Log replay は既存と同じライフサイクル
3. 永続化は fine-grained command → usecase（Encounter / Play session と同型）。UI 通知は Encounter 用イベントと分離（**Video playback history refresh**）
4. ADR 0002 の「VideoPlayback 未消費」条項は、本 ADR により **Video playback については置き換え**（AvatarSwitch の扱いは 0002 のまま未消費でよい）

## Consequences

- correlator と command 面が増えるが、world 文脈の二重管理を避けられる
- パーサーは試行行に加え ERROR / `resolved to` をイベント化し、outcome 規則（ERROR 優先）は domain で固定する
- 一覧 Wails API は `App.Encounters` と同型（`([]DTO, error)`。空は空スライス。infra 失敗は error）。取得失敗 UI はカード内表示（`ElMessage` なし）

## Test focus（v1）

- ERROR 優先（後続 `resolved to` で成功上書きしない）／成功 resolve／Open のみ
- 孤児結果の破棄／同 URL FIFO／world 文脈の有無／Reset 後 pending クリア
- Log replay で永続化（automation 非発火）／一覧取得失敗 UI／試行 URL コピーのみ
