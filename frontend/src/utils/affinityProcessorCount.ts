import { App } from "../wails/app";
import { DEFAULT_LOGICAL_PROCESSOR_COUNT } from "./affinityMask";

export async function resolveLogicalProcessorCount(): Promise<number> {
  try {
    const n = await App.getLogicalProcessorCount();
    return n > 0 ? n : DEFAULT_LOGICAL_PROCESSOR_COUNT;
  } catch {
    return DEFAULT_LOGICAL_PROCESSOR_COUNT;
  }
}
