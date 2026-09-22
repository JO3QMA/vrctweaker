import { App, logicalProcessorCountFallback } from "../wails/app";

const MAX_LOGICAL_PROCESSORS = 512;

export function normalizeLogicalProcessorCount(n: number): number {
  if (
    !Number.isFinite(n) ||
    !Number.isInteger(n) ||
    n <= 0 ||
    n > MAX_LOGICAL_PROCESSORS
  ) {
    return logicalProcessorCountFallback();
  }
  return n;
}

export async function resolveLogicalProcessorCount(): Promise<number> {
  try {
    const n = await App.getLogicalProcessorCount();
    return normalizeLogicalProcessorCount(n);
  } catch {
    return logicalProcessorCountFallback();
  }
}
