import { describe, it, expect } from "vitest";
import {
  buildGallerySearchFilter,
  resolveWorldSearchInput,
} from "../gallerySearchFilter";

describe("resolveWorldSearchInput", () => {
  it("returns empty for blank input", () => {
    expect(resolveWorldSearchInput("  ")).toEqual({});
  });

  it("uses worldId for wrld_ prefix (case insensitive)", () => {
    expect(resolveWorldSearchInput("wrld_abc")).toEqual({
      worldId: "wrld_abc",
    });
    expect(resolveWorldSearchInput("WRLD_XYZ")).toEqual({
      worldId: "WRLD_XYZ",
    });
  });

  it("uses worldName partial match for other input", () => {
    expect(resolveWorldSearchInput("My World")).toEqual({
      worldName: "My World",
    });
  });
});

describe("buildGallerySearchFilter", () => {
  it("returns null when no filters are active", () => {
    expect(buildGallerySearchFilter("", null)).toBeNull();
    expect(buildGallerySearchFilter("  ", undefined)).toBeNull();
  });

  it("combines world and date filters", () => {
    expect(
      buildGallerySearchFilter("wrld_test", ["2024-01-01", "2024-01-31"]),
    ).toEqual({
      worldId: "wrld_test",
      dateFrom: "2024-01-01",
      dateTo: "2024-01-31",
    });
  });

  it("supports world name only", () => {
    expect(buildGallerySearchFilter("Gallery", null)).toEqual({
      worldName: "Gallery",
    });
  });

  it("supports date range only", () => {
    expect(buildGallerySearchFilter("", ["2025-06-01", "2025-06-02"])).toEqual({
      dateFrom: "2025-06-01",
      dateTo: "2025-06-02",
    });
  });
});
