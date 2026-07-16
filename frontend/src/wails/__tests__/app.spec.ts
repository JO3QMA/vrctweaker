import {
  afterEach,
  beforeEach,
  describe,
  expect,
  it,
  vi,
  type Mock,
} from "vitest";
import {
  App,
  callApp,
  isWailsRuntime,
  PRIORITY_OMIT,
  type AppBindings,
} from "../app";

function setWindowGoApp(app: Partial<AppBindings>): void {
  window.go = { main: { App: app as AppBindings } };
}

function mockWithProbe(): {
  GetLanguage: Mock<AppBindings["GetLanguage"]>;
  RuntimeIsWindows: Mock<AppBindings["RuntimeIsWindows"]>;
} {
  return { GetLanguage: vi.fn(), RuntimeIsWindows: vi.fn() };
}

describe("app exports", () => {
  it("exports PRIORITY_OMIT as -999", () => {
    expect(PRIORITY_OMIT).toBe(-999);
  });
});

describe("isWailsRuntime", () => {
  let prevGo: typeof window.go;

  beforeEach(() => {
    prevGo = window.go;
    setWindowGoApp(mockWithProbe());
  });

  afterEach(() => {
    window.go = prevGo;
  });

  it("returns true when window.go.main.App exists", () => {
    expect(isWailsRuntime()).toBe(true);
  });

  it("returns false when bindings are missing", () => {
    window.go = undefined;
    expect(isWailsRuntime()).toBe(false);
  });
});

describe("callApp", () => {
  let mockBindings: ReturnType<typeof mockWithProbe>;
  let prevGo: typeof window.go;

  beforeEach(() => {
    prevGo = window.go;
    mockBindings = mockWithProbe();
    setWindowGoApp(mockBindings);
  });

  afterEach(() => {
    window.go = prevGo;
  });

  it("returns fallback when bindings are missing (Vitest skips dev wait)", async () => {
    window.go = undefined;
    const out = await callApp(async () => "invoked", "fallback");
    expect(out).toBe("fallback");
  });

  it("invokes fn with bindings and returns its result", async () => {
    mockBindings.GetLanguage.mockResolvedValue("ja");
    const out = await callApp((a) => a.GetLanguage(), "fallback");
    expect(mockBindings.GetLanguage).toHaveBeenCalled();
    expect(out).toBe("ja");
  });

  it("propagates rejection from the binding", async () => {
    mockBindings.GetLanguage.mockRejectedValue(new Error("backend failed"));
    await expect(callApp((a) => a.GetLanguage(), "fallback")).rejects.toThrow(
      "backend failed",
    );
  });
});

describe("App", () => {
  let mockBindings: ReturnType<typeof mockWithProbe>;
  let prevGo: typeof window.go;

  beforeEach(() => {
    prevGo = window.go;
    mockBindings = mockWithProbe();
    setWindowGoApp(mockBindings);
  });

  afterEach(() => {
    window.go = prevGo;
  });

  it("getLanguage delegates to GetLanguage binding via callApp", async () => {
    mockBindings.GetLanguage.mockResolvedValue("en");
    await expect(App.getLanguage()).resolves.toBe("en");
    expect(mockBindings.GetLanguage).toHaveBeenCalled();
  });

  it("runtimeIsWindows delegates to RuntimeIsWindows binding via callApp", async () => {
    mockBindings.RuntimeIsWindows.mockResolvedValue(true);
    await expect(App.runtimeIsWindows()).resolves.toBe(true);
    expect(mockBindings.RuntimeIsWindows).toHaveBeenCalled();
  });
});

describe("App fallbacks without bindings", () => {
  let prevGo: typeof window.go;

  beforeEach(() => {
    prevGo = window.go;
    window.go = undefined;
  });

  afterEach(() => {
    window.go = prevGo;
  });

  it("getLanguage returns empty string fallback", async () => {
    await expect(App.getLanguage()).resolves.toBe("");
  });

  it("runtimeIsWindows returns false fallback", async () => {
    await expect(App.runtimeIsWindows()).resolves.toBe(false);
  });

  it("parseLaunchArgsForGUI returns default DTO with PRIORITY_OMIT", async () => {
    const dto = await App.parseLaunchArgsForGUI("");
    expect(dto.processPriority).toBe(PRIORITY_OMIT);
    expect(dto.mainThreadPriority).toBe(PRIORITY_OMIT);
    expect(dto.fps).toBe(90);
  });
});

