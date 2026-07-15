import { describe, it, expect } from "vitest";
import { formatError } from "../formatError";

describe("formatError", () => {
  it("returns trimmed string errors", () => {
    expect(formatError("  oops  ", "fallback")).toBe("oops");
  });

  it("returns Error message", () => {
    expect(formatError(new Error("boom"), "fallback")).toBe("boom");
  });

  it("returns object message field", () => {
    expect(formatError({ message: "from object" }, "fallback")).toBe(
      "from object",
    );
  });

  it("returns fallback for empty values", () => {
    expect(formatError("", "fallback")).toBe("fallback");
    expect(formatError(null, "fallback")).toBe("fallback");
    expect(formatError({ message: "" }, "fallback")).toBe("fallback");
  });
});
