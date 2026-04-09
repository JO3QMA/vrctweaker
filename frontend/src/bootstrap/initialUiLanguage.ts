import { App } from "@/wails/app";

export const FALLBACK_UI_LANGUAGE_CODE = "ja";

/** Go IPC が返らない・遅い場合でも UI を出すための上限（Windows 本番でハング対策）。 */
export const GET_UI_LANGUAGE_TIMEOUT_MS = 3000;

/**
 * 起動時の UI 言語コードを取得する。IPC 失敗・タイムアウト・空文字のときは日本語にフォールバックする。
 */
export async function getInitialUILanguageCode(
  fetchFn: () => Promise<string> = () => App.getUILanguage(),
): Promise<string> {
  let timeoutId: ReturnType<typeof setTimeout> | undefined;
  const timeoutPromise = new Promise<string>((resolve) => {
    timeoutId = setTimeout(
      () => resolve(FALLBACK_UI_LANGUAGE_CODE),
      GET_UI_LANGUAGE_TIMEOUT_MS,
    );
  });
  try {
    const code = await Promise.race([
      fetchFn().then((c) => {
        if (timeoutId !== undefined) clearTimeout(timeoutId);
        return c;
      }),
      timeoutPromise,
    ]);
    const trimmed = typeof code === "string" ? code.trim() : "";
    return trimmed || FALLBACK_UI_LANGUAGE_CODE;
  } catch {
    if (timeoutId !== undefined) clearTimeout(timeoutId);
    return FALLBACK_UI_LANGUAGE_CODE;
  }
}
