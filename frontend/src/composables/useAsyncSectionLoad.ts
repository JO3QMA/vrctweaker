import { onUnmounted, ref } from "vue";

export type AsyncSectionLoadMode = "queue" | "drop-overlap";

export type AsyncSectionLoadEnd<TOptions> = {
  replay: boolean;
  options: TOptions | undefined;
};

/**
 * useAsyncSectionLoad – Dashboard 系セクションの generation + inFlight +
 * pendingRefresh ロード制御（ADR 0009/0010/0014 と同型）を集約する。
 *
 * - `begin()` が null を返したら重複呼び出し（queue 済み or drop）。そのまま return する。
 * - `end(gen)` は finally で呼び、`loading` 終了と queue 済み再実行の有無を返す。
 * - `isStale(gen)` は await 後の世代チェック用。stale なら ref 更新をスキップする。
 * - `onUnmounted` で generation を increment する（呼び出し元での個別管理は不要）。
 * - エラー表示方針（inline / toast / fetchState）は呼び出し元が持つ。ここでは扱わない。
 */
export function useAsyncSectionLoad<TOptions = void>(options?: {
  mode?: AsyncSectionLoadMode;
  mergeOptions?: (
    queued: TOptions | undefined,
    incoming: TOptions | undefined,
  ) => TOptions;
}) {
  const mode = options?.mode ?? "queue";
  const mergeOptions = options?.mergeOptions;
  const loading = ref(true);
  const hasLoadedOnce = ref(false);

  let generation = 0;
  let inFlight = false;
  let pendingRefresh = false;
  let pendingOptions: TOptions | undefined;

  onUnmounted(() => {
    generation += 1;
  });

  function isStale(gen: number): boolean {
    return gen !== generation;
  }

  function begin(requestOptions?: TOptions): number | null {
    if (inFlight) {
      if (mode === "queue") {
        pendingRefresh = true;
        if (mergeOptions) {
          pendingOptions = mergeOptions(pendingOptions, requestOptions);
        } else {
          pendingOptions = requestOptions ?? pendingOptions;
        }
      }
      return null;
    }
    inFlight = true;
    pendingRefresh = false;
    return generation;
  }

  function end(gen: number): AsyncSectionLoadEnd<TOptions> {
    inFlight = false;
    if (gen !== generation) {
      pendingRefresh = false;
      pendingOptions = undefined;
      return { replay: false, options: undefined };
    }
    loading.value = false;
    if (!pendingRefresh) {
      return { replay: false, options: undefined };
    }
    const next = pendingOptions;
    pendingRefresh = false;
    pendingOptions = undefined;
    return { replay: true, options: next };
  }

  function markLoaded(): void {
    hasLoadedOnce.value = true;
  }

  return { loading, hasLoadedOnce, begin, end, isStale, markLoaded };
}
