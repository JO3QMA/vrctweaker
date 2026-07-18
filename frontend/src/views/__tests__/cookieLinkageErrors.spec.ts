import { describe, it, expect } from "vitest";
import {
  classifyCookieLinkageError,
  cookieLinkageErrorI18nKey,
} from "../cookieLinkageErrors";

describe("classifyCookieLinkageError", () => {
  it("maps risk ack sentinel", () => {
    expect(classifyCookieLinkageError("errorRiskAckRequired")).toBe("riskAck");
    expect(
      classifyCookieLinkageError("cookie linkage risk acknowledgment required"),
    ).toBe("riskAck");
  });

  it("maps missing cookies file", () => {
    expect(classifyCookieLinkageError("errorCookiesFileMissing")).toBe(
      "cookiesFileMissing",
    );
  });

  it("maps config read failures", () => {
    expect(
      classifyCookieLinkageError(
        "cookie linkage config read: path is a directory",
      ),
    ).toBe("configRead");
  });

  it("handles empty", () => {
    expect(classifyCookieLinkageError(null)).toBe("generic");
    expect(cookieLinkageErrorI18nKey("x")).toBe(
      "settings.cookieLinkage.errors.generic",
    );
  });
});
