import { describe, expect, it } from "vitest";
import { formatBackendError } from "./formatBackendError";

describe("formatBackendError", () => {
  it("prefers an Error message", () => {
    expect(formatBackendError(new Error("save failed"), "fallback")).toBe(
      "save failed",
    );
  });

  it("accepts a string error", () => {
    expect(formatBackendError("plain", "fallback")).toBe("plain");
  });

  it("reads a message field from an object", () => {
    expect(formatBackendError({ message: "from go" }, "fallback")).toBe(
      "from go",
    );
  });

  it("falls back when the error has no message", () => {
    expect(formatBackendError(null, "fallback")).toBe("fallback");
    expect(formatBackendError(new Error(""), "fallback")).toBe("fallback");
  });
});
