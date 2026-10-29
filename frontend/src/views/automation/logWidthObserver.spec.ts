import { afterEach, describe, expect, it, vi } from "vitest";
import { startLogWidthObserver } from "./logWidthObserver";

const originalObserver = globalThis.ResizeObserver;

describe("startLogWidthObserver", () => {
  afterEach(() => {
    globalThis.ResizeObserver = originalObserver;
    vi.unstubAllGlobals();
  });

  it("reports offsetWidth before observe starts", () => {
    const calls: number[] = [];
    let observed = false;
    class SilentObserver implements ResizeObserver {
      observe(): void {
        observed = calls.length > 0;
      }
      unobserve(): void {}
      disconnect(): void {}
    }
    vi.stubGlobal("ResizeObserver", SilentObserver);
    const el = { offsetWidth: 500 } as HTMLElement;

    const stop = startLogWidthObserver(el, (width) => {
      calls.push(width);
    });

    expect(calls).toEqual([500]);
    expect(observed).toBe(true);
    stop();
  });

  it("measures and skips observe when ResizeObserver is missing", () => {
    vi.stubGlobal("ResizeObserver", undefined);
    const calls: number[] = [];
    const el = { offsetWidth: 500 } as HTMLElement;

    const stop = startLogWidthObserver(el, (width) => {
      calls.push(width);
    });

    expect(calls).toEqual([500]);
    stop();
  });
});
