/**
 * VRChat の status 文字列（join me / ask me / busy / offline / active など）に対する
 * Element Plus el-tag の type。空や未知は info / primary で ElTag の空 type を避ける。
 */
export type VrcStatusElementTagType =
  | "primary"
  | "success"
  | "info"
  | "warning"
  | "danger";

export function vrcStatusElementTagType(
  status: string | undefined | null,
): VrcStatusElementTagType {
  const s = status?.trim().toLowerCase() ?? "";
  if (s === "") return "info";
  if (s === "offline") return "info";
  if (s === "join me") return "success";
  if (s === "busy") return "danger";
  if (s === "ask me") return "warning";
  if (s === "active") return "primary";
  return "primary";
}
