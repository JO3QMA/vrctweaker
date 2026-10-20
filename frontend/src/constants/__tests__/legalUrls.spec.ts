import { describe, expect, it } from "vitest";
import { legalDocsBaseUrl } from "../../config/legalDocsBase";
import { privacyPolicyUrl, termsOfUseUrl } from "../legalUrls";

describe("legalUrls", () => {
  it("maps ja to Japanese docs and others to English", () => {
    expect(privacyPolicyUrl("ja")).toContain("privacy-policy-ja.md");
    expect(privacyPolicyUrl("en")).toContain("privacy-policy-en.md");
    expect(privacyPolicyUrl("ko")).toContain("privacy-policy-en.md");
    expect(privacyPolicyUrl("")).toContain("privacy-policy-en.md");
  });

  it("builds stable GitHub doc URLs under configured base", () => {
    const base = legalDocsBaseUrl();
    expect(privacyPolicyUrl("ja")).toBe(`${base}/privacy-policy-ja.md`);
    expect(privacyPolicyUrl("en")).toBe(`${base}/privacy-policy-en.md`);
    expect(termsOfUseUrl("zh-CN")).toBe(`${base}/terms-en.md`);
  });
});
