import { describe, expect, it } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent } from "vue";
import { useAsyncSectionLoad } from "./useAsyncSectionLoad";

type LoadOptions = {
  onlyIfNotDirty?: boolean;
  skipSnapshotUpdate?: boolean;
};

function mountLoader<TOptions = void>(
  options?: Parameters<typeof useAsyncSectionLoad<TOptions>>[0],
) {
  let loader!: ReturnType<typeof useAsyncSectionLoad<TOptions>>;
  const wrapper = mount(
    defineComponent({
      setup() {
        loader = useAsyncSectionLoad<TOptions>(options);
        return {};
      },
      template: "<div />",
    }),
  );
  return { loader: loader!, wrapper };
}

describe("useAsyncSectionLoad", () => {
  it("starts loading and not loaded", () => {
    const { loader, wrapper } = mountLoader();
    expect(loader.loading.value).toBe(true);
    expect(loader.hasLoadedOnce.value).toBe(false);
    wrapper.unmount();
  });

  it("queues overlapping load and replays once", async () => {
    const { loader, wrapper } = mountLoader();
    const gen = loader.begin();
    expect(gen).not.toBeNull();
    expect(loader.begin()).toBeNull();
    const next = loader.end(gen!);
    expect(next.replay).toBe(true);
    // Flight is free again after end.
    expect(loader.begin()).not.toBeNull();
    wrapper.unmount();
    await flushPromises();
  });

  it("merges queued options with OR", () => {
    const { loader, wrapper } = mountLoader<LoadOptions>({
      mergeOptions: (queued, incoming) => ({
        onlyIfNotDirty: Boolean(
          queued?.onlyIfNotDirty || incoming?.onlyIfNotDirty,
        ),
        skipSnapshotUpdate: Boolean(
          queued?.skipSnapshotUpdate || incoming?.skipSnapshotUpdate,
        ),
      }),
    });
    const gen = loader.begin()!;
    expect(loader.begin({ onlyIfNotDirty: true })).toBeNull();
    expect(loader.begin({ skipSnapshotUpdate: true })).toBeNull();
    const next = loader.end(gen);
    expect(next.replay).toBe(true);
    expect(next.options).toEqual({
      onlyIfNotDirty: true,
      skipSnapshotUpdate: true,
    });
    wrapper.unmount();
  });

  it("drops overlap in drop-overlap mode", () => {
    const { loader, wrapper } = mountLoader({
      mode: "drop-overlap",
    });
    const gen = loader.begin()!;
    expect(loader.begin()).toBeNull();
    const next = loader.end(gen);
    expect(next.replay).toBe(false);
    wrapper.unmount();
  });

  it("marks stale generations after unmount", () => {
    const { loader, wrapper } = mountLoader();
    const gen = loader.begin()!;
    expect(loader.isStale(gen)).toBe(false);
    wrapper.unmount();
    expect(loader.isStale(gen)).toBe(true);
  });

  it("does not finish loading or replay for stale generations", () => {
    const { loader, wrapper } = mountLoader();
    const gen = loader.begin()!;
    // Queue while in flight, then unmount before finishing.
    expect(loader.begin()).toBeNull();
    wrapper.unmount();
    const next = loader.end(gen);
    expect(next.replay).toBe(false);
    expect(loader.loading.value).toBe(true);
  });

  it("finishes loading and marks loaded on current generation", () => {
    const { loader, wrapper } = mountLoader();
    const gen = loader.begin()!;
    loader.markLoaded();
    expect(loader.hasLoadedOnce.value).toBe(true);
    loader.end(gen);
    expect(loader.loading.value).toBe(false);
    wrapper.unmount();
  });
});
