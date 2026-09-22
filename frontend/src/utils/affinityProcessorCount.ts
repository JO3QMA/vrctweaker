import { App, logicalProcessorCountFallback } from "../wails/app";

const MAX_LOGICAL_PROCESSORS = 512;

function cappedLogicalProcessorFallback(): number {
  return Math.min(logicalProcessorCountFallback(), MAX_LOGICAL_PROCESSORS);
}

export function normalizeLogicalProcessorCount(n: number): number {
  if (
    !Number.isFinite(n) ||
    !Number.isInteger(n) ||
    n <= 0 ||
    n > MAX_LOGICAL_PROCESSORS
  ) {
    return cappedLogicalProcessorFallback();
  }
  return n;
}

export async function resolveLogicalProcessorCount(): Promise<number> {
  try {
    const n = await App.getLogicalProcessorCount();
    return normalizeLogicalProcessorCount(n);
  } catch {
    return cappedLogicalProcessorFallback();
  }
}