describe("callApp DEV binding race recovery", () => {
  let prevGo: typeof window.go;
  const injectedScripts: HTMLScriptElement[] = [];
  let origCreateElement: typeof document.createElement;

  function removeAllWailsScripts() {
    for (const el of document.querySelectorAll('head script[src*="wails/"]')) {
      el.remove();
    }
  }

  function injectWailsScript(src = "/wails/runtime.js") {
    const script = document.createElement("script");
    script.setAttribute("src", src);
    document.head.appendChild(script);
    injectedScripts.push(script);
    return script;
  }

  beforeEach(() => {
    prevGo = window.go;
    removeAllWailsScripts();
    injectedScripts.length = 0;
    vi.resetModules();
    vi.stubEnv("DEV", true);
    vi.stubEnv("MODE", "development");
    origCreateElement = document.createElement.bind(document);
    vi.spyOn(document, "createElement").mockImplementation((tag, options) => {
      const el = origCreateElement(tag, options);
      if (String(tag).toLowerCase() === "script") {
        queueMicrotask(() => {
          el.dispatchEvent(new Event("load"));
        });
      }
      return el;
    });
  });

  afterEach(() => {
    window.go = prevGo;
    for (const script of injectedScripts.splice(0)) {
      script.remove();
    }
    removeAllWailsScripts();
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
    vi.resetModules();
  });

  it("waits for bindings via requestAnimationFrame when scripts expect Wails", async () => {
    injectWailsScript();
    window.go = undefined;

    const mockBindings = mockWithProbe();
    mockBindings.GetLanguage.mockResolvedValue("from-wait");

    let frames = 0;
    vi.spyOn(window, "requestAnimationFrame").mockImplementation((cb) => {
      frames += 1;
      if (frames === 1) {
        setWindowGoApp(mockBindings);
      }
      cb(0);
      return frames;
    });

    const { callApp: freshCallApp } = await import("../app");
    await expect(
      freshCallApp((a) => a.GetLanguage(), "fallback"),
    ).resolves.toBe("from-wait");
  });

  it("returns fallback after exhausting rAF wait without bindings", async () => {
    injectWailsScript();
    window.go = undefined;

    vi.spyOn(window, "requestAnimationFrame").mockImplementation((cb) => {
      cb(0);
      return 1;
    });

    const { callApp: freshCallApp } = await import("../app");
    await expect(freshCallApp(async () => "never", "fallback")).resolves.toBe(
      "fallback",
    );
  });

  it("reloads wails scripts once then falls back when bindings never appear", async () => {
    injectWailsScript("/wails/ipc.js");
    window.go = undefined;

    vi.spyOn(window, "requestAnimationFrame").mockImplementation((cb) => {
      cb(0);
      return 1;
    });

    const { callApp: freshCallApp } = await import("../app");
    await expect(freshCallApp(async () => "never", "fallback")).resolves.toBe(
      "fallback",
    );
    expect(
      document.querySelector('script[src*="wails/ipc.js?wailsRetry="]'),
    ).not.toBeNull();
  });

  it("falls back when dev script reload fails", async () => {
    injectWailsScript("/wails/ipc.js");
    window.go = undefined;

    vi.spyOn(window, "requestAnimationFrame").mockImplementation((cb) => {
      cb(0);
      return 1;
    });

    vi.mocked(document.createElement).mockImplementation((tag, options) => {
      const el = origCreateElement(tag, options);
      if (String(tag).toLowerCase() === "script") {
        queueMicrotask(() => {
          el.dispatchEvent(new Event("error"));
        });
      }
      return el;
    });

    const { callApp: freshCallApp } = await import("../app");
    await expect(freshCallApp(async () => "never", "fallback")).resolves.toBe(
      "fallback",
    );
  });

  it("skips dev wait when head has no wails script tags", async () => {
    window.go = undefined;
    const raf = vi.spyOn(window, "requestAnimationFrame");

    const { callApp: freshCallApp } = await import("../app");
    await expect(freshCallApp(async () => "never", "fallback")).resolves.toBe(
      "fallback",
    );
    expect(raf).not.toHaveBeenCalled();
  });

  it("treats page as not expecting Wails when document is unavailable", async () => {
    window.go = undefined;
    const doc = globalThis.document;
    // @ts-expect-error test-only: exercise pageExpectsWailsBindings guard
    delete globalThis.document;

    try {
      const { callApp: freshCallApp } = await import("../app");
      await expect(freshCallApp(async () => "never", "fallback")).resolves.toBe(
        "fallback",
      );
    } finally {
      globalThis.document = doc;
    }
  });

  it("rejects reload when runtime script fails to load", async () => {
    injectWailsScript("/wails/ipc.js");
    window.go = undefined;

    vi.spyOn(window, "requestAnimationFrame").mockImplementation((cb) => {
      cb(0);
      return 1;
    });

    vi.mocked(document.createElement).mockImplementation((tag, options) => {
      const el = origCreateElement(tag, options);
      if (String(tag).toLowerCase() === "script") {
        queueMicrotask(() => {
          const src = el.getAttribute("src") ?? "";
          el.dispatchEvent(
            new Event(src.includes("runtime.js") ? "error" : "load"),
          );
        });
      }
      return el;
    });

    const { callApp: freshCallApp } = await import("../app");
    await expect(freshCallApp(async () => "never", "fallback")).resolves.toBe(
      "fallback",
    );
  });
});
