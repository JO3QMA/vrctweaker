import { describe, it, expect } from "vitest";
import { pruneThumbnailUrlMap } from "../galleryThumbnailCache";

describe("pruneThumbnailUrlMap", () => {
  it("removes ids not in listIds", () => {
    const listIds = new Set(["a"]);
    const retained = new Set(["a", "b"]);
    const out = pruneThumbnailUrlMap(
      { a: "u1", b: "u2", c: "u3" },
      listIds,
      retained,
    );
    expect(out).toEqual({ a: "u1" });
  });

  it("removes ids not in retainedIds", () => {
    const listIds = new Set(["a", "b"]);
    const retained = new Set(["a"]);
    const out = pruneThumbnailUrlMap({ a: "u1", b: "u2" }, listIds, retained);
    expect(out).toEqual({ a: "u1" });
  });

  it("keeps intersection of list and retained", () => {
    const listIds = new Set(["x", "y"]);
    const retained = new Set(["y", "z"]);
    const out = pruneThumbnailUrlMap(
      { x: "1", y: "2", z: "3" },
      listIds,
      retained,
    );
    expect(out).toEqual({ y: "2" });
  });

  it("returns empty when map is empty", () => {
    expect(pruneThumbnailUrlMap({}, new Set(["a"]), new Set(["a"]))).toEqual(
      {},
    );
  });
});
