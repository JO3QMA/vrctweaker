import { describe, it, expect } from "vitest";
import { formatRepositoryDisplayUrl } from "../formatRepositoryDisplayUrl";

describe("formatRepositoryDisplayUrl", () => {
  it("strips https scheme", () => {
    expect(
      formatRepositoryDisplayUrl(
        "https://github.com/element-plus/element-plus-icons",
      ),
    ).toBe("github.com/element-plus/element-plus-icons");
  });

  it("strips http scheme", () => {
    expect(formatRepositoryDisplayUrl("http://example.com/path")).toBe(
      "example.com/path",
    );
  });

  it("leaves scheme-less URLs unchanged", () => {
    expect(formatRepositoryDisplayUrl("github.com/foo/bar")).toBe(
      "github.com/foo/bar",
    );
  });
});
