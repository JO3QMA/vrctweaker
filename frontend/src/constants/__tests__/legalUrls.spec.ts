import { describe, expect, it } from "vitest";
import { legalDocLocale, privacyPolicyUrl, termsOfUseUrl } from "../legalUrls";

describe("legalUrls", () => {
  it("maps ja to Japanese docs and others to English", () => {
    expect(legalDocLocale("ja")).toBe("ja");
    expect(legalDocLocale("en")).toBe("en");
    expect(legalDocLocale("ko")).toBe("en");
    expect(legalDocLocale("")).toBe("en");
    expect(privacyPolicyUrl("")).toContain("privacy-policy-en.md");
  });

  it("builds stable GitHub doc URLs", () => {
    expect(privacyPolicyUrl("ja")).toContain("privacy-policy-ja.md");
    expect(privacyPolicyUrl("en")).toContain("privacy-policy-en.md");
    expect(termsOfUseUrl("zh-CN")).toContain("terms-en.md");
  });
});
